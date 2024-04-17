package peer

import (
	"fmt"
	"sync"

	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/rs/zerolog"
)

const notFound = -1

type checkpoint struct {
	currentCheckpoint *chaincfg.Checkpoint
	currentIndex      int
	finalCheckpoint   *chaincfg.Checkpoint
	checkpoints       []chaincfg.Checkpoint
	log               *zerolog.Logger
	lock              sync.RWMutex
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

// Height returns the height of the current checkpoint in a thread safety way.
func (ch *checkpoint) Height() int32 {
	ch.lock.RLock()
	defer ch.lock.RUnlock()

	return ch.currentCheckpoint.Height
}

// Hash returns the hash of the current checkpoint in a thread safety way.
func (ch *checkpoint) Hash() *chainhash.Hash {
	ch.lock.RLock()
	defer ch.lock.RUnlock()

	return ch.currentCheckpoint.Hash
}

// LastReached returns true if the last checkpoint has been reached.
func (ch *checkpoint) LastReached() bool {
	ch.lock.RLock()
	defer ch.lock.RUnlock()

	return ch.currentCheckpoint == nil
}

// VerifyAndAdvance checks if the header is valid according to the checkpoint and marks switches to next checkpoint if reached.
func (ch *checkpoint) VerifyAndAdvance(header *domains.BlockHeader) error {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	if ch.currentCheckpoint == nil {
		return nil
	}

	if header.Height < ch.currentCheckpoint.Height {
		return nil
	}

	if header.Height == ch.currentCheckpoint.Height {
		if header.Hash != *ch.currentCheckpoint.Hash {
			return fmt.Errorf("corresponding checkpoint height does not match, got: %v, exp: %v", header.Height, ch.currentCheckpoint.Height)
		}

		ch.next(header.Height)
		return nil
	}

	return fmt.Errorf("unexpected header above next checkpoint height, got: %v, for checkpoint at height %d", header, ch.currentCheckpoint.Height)
}

func (ch *checkpoint) next(height int32) {
	nextCheckpoint, index := ch.findNextCheckpoint(height)
	if nextCheckpoint == nil {
		ch.log.Info().Msgf("Last checkpoint reached at height %d", height)
	} else {
		ch.log.Info().
			Int32("height", nextCheckpoint.Height).
			Msgf("Setting next checkpoint at height %d", nextCheckpoint.Height)
	}
	ch.currentCheckpoint = nextCheckpoint
	ch.currentIndex = index
}

func (ch *checkpoint) findNextCheckpoint(height int32) (nextCheckpoint *chaincfg.Checkpoint, index int) {
	if len(ch.checkpoints) == 0 {
		return nil, notFound
	}

	if height >= ch.finalCheckpoint.Height {
		return nil, notFound
	}

	if ch.currentCheckpoint != nil {
		index = ch.currentIndex + 1
		return &ch.checkpoints[index], index
	}

	for i := 0; i < len(ch.checkpoints); i++ {
		if height < ch.checkpoints[i].Height {
			return &ch.checkpoints[i], i
		}
	}

	return nil, notFound
}
