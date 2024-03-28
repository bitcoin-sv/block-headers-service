package peer

import (
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
)

func minUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func findNextHeaderCheckpoint(checkpoints []chaincfg.Checkpoint, height int32) *chaincfg.Checkpoint {
	if len(checkpoints) == 0 {
		return nil
	}

	finalCheckpoint := &checkpoints[len(checkpoints)-1]
	if height >= finalCheckpoint.Height {
		return nil
	}

	nextCheckpoint := finalCheckpoint
	for i := len(checkpoints) - 2; i >= 0; i-- {
		if height >= checkpoints[i].Height {
			break
		}
		nextCheckpoint = &checkpoints[i]
	}
	return nextCheckpoint
}
