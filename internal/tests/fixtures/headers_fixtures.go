package fixtures

import (
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg"
)

// Default settings for tests.
const (
	DefaultBlockVersion int32  = 0x00000001
	DefaultBits         uint32 = 0x1d00ffff
)

// LONGEST CHAIN.
var (
	// HashHeight1 is a hash of first block in the LongestChain() fixture result.
	HashHeight1 = HashOf("00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048")
	// HashHeight2 is a hash of Second block in the LongestChain() fixture result.
	HashHeight2 = HashOf("000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd")
	// HashHeight3 is a hash of Third block in the LongestChain() fixture result.
	HashHeight3 = HashOf("0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449")
	// HashHeight4 is a hash of Fourth block in the LongestChain() fixture result.
	HashHeight4 = HashOf("000000004ebadb55ee9096c9a2f8880e09da59c0d68b1c228da88e48844a1485")
	// HashHeight5 is a hash of Fifth block in the LongestChain() fixture result.
	HashHeight5 = HashOf("000000009b7262315dbf071787ad3656097b892abffd1f95a1a022f896f533fc")
	// HashHeight6 is a hash of Sixth block in the LongestChain() fixture result.
	HashHeight6 = HashOf("000000003031a0e73735690c5a1ff2a4be82553b2a12b776fbd3a215dc8f778d")

	// HeaderSourceHeight1 is an exact representation of longest chain header source on that height.
	HeaderSourceHeight1 = &domains.BlockHeaderSource{
		Version:    DefaultBlockVersion,
		PrevBlock:  chaincfg.GenesisHash,
		MerkleRoot: *HashOf("0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098"),
		Timestamp:  *BlockTimestampOf("2009-01-09 02:54:25"),
		Nonce:      2573394689,
		Bits:       DefaultBits,
	}
	// HeaderSourceHeight2 is an exact representation of longest chain header source on that height.
	HeaderSourceHeight2 = &domains.BlockHeaderSource{
		Version:    DefaultBlockVersion,
		PrevBlock:  *HashHeight1,
		MerkleRoot: *HashOf("9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5"),
		Timestamp:  *BlockTimestampOf("2009-01-09 02:55:44"),
		Nonce:      1639830024,
		Bits:       DefaultBits,
	}
	// HeaderSourceHeight3 is an exact representation of longest chain header source on that height.
	HeaderSourceHeight3 = &domains.BlockHeaderSource{
		Version:    DefaultBlockVersion,
		PrevBlock:  *HashHeight2,
		MerkleRoot: *HashOf("999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644"),
		Timestamp:  *BlockTimestampOf("2009-01-09 03:02:53"),
		Nonce:      1844305925,
		Bits:       DefaultBits,
	}
	// HeaderSourceHeight4 is an exact representation of longest chain header source on that height.
	HeaderSourceHeight4 = &domains.BlockHeaderSource{
		Version:    DefaultBlockVersion,
		PrevBlock:  *HashHeight3,
		MerkleRoot: *HashOf("df2b060fa2e5e9c8ed5eaf6a45c13753ec8c63282b2688322eba40cd98ea067a"),
		Timestamp:  *BlockTimestampOf("2009-01-09 03:16:28"),
		Nonce:      2850094635,
		Bits:       DefaultBits,
	}
	// HeaderSourceHeight5 is an exact representation of longest chain header source on that height.
	HeaderSourceHeight5 = &domains.BlockHeaderSource{
		Version:    DefaultBlockVersion,
		PrevBlock:  *HashHeight4,
		MerkleRoot: *HashOf("63522845d294ee9b0188ae5cac91bf389a0c3723f084ca1025e7d9cdfe481ce1"),
		Timestamp:  *BlockTimestampOf("2009-01-09 03:23:48"),
		Nonce:      2011431709,
		Bits:       DefaultBits,
	}
	// HeaderSourceHeight6 is an exact representation of longest chain header source on that height.
	HeaderSourceHeight6 = &domains.BlockHeaderSource{
		Version:    DefaultBlockVersion,
		PrevBlock:  *HashHeight5,
		MerkleRoot: *HashOf("20251a76e64e920e58291a30d4b212939aae976baca40e70818ceaa596fb9d37"),
		Timestamp:  *BlockTimestampOf("2009-01-09 03:29:49"),
		Nonce:      2538380312,
		Bits:       DefaultBits,
	}
)

