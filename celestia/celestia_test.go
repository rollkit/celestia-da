package celestia

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/celestiaorg/celestia-app/pkg/appconsts"
	rpc "github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/celestiaorg/celestia-node/share"
	"github.com/celestiaorg/nmt"
	"github.com/stretchr/testify/assert"
)

// Blob is the data submitted/received from DA interface.
type Blob = []byte

// ID should contain serialized data required by the implementation to find blob in Data Availability layer.
type ID = []byte

// Commitment should contain serialized cryptographic commitment to Blob value.
type Commitment = []byte

// Proof should contain serialized proof of inclusion (publication) of Blob in Data Availability layer.
type Proof = []byte

// setup initializes the test instance and sets up common resources.
func setup(t *testing.T) *mockDA {
	mockService := NewMockService()

	t.Logf("mock json-rpc server listening on: %s", mockService.server.URL)

	ctx := context.TODO()
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

// teardown closes the client
func teardown(m *mockDA) {
	m.client.Close()
	m.s.Close()
}

// TestCelestiaDA is the test suite function.
func TestCelestiaDA(t *testing.T) {
	m := setup(t)
	defer teardown(m)

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
		blobs, proofs, err := m.Submit(nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(blobs))
		assert.Equal(t, 0, len(proofs))
	})

	t.Run("Validate_empty", func(t *testing.T) {
		valids, err := m.Validate(nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(valids))
	})

	t.Run("Get_existing", func(t *testing.T) {
		commitment, err := hex.DecodeString("1b454951cd722b2cf7be5b04554b76ccf48f65a7ad6af45055006994ce70fd9d")
		assert.NoError(t, err)
		blobs, err := m.Get([]ID{makeID(42, commitment)})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(blobs))
		blob1 := blobs[0]
		assert.Equal(t, "This is an example of some blob data", string(blob1))
	})

	t.Run("GetIDs_existing", func(t *testing.T) {
		ids, err := m.GetIDs(42)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ids))
		id1 := ids[0]
		commitment, err := hex.DecodeString("1b454951cd722b2cf7be5b04554b76ccf48f65a7ad6af45055006994ce70fd9d")
		assert.NoError(t, err)
		assert.Equal(t, makeID(42, commitment), id1)
	})

	t.Run("Commit_existing", func(t *testing.T) {
		commitments, err := m.Commit([]Blob{[]byte{0x00, 0x01, 0x02}})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(commitments))
	})

	t.Run("Submit_existing", func(t *testing.T) {
		blobs, proofs, err := m.Submit([]Blob{[]byte{0x00, 0x01, 0x02}}, nil)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(blobs))
		assert.Equal(t, 1, len(proofs))
	})

	t.Run("Validate_existing", func(t *testing.T) {
		commitment, err := hex.DecodeString("1b454951cd722b2cf7be5b04554b76ccf48f65a7ad6af45055006994ce70fd9d")
		assert.NoError(t, err)
		proof := nmt.NewInclusionProof(0, 4, [][]byte{[]byte("test")}, true)
		proofJSON, err := proof.MarshalJSON()
		assert.NoError(t, err)
		ids := []ID{makeID(42, commitment)}
		proofs := []Proof{proofJSON}
		valids, err := m.Validate(ids, proofs)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(valids))
	})
}
