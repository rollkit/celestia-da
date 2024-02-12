package celestia

import (
	"context"
	"encoding/hex"
	"net/http/httptest"

	"github.com/celestiaorg/celestia-node/blob"
	"github.com/celestiaorg/celestia-node/share"
	"github.com/celestiaorg/nmt"
	"github.com/filecoin-project/go-jsonrpc"
)

// MockBlobAPI mocks the blob API
type MockBlobAPI struct {
	height uint64
}

// Submit mocks the blob.Submit method
func (m *MockBlobAPI) Submit(ctx context.Context, blobs []*blob.Blob, gasPrice float64) (uint64, error) {
	m.height += 1
	return m.height, nil
}

// Get mocks the blob.Get method
func (m *MockBlobAPI) Get(ctx context.Context, height uint64, ns share.Namespace, _ blob.Commitment) (*blob.Blob, error) {
	data, err := hex.DecodeString("5468697320697320616e206578616d706c65206f6620736f6d6520626c6f622064617461")
	if err != nil {
		return nil, err
	}
	return blob.NewBlobV0(ns, data)
}

// GetAll mocks the blob.GetAll method
func (m *MockBlobAPI) GetAll(ctx context.Context, height uint64, ns []share.Namespace) ([]*blob.Blob, error) {
	if height == 0 {
		return []*blob.Blob{}, nil
	}
	data, err := hex.DecodeString("5468697320697320616e206578616d706c65206f6620736f6d6520626c6f622064617461")
	if err != nil {
		return nil, err
	}
	b, err := blob.NewBlobV0(ns[0], data)
	if err != nil {
		return nil, err
	}
	return []*blob.Blob{b}, nil
}

// GetProof mocks the blob.GetProof method
func (m *MockBlobAPI) GetProof(context.Context, uint64, share.Namespace, blob.Commitment) (*blob.Proof, error) {
	proof := nmt.NewInclusionProof(0, 4, [][]byte{[]byte("test")}, true)
	return &blob.Proof{&proof}, nil
}

// Included mocks the blob.Included method
func (m *MockBlobAPI) Included(context.Context, uint64, share.Namespace, *blob.Proof, blob.Commitment) (bool, error) {
	return true, nil
}

// MockService mocks the node RPC service
type MockService struct {
	blob   *MockBlobAPI
	server *httptest.Server
}

// Close closes the server
func (m *MockService) Close() {
	m.server.Close()
}

// NewMockService returns the mock service
func NewMockService() *MockService {
	rpcServer := jsonrpc.NewServer()

	blobAPI := &MockBlobAPI{}
	rpcServer.Register("blob", blobAPI)

	testServ := httptest.NewServer(rpcServer)

	mockService := &MockService{
		blob:   blobAPI,
		server: testServ,
	}

	return mockService
}

// mockDA returns the mock DA
type mockDA struct {
	s *MockService
	CelestiaDA
}
