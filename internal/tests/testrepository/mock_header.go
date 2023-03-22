package testrepository

import (
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
	"math/big"
	"time"
)

const (
	DefaultChainWork = 4295032833
	DefaultBits      = 0x1d00ffff
)

var (
	FirstHash  = HashOf("00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048")
	SecondHash = HashOf("000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd")
	ThirdHash  = HashOf("0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449")
	FourthHash = HashOf("000000004ebadb55ee9096c9a2f8880e09da59c0d68b1c228da88e48844a1485")

	FirstStaleHash  = HashOf("000000000000839a77380d4690994eb38b7a8b67e4295121079ee6e98c7a8c5")
	SecondStaleHash = HashOf("0000000000006a6277380d4690994eb38b7a8b67e4295121079ee6e98c7a8c5")
	ThirdStaleHash  = HashOf("00000000000082b577380d4690994eb38b7a8b67e4295121079ee6e98c7a8c5")
	ForthStaleHash  = HashOf("0000000000004eba77380d4690994eb38b7a8b67e4295121079ee6e98c7a8c5")
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

func startingChain() HeaderChainFixture {
	genesisBlock := domains.CreateGenesisHeaderBlock()
	return []domains.BlockHeader{
		genesisBlock,
	}
}

// LongestChain creates mocked the longest chain entries (containing Genesis Block)
func LongestChain() (db HeaderChainFixture, tip *domains.BlockHeader) {
	db = startingChain().
		Add(*FirstHash, *HashOf("0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098"), 2573394689, domains.LongestChain).
		Add(*SecondHash, *HashOf("9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5"), 1639830024, domains.LongestChain).
		Add(*ThirdHash, *HashOf("999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644"), 1844305925, domains.LongestChain).
		Add(*FourthHash, *HashOf("df2b060fa2e5e9c8ed5eaf6a45c13753ec8c63282b2688322eba40cd98ea067a"), 2850094635, domains.LongestChain)

	return db, db.Tip()
}

func StaleChain() (db HeaderChainFixture, tip *domains.BlockHeader) {
	db = startingChain().
		Add(*FirstStaleHash, *HashOf("0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098"), 2573394689, domains.Stale).
		Add(*SecondStaleHash, *HashOf("9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5"), 1639830024, domains.Stale).
		Add(*ThirdStaleHash, *HashOf("00e75a406bdfe21d58772604499a9be6056de216be1a457a8351b077824d8568"), 1639830024, domains.Stale).
		Add(*ForthStaleHash, *HashOf("391eb3be1b96b8d98c61c0581cff54ef079168098987b99449aad22654baeda4"), 1639830024, domains.Stale)

	return db, db.Tip()
}

func OrphanChain() (db HeaderChainFixture, tip *domains.BlockHeader) {
	orphan := newBlockHeader(1, *HashOf("881f062288a1b104603bf9d091011c4ae787ff70022e2e6a2fd27c46c604062"), *HashOf("0000000000000000000000000000000000000000000000000000000000000000"), *HashOf("63522845d294ee9b0188ae5cac91bf389a0c3723f084ca1025e7d9cdfe481ce1"), 2011431709, "")
	orphan.State = domains.Orphan

	db = []domains.BlockHeader{orphan}
	return db, &orphan
}

func (c HeaderChainFixture) Add(hash chainhash.Hash, markleRoot chainhash.Hash, nonce uint32, s domains.HeaderState) HeaderChainFixture {
	height := int32(len(c))
	return append(c, newBlockHeader(height, hash, c.Tip().Hash, markleRoot, nonce, s))
}

func (c HeaderChainFixture) Db() []domains.BlockHeader {
	return c
}

func (c HeaderChainFixture) Tip() *domains.BlockHeader {
	return &c[len(c)-1]
}

func newBlockHeader(height int32, hash chainhash.Hash, prev chainhash.Hash, markleRoot chainhash.Hash, nonce uint32, s domains.HeaderState) domains.BlockHeader {
	bt, _ := time.Parse("yyyy-MM-dd hh:mm:ss", "2009-01-09 03:54:25")
	h := int64(height)
	return domains.BlockHeader{
		Height:           height,
		Hash:             hash,
		Version:          1,
		State:            s,
		MerkleRoot:       markleRoot,
		Timestamp:        bt.Add(time.Duration(10*h) * time.Minute),
		Bits:             DefaultBits,
		Nonce:            nonce,
		Chainwork:        DefaultChainWork,
		CumulatedWork:    big.NewInt(0).Mul(big.NewInt(DefaultChainWork), big.NewInt(h)),
		PreviousBlock:    prev,
		DifficultyTarget: 0,
	}
}

type HeaderChainFixture []domains.BlockHeader
