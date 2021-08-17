package bc

import (
	"context"
	"errors"
)

var (
	// ErrHeaderNotFound can be returned if the blockHash isn't found on the network.
	ErrHeaderNotFound = errors.New("header with not found")
	// ErrNotOnLongestChain indicates the blockhash is present but isn't on the longest current chain.
	ErrNotOnLongestChain = errors.New("header exists but is not on the longest chain")
)

// A BlockHeaderChain is a generic interface used to map things in the block header chain
// (chain of block headers). For example, it is used to get a block Header from a bitcoin
// block hash if it exists in the longest block header chain.
//
// Errors can be returned if the header isn't found or is on a stale chain, you may also use the
// ErrHeaderNotFound & ErrNotOnLongestChain sentinel errors when implementing the interface.
type BlockHeaderChain interface {
	BlockHeader(ctx context.Context, blockHash string) (*BlockHeader, error)
}