package celestia

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"net/http/httptest"

	"github.com/celestiaorg/celestia-app/pkg/appconsts"
	rpc "github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/celestiaorg/celestia-node/share"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/stretchr/testify/assert"
)

// / MockBlobAPI mocks the blob API
type MockBlobAPI struct {
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
	defer testServ.Close()

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

// teardown closes the client
func (m *mockDA) teardown() {
	m.client.Close()
}

// setup initializes the test instance and sets up common resources.
func setup(t *testing.T) *mockDA {
	mockService := NewMockService()
	defer mockService.Close()

	println("mock json-rpc server listening on: ", mockService.server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client, err := rpc.NewClient(ctx, mockService.server.URL, "test")
	assert.NoError(t, err)
	ns, err := hex.DecodeString("0000c9761e8b221ae42f")
	assert.NoError(t, err)
	namespace, err := share.NewBlobNamespaceV0(ns)
	assert.NoError(t, err)
	da := NewCelestiaDA(client, namespace, ctx)
	assert.Equal(t, da.client, client)

	return &mockDA{mockService, *da}
}

// TestCelestiaDA is the test suite function.
func TestCelestiaDA(t *testing.T) {
	m := setup(t)
	defer m.teardown()

	t.Run("MaxBlobSize", func(t *testing.T) {
		maxBlobSize, err := m.MaxBlobSize()
		assert.NoError(t, err)
		assert.Equal(t, uint64(appconsts.DefaultMaxBytes), maxBlobSize)
	})

	t.Run("Get_empty", func(t *testing.T) {
		blobs, err := m.Get(nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(blobs))
	})

	t.Run("GetIDs_empty", func(t *testing.T) {
		blobs, err := m.GetIDs(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(blobs))
	})

	t.Run("Commit_empty", func(t *testing.T) {
		commitments, err := m.Commit(nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(commitments))
	})

	t.Run("Submit_empty", func(t *testing.T) {
		blobs, proofs, err := m.Submit(nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(blobs))
		assert.Equal(t, 0, len(proofs))
	})

	t.Run("Validate_empty", func(t *testing.T) {
		valids, err := m.Validate(nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(valids))
	})
}
