package celestia

import (
	"context"
	"time"

	openrpc "github.com/rollkit/celestia-openrpc"
	"github.com/rollkit/celestia-openrpc/types/blob"
	openrpcns "github.com/rollkit/celestia-openrpc/types/namespace"
	"github.com/rollkit/celestia-openrpc/types/share"
	"github.com/rollkit/go-da"
	"github.com/rollkit/rollkit/log"
)

// Config contains the node RPC configuration
type Config struct {
	AuthToken string        `json:"auth_token"`
	BaseURL   string        `json:"base_url"`
	Timeout   time.Duration `json:"timeout"`
	Fee       int64         `json:"fee"`
	GasLimit  uint64        `json:"gas_limit"`
}

// / CelestiaDA implements the celestia backend for the DA interface
type CelestiaDA struct {
	rpc       *openrpc.Client
	height    uint64
	namespace openrpcns.Namespace
	config    Config
	logger    log.Logger
	ctx       context.Context
}

func (c *CelestiaDA) Get(ids []da.ID) ([]da.Blob, error) {
	var blobs []da.Blob
	for _, id := range ids {
		// TODO: extract commitment from ID
		blob, err := c.rpc.Blob.Get(c.ctx, c.height, share.Namespace(c.namespace.Bytes()), blob.Commitment(id))
		if err != nil {
			return nil, err
		}
		blobs = append(blobs, blob.Data)
	}
	return blobs, nil
}

func (c *CelestiaDA) GetIDs(height uint64) ([]da.ID, error) {
	var ids []da.ID
	blobs, err := c.rpc.Blob.GetAll(c.ctx, c.height, []share.Namespace{c.namespace.Bytes()})
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
	var blobs []*blob.Blob
	for _, daBlob := range daBlobs {
		b, err := blob.NewBlobV0(c.namespace.Bytes(), daBlob)
		if err != nil {
			return nil, err
		}
		blobs = append(blobs, b)
	}
	commitments, err := blob.CreateCommitments(blobs)
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
		b, err := blob.NewBlobV0(c.namespace.Bytes(), daBlob)
		if err != nil {
			return nil, nil, err
		}
		blobs = append(blobs, b)
	}
	c.rpc.Blob.Submit(c.ctx, blobs, openrpc.DefaultSubmitOptions())
	return nil, nil, nil
}

func (c *CelestiaDA) Validate(ids []da.ID, daProofs []da.Proof) ([]bool, error) {
	var included []bool
	var proofs []*blob.Proof
	for _, daProof := range daProofs {
		proof := &blob.Proof{}
		if err := proof.UnmarshalJSON(daProof); err != nil {
			return nil, err
		}
		proofs = append(proofs, proof)
	}
	for i, id := range ids {
		// TODO: extract commitment from ID
		isIncluded, err := c.rpc.Blob.Included(c.ctx, c.height, share.Namespace(c.namespace.Bytes()), proofs[i], blob.Commitment(id))
		if err != nil {
			return nil, err
		}
		included = append(included, isIncluded)
	}
	return included, nil
}

var _ da.DA = &CelestiaDA{}
