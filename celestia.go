package celestia

import (
	"context"
	"errors"
	"fmt"

	"github.com/celestiaorg/celestia-app/x/blob/types"
	rpc "github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/celestiaorg/celestia-node/blob"
	"github.com/celestiaorg/celestia-node/share"
	"github.com/celestiaorg/nmt"
	"github.com/rollkit/go-da"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// / CelestiaDA implements the celestia backend for the DA interface
type CelestiaDA struct {
	client    *rpc.Client
	height    uint64
	namespace share.Namespace
	ctx       context.Context
}

func NewCelestiaDA(client *rpc.Client, height uint64, namespace share.Namespace, ctx context.Context) *CelestiaDA {
	return &CelestiaDA{
		client:    client,
		height:    height,
		namespace: namespace,
		ctx:       ctx,
	}
}

func (c *CelestiaDA) Get(ids []da.ID) ([]da.Blob, error) {
	var blobs []da.Blob
	for _, id := range ids {
		// TODO: extract commitment from ID
		blob, err := c.client.Blob.Get(c.ctx, c.height, c.namespace, blob.Commitment(id))
		if err != nil {
			return nil, err
		}
		blobs = append(blobs, blob.Data)
	}
	return blobs, nil
}

func (c *CelestiaDA) GetIDs(height uint64) ([]da.ID, error) {
	var ids []da.ID
	blobs, err := c.client.Blob.GetAll(c.ctx, c.height, []share.Namespace{c.namespace})
	if errors.Is(err, blob.ErrBlobNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	for _, blob := range blobs {
		// TODO: commitment -> id
		ids = append(ids, blob.Commitment)
	}
	return ids, nil
}

func (c *CelestiaDA) Commit(daBlobs []da.Blob) ([]da.Commitment, error) {
	var blobs []*tmproto.Blob
	for _, daBlob := range daBlobs {
		b, err := blob.NewBlobV0(c.namespace, daBlob)
		if err != nil {
			return nil, err
		}
		blobs = append(blobs, &b.Blob)
	}
	commitments, err := types.CreateCommitments(blobs)
	if err != nil {
		return nil, err
	}
	var daCommitments []da.Commitment
	for _, commitment := range commitments {
		daCommitments = append(daCommitments, da.Commitment(commitment))
	}
	return daCommitments, nil
}

func (c *CelestiaDA) Submit(daBlobs []da.Blob) ([]da.ID, []da.Proof, error) {
	var blobs []*blob.Blob
	for _, daBlob := range daBlobs {
		b, err := blob.NewBlobV0(c.namespace, daBlob)
		if err != nil {
			return nil, nil, err
		}
		blobs = append(blobs, b)
	}
	height, err := c.client.Blob.Submit(c.ctx, blobs, blob.DefaultSubmitOptions())
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("succesfully submitted blobs", "height", height)
	return nil, nil, nil
}

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
		// TODO: extract commitment from ID
		isIncluded, err := c.client.Blob.Included(c.ctx, c.height, share.Namespace(c.namespace), proofs[i], blob.Commitment(id))
		if err != nil {
			return nil, err
		}
		included = append(included, isIncluded)
	}
	return included, nil
}

var _ da.DA = &CelestiaDA{}
