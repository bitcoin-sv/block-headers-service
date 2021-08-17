package service

import (
	"context"

	"github.com/pkg/errors"

	headers "github.com/libsv/bitcoin-hc"
)

type syncService struct {
	blockRdr  headers.BlockReader
	blockWtr  headers.BlockheaderWriter
	heightRdr headers.HeightReader
}

func NewSyncService(blockRdr headers.BlockReader, heightRdr headers.HeightReader, blockWtr headers.BlockheaderWriter) *syncService {
	return &syncService{
		blockRdr:  blockRdr,
		heightRdr: heightRdr,
		blockWtr:  blockWtr,
	}
}

// Sync will check the current height we have cached and sync until we reach the current tip.
func (s *syncService) Sync(ctx context.Context) error {
	height, err := s.heightRdr.Height(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get cached block height when starting sync")
	}
	if height > 0 {
		height++
	}
	// find the current best block which will give us the height to work towards.
	bb, err := s.blockRdr.BestBlock(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get best block height when starting sync")
	}
	for i := height; i <= int(bb.Height); i++ {
		block, err := s.blockRdr.BlockByHeight(ctx, uint64(i))
		if err != nil {
			return errors.Wrap(err, "failed to get block when syncing")
		}
		if err := s.blockWtr.Create(ctx, *block); err != nil {
			return errors.Wrap(err, "failed to store block when syncing")
		}
	}
	return nil
}
