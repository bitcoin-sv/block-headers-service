package testrepository

import (
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
	"math/big"
	"time"
)

var (
	genesisHash = &chaincfg.GenesisHash
	firstHash   = HashOf("00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048")
	secondHash  = HashOf("000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd")
	thirdHash   = HashOf("0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449")
	fourthHash  = HashOf("000000004ebadb55ee9096c9a2f8880e09da59c0d68b1c228da88e48844a1485")
)

func HashOf(s string) *chainhash.Hash {
	h, _ := chainhash.NewHashFromStr(s)
	return h
}

// StartingChain creates mocked chain entries containing only Genesis Block
func StartingChain() (db []domains.BlockHeader, tip *domains.BlockHeader) {
	db = startingChain()
	return db, &db[len(db)-1]
}

func startingChain() headerChain {
	genesisBlock := domains.CreateGenesisHeaderBlock()
	return []domains.BlockHeader{
		genesisBlock,
	}
}

// LongestChain creates mocked the longest chain entries (containing Genesis Block)
func LongestChain() (db []domains.BlockHeader, tip *domains.BlockHeader) {
	db = startingChain().
		add(*firstHash, *HashOf("0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098"), 2573394689).
		add(*secondHash, *HashOf("9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5"), 1639830024).
		add(*thirdHash, *HashOf("999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644"), 1844305925).
		add(*fourthHash, *HashOf("df2b060fa2e5e9c8ed5eaf6a45c13753ec8c63282b2688322eba40cd98ea067a"), 2850094635)

	return db, headerChain(db).last()
}

func OrphanChain() (db []domains.BlockHeader, tip *domains.BlockHeader) {
	orphan := newBlockHeader(
		1,
		*HashOf("881f062288a1b104603bf9d091011c4ae787ff70022e2e6a2fd27c46c604062"),
		*HashOf("0000000000000000000000000000000000000000000000000000000000000000"),
		*HashOf("63522845d294ee9b0188ae5cac91bf389a0c3723f084ca1025e7d9cdfe481ce1"),
		2011431709,
	)
	orphan.State = domains.Orphan

	return []domains.BlockHeader{orphan}, &orphan
}

func (c headerChain) add(hash chainhash.Hash, markleRoot chainhash.Hash, nonce uint32) headerChain {
	height := int32(len(c))
	return append(c, newBlockHeader(height, hash, c.last().Hash, markleRoot, nonce))
}

func (c headerChain) last() *domains.BlockHeader {
	return &c[len(c)-1]
}

func newBlockHeader(height int32, hash chainhash.Hash, prev chainhash.Hash, markleRoot chainhash.Hash, nonce uint32) domains.BlockHeader {
	bt, _ := time.Parse("yyyy-MM-dd hh:mm:ss", "2009-01-09 03:54:25")
	h := int64(height)
	return domains.BlockHeader{
		Height:           height,
		Hash:             hash,
		Version:          1,
		MerkleRoot:       markleRoot,
		Timestamp:        bt.Add(time.Duration(10*h) * time.Minute),
		Bits:             486604799,
		Nonce:            nonce,
		Chainwork:        4295032833,
		CumulatedWork:    big.NewInt(0).Mul(big.NewInt(4295032833), big.NewInt(h)),
		PreviousBlock:    prev,
		DifficultyTarget: 0,
	}
}

type headerChain []domains.BlockHeader
