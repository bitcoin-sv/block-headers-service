package service

import (
	"bytes"
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
	"github.com/libsv/bitcoin-hc/internal/wire"
)

// MaxBlockHeaderPayload is the maximum number of bytes a block header can be.
// Version 4 bytes + Timestamp 4 bytes + Bits 4 bytes + Nonce 4 bytes +
// PrevBlock and MerkleRoot hashes.
const maxBlockHeaderPayload blockHashBufferMax = 16 + (chainhash.HashSize * 2)

type blockHashBufferMax int

func DefaultBlockHasher() BlockHasher {
	return maxBlockHeaderPayload
}

func (max blockHashBufferMax) BlockHash(h *BlockHeaderSource) domains.BlockHash {
	buf := bytes.NewBuffer(make([]byte, 0, max))
	bh := wire.BlockHeader(*h)
	_ = wire.WriteBlockHeader(buf, 0, &bh)

	return domains.BlockHash(chainhash.DoubleHashH(buf.Bytes()))
}
