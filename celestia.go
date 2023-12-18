package celestia

import (
	"context"
	"encoding/binary"
	"log"
	"strings"

	"github.com/celestiaorg/celestia-app/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/x/blob/types"
	rpc "github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/celestiaorg/celestia-node/blob"
	"github.com/celestiaorg/celestia-node/share"
	"github.com/celestiaorg/nmt"

	"github.com/rollkit/go-da"
)

// CelestiaDA implements the celestia backend for the DA interface
type CelestiaDA struct {
	client    *rpc.Client
	namespace share.Namespace
	ctx       context.Context
}

// NewCelestiaDA returns an instance of CelestiaDA
func NewCelestiaDA(client *rpc.Client, namespace share.Namespace, ctx context.Context) *CelestiaDA {
	return &CelestiaDA{
		client:    client,
		namespace: namespace,
		ctx:       ctx,
	}
}

// MaxBlobSize returns the max blob size
func (c *CelestiaDA) MaxBlobSize() (uint64, error) {
	return appconsts.DefaultMaxBytes, nil
}

// Get returns Blob for each given ID, or an error.
func (c *CelestiaDA) Get(ids []da.ID) ([]da.Blob, error) {
	var blobs []da.Blob
	for _, id := range ids {
		height, commitment := splitID(id)
		blob, err := c.client.Blob.Get(c.ctx, height, c.namespace, commitment)
		if err != nil {
			return nil, err
		}
		blobs = append(blobs, blob.Data)
	}
	return blobs, nil
}

// GetIDs returns IDs of all Blobs located in DA at given height.
func (c *CelestiaDA) GetIDs(height uint64) ([]da.ID, error) {
	var ids []da.ID
	blobs, err := c.client.Blob.GetAll(c.ctx, height, []share.Namespace{c.namespace})
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
func (c *CelestiaDA) Commit(daBlobs []da.Blob) ([]da.Commitment, error) {
	_, commitments, err := c.blobsAndCommitments(daBlobs)
	return commitments, err
}

// Submit submits the Blobs to Data Availability layer.
func (c *CelestiaDA) Submit(daBlobs []da.Blob, options *da.SubmitOptions) ([]da.ID, []da.Proof, error) {
	blobs, commitments, err := c.blobsAndCommitments(daBlobs)
	if err != nil {
		return nil, nil, err
	}
	height, err := c.client.Blob.Submit(c.ctx, blobs, &blob.SubmitOptions{Fee: options.Fee, GasLimit: options.Gas})
	if err != nil {
		return nil, nil, err
	}
	log.Println("successfully submitted blobs", "height", height)
	ids := make([]da.ID, len(daBlobs))
	proofs := make([]da.Proof, len(daBlobs))
	for i, commitment := range commitments {
		ids[i] = makeID(height, commitment)
		proof, err := c.client.Blob.GetProof(c.ctx, height, c.namespace, commitment)
		if err != nil {
			return nil, nil, err
		}
		// TODO(tzdybal): does always len(*proof) == 1?
		proofs[i], err = (*proof)[0].MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
	}
	return ids, proofs, nil
}

// blobsAndCommitments converts []da.Blob to []*blob.Blob and generates corresponding []da.Commitment
func (c *CelestiaDA) blobsAndCommitments(daBlobs []da.Blob) ([]*blob.Blob, []da.Commitment, error) {
	var blobs []*blob.Blob
	var commitments []da.Commitment
	for _, daBlob := range daBlobs {
		b, err := blob.NewBlobV0(c.namespace, daBlob)
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
func (c *CelestiaDA) Validate(ids []da.ID, daProofs []da.Proof) ([]bool, error) {
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
		isIncluded, _ := c.client.Blob.Included(c.ctx, height, c.namespace, proofs[i], commitment)
		included = append(included, isIncluded)
	}
	return included, nil
}

// heightLen is a length (in bytes) of serialized height.
//
// This is 8 as uint64 consist of 8 bytes.
const heightLen = 8

func makeID(height uint64, commitment da.Commitment) da.ID {
	id := make([]byte, heightLen+len(commitment))
	binary.LittleEndian.PutUint64(id, height)
	copy(id[heightLen:], commitment)
	return id
}

func splitID(id da.ID) (uint64, da.Commitment) {
	if len(id) <= heightLen {
		return 0, nil
	}
	return binary.LittleEndian.Uint64(id[:heightLen]), id[heightLen:]
}

var _ da.DA = &CelestiaDA{}
