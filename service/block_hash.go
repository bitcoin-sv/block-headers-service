package service

import (
	"bytes"
	"github.com/bitcoin-sv/pulse/domains"
	"github.com/bitcoin-sv/pulse/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/pulse/internal/wire"
)

// MaxBlockHeaderPayload is the maximum number of bytes a block header can be.
// Version 4 bytes + Timestamp 4 bytes + Bits 4 bytes + Nonce 4 bytes +
// PrevBlock and MerkleRoot hashes.
const maxBlockHeaderPayload blockHashBufferMax = 16 + (chainhash.HashSize * 2)

type blockHashBufferMax int

// DefaultBlockHasher return default BlockHasher interface implementation.
func DefaultBlockHasher() BlockHasher {
	return maxBlockHeaderPayload
}

func (max blockHashBufferMax) BlockHash(h *domains.BlockHeaderSource) domains.BlockHash {
	buf := bytes.NewBuffer(make([]byte, 0, max))
	bh := wire.BlockHeader(*h)
	_ = wire.WriteBlockHeader(buf, &bh)

	return domains.BlockHash(chainhash.DoubleHashH(buf.Bytes()))
}
