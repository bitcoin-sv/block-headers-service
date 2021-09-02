package headers

import (
	"context"
)

// BlockReader can be used to read block information from a 3rd party service.
type BlockReader interface {
	BlockheaderReader
	// BestBlock wil return the block at the tip of the longest chain.
	BestBlock(ctx context.Context) (*BlockHeader, error)
	// BlockByHeight will return a block at the specified index on the longest chain.
	BlockByHeight(ctx context.Context, height uint64) (*BlockHeader, error)
}
