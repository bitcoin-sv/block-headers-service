package fixtures

import (
	"math/big"
	"time"

	"github.com/bitcoin-sv/pulse/domains"
	"github.com/bitcoin-sv/pulse/internal/chaincfg/chainhash"
)

const (
	// DefaultChainWork is a default chain work result that is set in block headers fixtures.
	DefaultChainWork = 4295032833
)

// HashOf returns chainhash.Hash representation of string, ignoring errors.
func HashOf(s string) *chainhash.Hash {
	h, _ := chainhash.NewHashFromStr(s)
	return h
}

// BlockTimestampOf returns time representation of string in patter yyyy-MM-dd hh:mm:ss.
func BlockTimestampOf(s string) *time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		panic(err)
	}
	return &t
}

// BlockHeaderSourceOf creates BlockHeaderSource from BlockHeader for testing purposes.
func BlockHeaderSourceOf(header *domains.BlockHeader) *domains.BlockHeaderSource {
	return &domains.BlockHeaderSource{
		Version:    header.Version,
		PrevBlock:  header.PreviousBlock,
		MerkleRoot: header.MerkleRoot,
		Timestamp:  header.Timestamp,
		Bits:       header.Bits,
		Nonce:      header.Nonce,
	}
}

// BlockHeaderOf creates BlockHeader for testing purposes.
func BlockHeaderOf(height int32, hash *chainhash.Hash, hs *domains.BlockHeaderSource, s domains.HeaderState) *domains.BlockHeader {
	return &domains.BlockHeader{
		Height:        height,
		Hash:          *hash,
		Version:       hs.Version,
		PreviousBlock: hs.PrevBlock,
		MerkleRoot:    hs.MerkleRoot,
		Timestamp:     hs.Timestamp,
		Bits:          hs.Bits,
		Nonce:         hs.Nonce,
		Chainwork:     big.NewInt(DefaultChainWork),
		CumulatedWork: big.NewInt(0).Mul(big.NewInt(DefaultChainWork), big.NewInt(int64(height))),
		State:         s,
	}
}
