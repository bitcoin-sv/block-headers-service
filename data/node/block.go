package node

import (
	"context"

	"github.com/ordishs/go-bitcoin"
	"github.com/pkg/errors"

	headers "github.com/libsv/bitcoin-hc"
)

type block struct {
	node *bitcoin.Bitcoind
}

func NewBlock(node *bitcoin.Bitcoind) *block {
	return &block{
		node: node,
	}
}

// BlockInfo will return extended info for a given block hash.
func (b *block) Header(ctx context.Context, args headers.HeaderArgs) (*headers.BlockHeader, error) {
	bh, err := b.node.GetBlockHeader(args.Blockhash)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find block with hash %s", args.Blockhash)
	}
	// TODO - handle not found error properly
	return &headers.BlockHeader{
		Hash:              bh.Hash,
		Versionhex:        bh.VersionHex,
		Merkleroot:        bh.MerkleRoot,
		Bits:              bh.Bits,
		Chainwork:         bh.Chainwork,
		Previousblockhash: bh.PreviousBlockHash,
		Nextblockhash:     bh.NextBlockHash,
		Confirmations:     uint64(bh.Confirmations),
		Height:            bh.Height,
		Mediantime:        bh.MedianTime,
		Difficulty:        bh.Difficulty,
		Version:           bh.Version,
		Time:              bh.Time,
		Nonce:             bh.Nonce,
	}, nil
}

// BestBlock will return the current block of the longest (best) chain.
func (b *block) BestBlock(ctx context.Context) (*headers.BlockHeader, error) {
	hash, err := b.node.GetBestBlockHash()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get best block hash")
	}
	return b.Header(ctx, headers.HeaderArgs{Blockhash: hash})
}

// BlockByHeight will return a block on the longest chain by index (height).
func (b *block) BlockByHeight(ctx context.Context, height uint64) (*headers.BlockHeader, error) {
	hash, err := b.node.GetBlockHash(int(height))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get best block hash")
	}

	block, err := b.Header(ctx, headers.HeaderArgs{Blockhash: hash})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return block, nil
}

// Height will return the current height of the longest (best) chain.
func (b *block) Height(ctx context.Context) (int, error) {
	hash, err := b.node.GetBestBlockHash()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get best block hash")
	}
	hdr, err := b.Header(ctx, headers.HeaderArgs{Blockhash: hash})
	if err != nil {
		return 0, err
	}
	return int(hdr.Height), nil
}
