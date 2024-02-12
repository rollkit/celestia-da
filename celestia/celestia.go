package celestia

import (
	"context"
	"encoding/binary"
	"log"
	"strings"

	"github.com/celestiaorg/celestia-app/x/blob/types"
	rpc "github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/celestiaorg/celestia-node/blob"
	"github.com/celestiaorg/celestia-node/share"
	"github.com/celestiaorg/nmt"

	"github.com/rollkit/go-da"
)

const (
	// DefaultMaxBytes is the maximum blob size accepted by celestia core
	// ADR-13 claims worst case padding approaches 2 rows for a full data square:
	// see: https://github.com/celestiaorg/celestia-app/blob/main/docs/architecture/adr-013-non-interactive-default-rules-for-zero-padding.md
	// square size (64) * two rows = 128 shares
	// 128 shares * 512 bytes per share = 65,536 bytes to account for padding
	// 1,973,786 - 65,536 = 1,908,250 bytes
	DefaultMaxBytes = 1908250
)

// CelestiaDA implements the celestia backend for the DA interface
type CelestiaDA struct {
	client    *rpc.Client
	namespace share.Namespace
	gasPrice  float64
	ctx       context.Context
}

// NewCelestiaDA returns an instance of CelestiaDA
func NewCelestiaDA(client *rpc.Client, namespace share.Namespace, gasPrice float64, ctx context.Context) *CelestiaDA {
	return &CelestiaDA{
		client:    client,
		namespace: namespace,
		gasPrice:  gasPrice,
		ctx:       ctx,
	}
}

// MaxBlobSize returns the max blob size
func (c *CelestiaDA) MaxBlobSize(ctx context.Context) (uint64, error) {
	// TODO: pass-through query to node, app
	return DefaultMaxBytes, nil
}

// Get returns Blob for each given ID, or an error.
func (c *CelestiaDA) Get(ctx context.Context, ids []da.ID, ns da.Namespace) ([]da.Blob, error) {
	if ns == nil {
		ns = c.namespace
	}
	var blobs []da.Blob
	for _, id := range ids {
		height, commitment := splitID(id)
		blob, err := c.client.Blob.Get(ctx, height, ns, commitment)
		if err != nil {
			return nil, err
		}
		blobs = append(blobs, blob.Data)
	}
	return blobs, nil
}

// GetIDs returns IDs of all Blobs located in DA at given height.
func (c *CelestiaDA) GetIDs(ctx context.Context, height uint64, ns da.Namespace) ([]da.ID, error) {
	var ids []da.ID
	blobs, err := c.client.Blob.GetAll(ctx, height, []share.Namespace{ns})
	if err != nil {
		if strings.Contains(err.Error(), blob.ErrBlobNotFound.Error()) {
			return nil, nil
		}
		return nil, err
	}
	for _, b := range blobs {
		ids = append(ids, makeID(height, b.Commitment))
	}
	return ids, nil
}

// Commit creates a Commitment for each given Blob.
func (c *CelestiaDA) Commit(ctx context.Context, daBlobs []da.Blob, ns da.Namespace) ([]da.Commitment, error) {
	if ns == nil {
		ns = c.namespace
	}
	_, commitments, err := c.blobsAndCommitments(daBlobs, ns)
	return commitments, err
}

// Submit submits the Blobs to Data Availability layer.
func (c *CelestiaDA) Submit(ctx context.Context, daBlobs []da.Blob, gasPrice float64, ns da.Namespace) ([]da.ID, error) {
	if ns == nil {
		ns = c.namespace
	}
	blobs, _, err := c.blobsAndCommitments(daBlobs, ns)
	if err != nil {
		return nil, err
	}
	height, err := c.client.Blob.Submit(ctx, blobs, blob.GasPrice(gasPrice))
	if err != nil {
		return nil, err
	}
	log.Println("successfully submitted blobs", "height", height, "gasPrice", gasPrice)
	ids := make([]da.ID, len(blobs))
	for i, blob := range blobs {
		ids[i] = makeID(height, blob.Commitment)
	}
	return ids, nil
}

// GetProofs returns the inclusion proofs for the given IDs.
func (c *CelestiaDA) GetProofs(ctx context.Context, daIDs []da.ID, ns da.Namespace) ([]da.Proof, error) {
	if ns == nil {
		ns = c.namespace
	}
	proofs := make([]da.Proof, len(daIDs))
	for i, id := range daIDs {
		height, commitment := splitID(id)
		proof, err := c.client.Blob.GetProof(ctx, height, ns, commitment)
		if err != nil {
			return nil, err
		}
		proofs[i], err = proof.MarshalJSON()
		if err != nil {
			return nil, err
		}
	}
	return proofs, nil
}

// blobsAndCommitments converts []da.Blob to []*blob.Blob and generates corresponding []da.Commitment
func (c *CelestiaDA) blobsAndCommitments(daBlobs []da.Blob, ns da.Namespace) ([]*blob.Blob, []da.Commitment, error) {
	var blobs []*blob.Blob
	var commitments []da.Commitment
	for _, daBlob := range daBlobs {
		b, err := blob.NewBlobV0(ns, daBlob)
		if err != nil {
			return nil, nil, err
		}
		blobs = append(blobs, b)

		commitment, err := types.CreateCommitment(&b.Blob)
		if err != nil {
			return nil, nil, err
		}
		commitments = append(commitments, commitment)
	}
	return blobs, commitments, nil
}

// Validate validates Commitments against the corresponding Proofs. This should be possible without retrieving the Blobs.
func (c *CelestiaDA) Validate(ctx context.Context, ids []da.ID, daProofs []da.Proof, ns da.Namespace) ([]bool, error) {
	if ns == nil {
		ns = c.namespace
	}
	var included []bool
	var proofs []*blob.Proof
	for _, daProof := range daProofs {
		nmtProof := &nmt.Proof{}
		if err := nmtProof.UnmarshalJSON(daProof); err != nil {
			return nil, err
		}
		proof := &blob.Proof{nmtProof}
		proofs = append(proofs, proof)
	}
	for i, id := range ids {
		height, commitment := splitID(id)
		// TODO(tzdybal): for some reason, if proof doesn't match commitment, API returns (false, "blob: invalid proof")
		//    but analysis of the code in celestia-node implies this should never happen - maybe it's caused by openrpc?
		//    there is no way of gently handling errors here, but returned value is fine for us
		isIncluded, _ := c.client.Blob.Included(ctx, height, ns, proofs[i], commitment)
		included = append(included, isIncluded)
	}
	return included, nil
}

// heightLen is a length (in bytes) of serialized height.
//
// This is 8 as uint64 consist of 8 bytes.
const heightLen = 8

// MakeID returns the ID from the height and commitment.
func makeID(height uint64, commitment da.Commitment) da.ID {
	id := make([]byte, heightLen+len(commitment))
	binary.LittleEndian.PutUint64(id, height)
	copy(id[heightLen:], commitment)
	return id
}

// SplitID returns the height and commitment from the ID.
func splitID(id da.ID) (uint64, da.Commitment) {
	if len(id) <= heightLen {
		return 0, nil
	}
	commitment := id[heightLen:]
	return binary.LittleEndian.Uint64(id[:heightLen]), commitment
}

var _ da.DA = &CelestiaDA{}
