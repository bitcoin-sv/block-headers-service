package headers

import (
	"context"
)

type BlockArgs struct {
	BlockHash string
}

// BlockReader can be used to read block information from a 3rd party service.
type BlockReader interface {
	// BlockInfo will return information on a block by hash.
	BlockInfo(ctx context.Context, args BlockArgs) (*BlockHeader, error)
	// BestBlock wil return the block at the tip of the longest chain.
	BestBlock(ctx context.Context) (*BlockHeader, error)
	// BlockByHeight will return a block at the specified index on the longest chain.
	BlockByHeight(ctx context.Context, height uint64) (*BlockHeader, error)
}
