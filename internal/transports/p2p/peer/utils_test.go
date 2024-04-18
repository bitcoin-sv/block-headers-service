package peer

import (
	"testing"

	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
)

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
