package fixtures

import (
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
)

// StartingChain creates mocked chain entries containing only Genesis Block.
func StartingChain() (db []domains.BlockHeader, tip *domains.BlockHeader) {
	db = startingChain()
	return db, &db[len(db)-1]
}

func startingChain() headerChainFixture {
	genesisBlock := domains.CreateGenesisHeaderBlock()
	return []domains.BlockHeader{
		genesisBlock,
	}
}

// LongestChain creates mocked the longest chain entries (containing Genesis Block and 4 first blocks).
func LongestChain() (db headerChainFixture, tip *domains.BlockHeader) {
	db = startingChain().
		addToLongestChain(HashHeight1, HeaderSourceHeight1).
		addToLongestChain(HashHeight2, HeaderSourceHeight2).
		addToLongestChain(HashHeight3, HeaderSourceHeight3).
		addToLongestChain(HashHeight4, HeaderSourceHeight4)

	return db, db.tip()
}

// AddLongestChain adds mocked longest chain to already initialized (for example with GenesisBlock) db.
func AddLongestChain(initializedDb headerChainFixture) (db headerChainFixture, tip *domains.BlockHeader) {
	withLongestChain := initializedDb.
		addToLongestChain(HashHeight1, HeaderSourceHeight1).
		addToLongestChain(HashHeight2, HeaderSourceHeight2).
		addToLongestChain(HashHeight3, HeaderSourceHeight3).
		addToLongestChain(HashHeight4, HeaderSourceHeight4)
	return withLongestChain, withLongestChain.tip()
}

// StaleChain creates mocked the stale chain entries starting and containing Genesis Block.
func StaleChain() (db headerChainFixture, tip *domains.BlockHeader) {
	db = startingChain().
		addToStaleChain(StaleHashHeight1, StaleHeaderSourceHeight1).
		addToStaleChain(StaleHashHeight2, StaleHeaderSourceHeight2).
		addToStaleChain(StaleHashHeight3, StaleHeaderSourceHeight3).
		addToStaleChain(StaleHashHeight4, StaleHeaderSourceHeight4)
	return db, db.tip()
}

// OrphanChain returns chain build from orphaned blocks.
func OrphanChain() (db headerChainFixture, tip *domains.BlockHeader) {
	orphan := BlockHeaderOf(1, OrphanHash, OrphanHeaderSource, domains.Orphan)
	db = []domains.BlockHeader{*orphan}
	return db, orphan
}

func (c headerChainFixture) addToLongestChain(hash *chainhash.Hash, hs *domains.BlockHeaderSource) headerChainFixture {
	return c.addFromSource(hash, hs, domains.LongestChain)
}

func (c headerChainFixture) addToStaleChain(hash *chainhash.Hash, hs *domains.BlockHeaderSource) headerChainFixture {
	return c.addFromSource(hash, hs, domains.Stale)
}

func (c headerChainFixture) addFromSource(hash *chainhash.Hash, hs *domains.BlockHeaderSource, s domains.HeaderState) headerChainFixture {
	height := int32(len(c))
	return append(c, *BlockHeaderOf(height, hash, hs, s))
}

func (c headerChainFixture) tip() *domains.BlockHeader {
	return &c[len(c)-1]
}

type headerChainFixture []domains.BlockHeader