// STALE CHAIN.
var (
	//StaleHashHeight1 is a hash of first block in the StaleChain() fixture result.
	StaleHashHeight1 = HashOf("3930673fe039a7bfe4e506900c940b40e956114601ba4f59f7e21db78110e1a3")
	//StaleHashHeight2 is a hash of Second block in the StaleChain() fixture result.
	StaleHashHeight2 = HashOf("a562fd6a288f046f7a46023ce500667f74f54b893dcc047aa5faa7f4b40ee547")
	//StaleHashHeight3 is a hash of Third block in the StaleChain() fixture result.
	StaleHashHeight3 = HashOf("8e23f1eda5ad347e83638901c7072ce16f953718685d1f0521f352b8dd4a4ef7")
	//StaleHashHeight4 is a hash of Fourth block in the StaleChain() fixture result.
	StaleHashHeight4 = HashOf("34d61c9b10f013990716cc63a308e16e802e00f560c6e4ab0b7c9eb416c50e01")

	// StaleHeaderSourceHeight1 is example representation of stale chain header source on that height.
	StaleHeaderSourceHeight1 = &domains.BlockHeaderSource{
		Version:    DefaultBlockVersion,
		PrevBlock:  chaincfg.GenesisHash,
		MerkleRoot: *HashOf("cfe6b74cab677b52d850ac659d54266a3a4d7bb48270234d2fec836881d0a5f7"),
		Timestamp:  *BlockTimestampOf("2009-01-09 02:54:25"),
		Nonce:      4136106517,
		Bits:       DefaultBits,
	}
	// StaleHeaderSourceHeight2 is example representation of stale chain header source on that height.
	StaleHeaderSourceHeight2 = &domains.BlockHeaderSource{
		Version:    DefaultBlockVersion,
		PrevBlock:  *StaleHashHeight1,
		MerkleRoot: *HashOf("3b44b5ce02d36db604f0d6b7cc761685484c370a235c54539700b1ad23afefce"),
		Timestamp:  *BlockTimestampOf("2009-01-09 02:55:44"),
		Nonce:      1906126361,
		Bits:       DefaultBits,
	}
	// StaleHeaderSourceHeight3 is example representation of stale chain header source on that height.
	StaleHeaderSourceHeight3 = &domains.BlockHeaderSource{
		Version:    DefaultBlockVersion,
		PrevBlock:  *StaleHashHeight2,
		MerkleRoot: *HashOf("88d2a4e04a96b45e3ba04637098a92fd0786daf3fc8ff88314f8e739a9918bf3"),
		Timestamp:  *BlockTimestampOf("2009-01-09 03:02:53"),
		Nonce:      1334001941,
		Bits:       DefaultBits,
	}
	// StaleHeaderSourceHeight4 is example representation of stale chain header source on that height.
	StaleHeaderSourceHeight4 = &domains.BlockHeaderSource{
		Version:    DefaultBlockVersion,
		PrevBlock:  *StaleHashHeight3,
		MerkleRoot: *HashOf("88d2a4e04a96b45e3ba04637098a92fd0786daf3fc8ff88314f8e739a9918bf3"),
		Timestamp:  *BlockTimestampOf("2009-01-09 03:16:28"),
		Nonce:      1334001941,
		Bits:       DefaultBits,
	}
)

// ORPHAN CHAIN.
var (
	// OrphanHash is a hash of example orphan block.
	OrphanHash = HashOf("094697ce290d08e0c8b033754b4368026d5e64a0723951dafacf78fc342c7993")

	// OrphanHeaderSource is example representation of orphan chain header source on that height.
	OrphanHeaderSource = &domains.BlockHeaderSource{
		Version:    DefaultBlockVersion,
		PrevBlock:  *HashOf("0000000000000000000000000000000000000000000000000000000000000000"),
		MerkleRoot: *HashOf("63522845d294ee9b0188ae5cac91bf389a0c3723f084ca1025e7d9cdfe481ce1"),
		Timestamp:  *BlockTimestampOf("2009-01-09 04:04:25"),
		Nonce:      2011431709,
		Bits:       DefaultBits,
	}
)
