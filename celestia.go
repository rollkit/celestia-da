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
		// TODO: id -> commitment
		blob, err := c.rpc.Blob.Get(c.ctx, c.height, share.Namespace(c.namespace.Bytes()), blob.Commitment(id))
		if err != nil {
			return nil, err
		}
		blobs = append(blobs, blob.Data)
	}
	return blobs, nil
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

func (c *CelestiaDA) Validate(ids []da.ID, proofs []da.Proof) ([]bool, error) {
	return nil, nil
}

var _ da.DA = &CelestiaDA{}
