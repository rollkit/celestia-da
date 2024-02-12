package celestia

import (
	"context"
	"encoding/hex"
	"testing"

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

	nsHex, err := hex.DecodeString("0000c9761e8b221ae42f")
	assert.NoError(t, err)
	ns, err := share.NewBlobNamespaceV0(nsHex)
	assert.NoError(t, err)

	ctx := context.TODO()
	client, err := rpc.NewClient(ctx, mockService.server.URL, "test")
	assert.NoError(t, err)
	da := NewCelestiaDA(client, ns, -1, ctx)
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
	ctx := context.TODO()
	m := setup(t)
	defer teardown(m)

	nsHex, err := hex.DecodeString("0000c9761e8b221ae42f")
	assert.NoError(t, err)
	ns, err := share.NewBlobNamespaceV0(nsHex)
	assert.NoError(t, err)

	assert.NoError(t, err)
	t.Run("MaxBlobSize", func(t *testing.T) {
		maxBlobSize, err := m.MaxBlobSize(ctx)
		assert.NoError(t, err)
		assert.Equal(t, uint64(DefaultMaxBytes), maxBlobSize)
	})

	t.Run("Get_empty", func(t *testing.T) {
		blobs, err := m.Get(ctx, nil, ns)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(blobs))
	})

	t.Run("GetIDs_empty", func(t *testing.T) {
		blobs, err := m.GetIDs(ctx, 0, ns)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(blobs))
	})

	t.Run("Commit_empty", func(t *testing.T) {
		commitments, err := m.Commit(ctx, nil, ns)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(commitments))
	})

	t.Run("Submit_empty", func(t *testing.T) {
		blobs, err := m.Submit(ctx, nil, -1, ns)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(blobs))
	})

	t.Run("Validate_empty", func(t *testing.T) {
		valids, err := m.Validate(ctx, nil, nil, ns)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(valids))
	})

	t.Run("Get_existing", func(t *testing.T) {
		commitment, err := hex.DecodeString("1b454951cd722b2cf7be5b04554b76ccf48f65a7ad6af45055006994ce70fd9d")
		assert.NoError(t, err)
		blobs, err := m.Get(ctx, []ID{makeID(42, commitment)}, ns)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(blobs))
		blob1 := blobs[0]
		assert.Equal(t, "This is an example of some blob data", string(blob1))
	})

	t.Run("GetIDs_existing", func(t *testing.T) {
		ids, err := m.GetIDs(ctx, 42, ns)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ids))
		id1 := ids[0]
		commitment, err := hex.DecodeString("1b454951cd722b2cf7be5b04554b76ccf48f65a7ad6af45055006994ce70fd9d")
		assert.NoError(t, err)
		assert.Equal(t, makeID(42, commitment), id1)
	})

	t.Run("Commit_existing", func(t *testing.T) {
		commitments, err := m.Commit(ctx, []Blob{[]byte{0x00, 0x01, 0x02}}, ns)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(commitments))
	})

	t.Run("Submit_existing", func(t *testing.T) {
		blobs, err := m.Submit(ctx, []Blob{[]byte{0x00, 0x01, 0x02}}, -1, ns)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(blobs))
	})

	t.Run("Submit_existing_with_gasprice_global", func(t *testing.T) {
		m.CelestiaDA.gasPrice = 0.01
		blobs, err := m.Submit(ctx, []Blob{[]byte{0x00, 0x01, 0x02}}, -1, ns)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(blobs))
	})

	t.Run("Submit_existing_with_gasprice_override", func(t *testing.T) {
		blobs, err := m.Submit(ctx, []Blob{[]byte{0x00, 0x01, 0x02}}, 0.5, ns)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(blobs))
	})

	t.Run("Validate_existing", func(t *testing.T) {
		commitment, err := hex.DecodeString("1b454951cd722b2cf7be5b04554b76ccf48f65a7ad6af45055006994ce70fd9d")
		assert.NoError(t, err)
		proof := nmt.NewInclusionProof(0, 4, [][]byte{[]byte("test")}, true)
		proofJSON, err := proof.MarshalJSON()
		assert.NoError(t, err)
		ids := []ID{makeID(42, commitment)}
		proofs := []Proof{proofJSON}
		valids, err := m.Validate(ctx, ids, proofs, ns)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(valids))
	})
}
