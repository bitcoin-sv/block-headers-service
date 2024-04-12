package peer

import (
	"fmt"
	"sync"

	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/rs/zerolog"
)

type checkpoint struct {
	checkpoint      *chaincfg.Checkpoint
	currentIndex    int
	finalCheckpoint *chaincfg.Checkpoint
	checkpoints     []chaincfg.Checkpoint
	log             *zerolog.Logger
	lock            sync.RWMutex
}

func newCheckpoint(checkpoints []chaincfg.Checkpoint, tipHeight int32, log *zerolog.Logger) *checkpoint {
	logger := log.With().Str("subservice", "checkpoint").Logger()
	ch := &checkpoint{
		checkpoints: checkpoints,
		log:         &logger,
	}

	if len(checkpoints) != 0 {
		ch.finalCheckpoint = &checkpoints[len(checkpoints)-1]
		ch.next(tipHeight)
	}

	return ch
}

func (ch *checkpoint) Height() int32 {
	ch.lock.RLock()
	defer ch.lock.RUnlock()

	return ch.checkpoint.Height
}

func (ch *checkpoint) Hash() *chainhash.Hash {
	ch.lock.RLock()
	defer ch.lock.RUnlock()

	return ch.checkpoint.Hash
}

// LastReached returns true if the last checkpoint has been reached.
func (ch *checkpoint) LastReached() bool {
	ch.lock.RLock()
	defer ch.lock.RUnlock()

	return ch.checkpoint == nil
}

// Check checks if the header is valid according to the checkpoint and marks switches to next checkpoint if reached.
func (ch *checkpoint) Check(header *domains.BlockHeader) error {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	if ch.checkpoint == nil {
		return nil
	}

	if header.Height < ch.checkpoint.Height {
		return nil
	}

	if header.Height == ch.checkpoint.Height {
		if header.Hash != *ch.checkpoint.Hash {
			return fmt.Errorf("corresponding checkpoint height does not match, got: %v, exp: %v", header.Height, ch.checkpoint.Height)
		}

		ch.next(header.Height)
		return nil
	}

	if header.Height > ch.checkpoint.Height {
		return fmt.Errorf("unexpected header above next checkpoint height, got: %v, for checkpoint at height %d", header, ch.checkpoint.Height)
	}

	return nil
}

// UseTip sets the next checkpoint to the valid checkpoint for given tip height.
func (ch *checkpoint) UseTip(height int32) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	ch.checkpoint = nil
	ch.currentIndex = -1
	ch.next(height)
}

func (ch *checkpoint) next(height int32) {
	if len(ch.checkpoints) == 0 {
		return
	}

	nextCheckpoint, index := ch.findNextCheckpoint(height)
	if nextCheckpoint == nil {
		ch.log.Info().Msgf("Last checkpoint reached at height %d", height)
	} else {
		ch.log.Info().
			Int32("height", nextCheckpoint.Height).
			Msgf("Setting next checkpoint at height %d", nextCheckpoint.Height)
	}
	ch.checkpoint = nextCheckpoint
	ch.currentIndex = index
}

func (ch *checkpoint) findNextCheckpoint(height int32) (nextCheckpoint *chaincfg.Checkpoint, index int) {
	if len(ch.checkpoints) == 0 {
		return nil, -1
	}

	if height >= ch.finalCheckpoint.Height {
		return nil, -1
	}

	if ch.checkpoint != nil {
		index = ch.currentIndex + 1
		return &ch.checkpoints[index], index
	}

	nextCheckpoint = ch.finalCheckpoint
	index = len(ch.checkpoints) - 1
	for i := index - 1; i >= 0; i-- {
		if height >= ch.checkpoints[i].Height {
			break
		}
		nextCheckpoint = &ch.checkpoints[i]
		index = i
	}
	return nextCheckpoint, index
}
