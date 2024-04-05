package peer

import (
	"testing"

	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
)

func TestFindNextHeaderCheckpoint(t *testing.T) {
	// given
	checkpoints := chaincfg.MainNetParams.Checkpoints

	t.Run("height 0", func(t *testing.T) {
		// given
		height := int32(0)
		expectedCheckpoint := checkpoints[0]

		// when
		checkpoint := findNextHeaderCheckpoint(checkpoints, height)

		// then
		assert.Equal(t, *checkpoint, expectedCheckpoint)
	})

	t.Run("height 810000", func(t *testing.T) {
		// given
		height := int32(810000)
		expectedCheckpoint := checkpoints[len(checkpoints)-1]

		// when
		checkpoint := findNextHeaderCheckpoint(checkpoints, height)

		// then
		assert.Equal(t, *checkpoint, expectedCheckpoint)
	})

	t.Run("height 999999999 - nil checkpoint", func(t *testing.T) {
		// given
		height := int32(999999999)

		// when
		checkpoint := findNextHeaderCheckpoint(checkpoints, height)

		// then
		assert.Equal(t, checkpoint, nil)
	})
}

func TestSearchForFinalBlock(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// given
		invVects := []*wire.InvVect{
			{
				Type: wire.InvTypeError,
				Hash: [32]byte{},
			},
			{
				Type: wire.InvTypeBlock,
				Hash: [32]byte{},
			},
			{
				Type: wire.InvTypeBlock,
				Hash: [32]byte{},
			},
			{
				Type: wire.InvTypeBlock,
				Hash: [32]byte{},
			},
			{
				Type: wire.InvTypeTx,
				Hash: [32]byte{},
			},
			{
				Type: wire.InvTypeError,
				Hash: [32]byte{},
			},
		}
		expectedFinalBlock := 3

		// when
		finalBlock := searchForFinalBlockIndex(invVects)

		// then
		assert.Equal(t, finalBlock, expectedFinalBlock)
	})

	t.Run("not found", func(t *testing.T) {
		// given
		invVects := []*wire.InvVect{
			{
				Type: wire.InvTypeError,
				Hash: [32]byte{},
			},
			{
				Type: wire.InvTypeFilteredBlock,
				Hash: [32]byte{},
			},
			{
				Type: wire.InvTypeTx,
				Hash: [32]byte{},
			},
			{
				Type: wire.InvTypeError,
				Hash: [32]byte{},
			},
		}
		expectedFinalBlock := -1

		// when
		finalBlock := searchForFinalBlockIndex(invVects)

		// then
		assert.Equal(t, finalBlock, expectedFinalBlock)
	})
}
