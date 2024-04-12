package peer

import (
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
)

func searchForFinalBlockIndex(invVects []*wire.InvVect) int {
	lastBlock := -1
	for i := len(invVects) - 1; i >= 0; i-- {
		if invVects[i].Type == wire.InvTypeBlock {
			lastBlock = i
			break
		}
	}
	return lastBlock
}
