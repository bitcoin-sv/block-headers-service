package domains

import (
	"math/big"
	"strconv"
	"time"

	"github.com/libsv/bitcoin-hc/internal/chaincfg"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
)

// HeaderState enum representing header state.
type HeaderState string

const (
	// Orphan headers that are not connected to existing block or are connected with other orphan block.
	Orphan HeaderState = "ORPHAN"
	// Stale headers that are not part of the longest chain - it means they're forms concurrent chain to the longest chain.
	Stale HeaderState = "STALE"
	// LongestChain headers that are part of the longest chain - it means they're forms chain which cumulated chain work is the biggest.
	LongestChain HeaderState = "LONGEST_CHAIN"
	// Rejected headers that are on the black list of headers.
	Rejected HeaderState = "REJECTED"
)

func (s *HeaderState) String() string {
	return string(*s)
}

// BlockHeader defines a single block header, used in SPV validations.
type BlockHeader struct {
	Height           int32          `json:"-"`
	Hash             chainhash.Hash `json:"hash"`
	Version          int32          `json:"version"`
	MerkleRoot       chainhash.Hash `json:"merkleRoot"`
	Timestamp        time.Time      `json:"creationTimestamp"`
	Bits             uint32         `json:"-"`
	Nonce            uint32         `json:"nonce"`
	State            HeaderState    `json:"-"`
	Chainwork        uint64         `json:"-"`
	CumulatedWork    *big.Int       `json:"work"`
	PreviousBlock    chainhash.Hash `json:"prevBlockHash"`
	DifficultyTarget uint32         `json:"difficultyTarget"`
}

// DbBlockHeader represent header saved in db.
type DbBlockHeader struct {
	Height           int32     `db:"height"`
	Hash             string    `db:"hash"`
	Version          int32     `db:"version"`
	MerkleRoot       string    `db:"merkleroot"`
	Timestamp        time.Time `db:"timestamp"`
	Bits             uint32    `db:"bits"`
	Nonce            uint32    `db:"nonce"`
	State            string    `db:"header_state"`
	Chainwork        string    `db:"chainwork"`
	CumulatedWork    string    `db:"cumulatedWork"`
	PreviousBlock    string    `db:"previousblock"`
	DifficultyTarget uint32    `db:"difficultytarget"`
}

// HeaderArgs are sued to retrieve a single block header.
type HeaderArgs struct {
	Blockhash string `param:"blockhash" db:"blockHash"`
}

// BlockHeaderState is an extended version of the BlockHeader
// that has more important informations. Mostly used in http server endpoints.
type BlockHeaderState struct {
	Header        BlockHeader `json:"header"`
	State         string      `json:"state"`
	ChainWork     uint64      `json:"chainWork"`
	Height        int32       `json:"height"`
	Confirmations int         `json:"confirmations"`
}

// BlockHeaderSource defines source of information about a block header used by system.
type BlockHeaderSource struct {
	// Version of the block. This is not the same as the protocol version.
	Version int32

	// Hash of the previous block header in the block chain.
	PrevBlock chainhash.Hash

	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot chainhash.Hash

	// Time the block was created.
	Timestamp time.Time

	// Difficulty target for the block.
	Bits uint32

	// Nonce used to generate the block.
	Nonce uint32
}

// BlockHash is a representation of Hash of the block.
type BlockHash chainhash.Hash

// String returns the Hash as the hexadecimal string of the byte-reversed
// hash.
func (h *BlockHash) String() string {
	return h.ChainHash().String()
}

// ChainHash returns chainhash.Hash representation of the block hash.
func (h *BlockHash) ChainHash() chainhash.Hash {
	return chainhash.Hash(*h)
}

// CreateHeader constructor for BlockHeader.
func CreateHeader(hash *BlockHash, bs *BlockHeaderSource, ph *BlockHeader) BlockHeader {
	cw := CalculateWork(bs.Bits)
	ccw := CumulatedChainWorkOf(*ph.CumulatedWork).Add(cw)

	var state HeaderState
	if ph.IsOrphan() {
		state = Orphan
	} else {
		state = LongestChain
	}

	return BlockHeader{
		Height:        ph.Height + 1,
		Hash:          hash.ChainHash(),
		Version:       bs.Version,
		MerkleRoot:    bs.MerkleRoot,
		Timestamp:     bs.Timestamp,
		Bits:          bs.Bits,
		Nonce:         bs.Nonce,
		State:         state,
		Chainwork:     cw.Uint64(),
		CumulatedWork: ccw.BigInt(),
		PreviousBlock: bs.PrevBlock,
	}
}

// NewRejectedBlockHeader constructs rejected block header.
func NewRejectedBlockHeader(hash BlockHash) *BlockHeader {
	return &BlockHeader{
		Hash:  chainhash.Hash(hash),
		State: Rejected,
	}
}

// NewOrphanPreviousBlockHeader constructor for previous block for orphaned block.
func NewOrphanPreviousBlockHeader() *BlockHeader {
	return &BlockHeader{
		Height:        0,
		State:         Orphan,
		Bits:          0,
		CumulatedWork: big.NewInt(0),
	}
}

// IsOrphan is the block an orphan.
func (bh *BlockHeader) IsOrphan() bool {
	return bh.State == Orphan
}

// IsLongestChain is the block part of the longest chain.
func (bh *BlockHeader) IsLongestChain() bool {
	return bh.State == LongestChain
}

// ConvertToBlockHeader converts one or whole slice of DbBlockHeaders to BlockHeaders
// used after getting records from db.
func ConvertToBlockHeader(dbBlockHeaders []*DbBlockHeader) []*BlockHeader {
	if dbBlockHeaders != nil {
		var blockHeaders []*BlockHeader

		for _, header := range dbBlockHeaders {
			h := header.ToBlockHeader()
			blockHeaders = append(blockHeaders, h)
		}
		return blockHeaders
	}
	return nil
}

// ToBlockHeader converts work from string to big.Int and return BlockHeader.
func (dbh *DbBlockHeader) ToBlockHeader() *BlockHeader {
	if dbh.CumulatedWork == "" {
		dbh.CumulatedWork = "0"
	}
	cumulatedWork, ok := new(big.Int).SetString(dbh.CumulatedWork, 10)
	if !ok {
		cumulatedWork = big.NewInt(0)
	}

	chainWork, err := strconv.ParseUint(dbh.Chainwork, 10, 64)
	if err != nil {
		chainWork = 0
	}

	hash, _ := chainhash.NewHashFromStr(dbh.Hash)
	merkleTree, _ := chainhash.NewHashFromStr(dbh.MerkleRoot)
	prevBlock, _ := chainhash.NewHashFromStr(dbh.PreviousBlock)

	return &BlockHeader{
		Height:           dbh.Height,
		Hash:             *hash,
		Version:          dbh.Version,
		MerkleRoot:       *merkleTree,
		Timestamp:        dbh.Timestamp,
		Bits:             dbh.Bits,
		Nonce:            dbh.Nonce,
		Chainwork:        chainWork,
		CumulatedWork:    cumulatedWork,
		State:            HeaderState(dbh.State),
		PreviousBlock:    *prevBlock,
		DifficultyTarget: dbh.DifficultyTarget,
	}
}

// ToDbBlockHeader converts BlockHeader to DbBlockHeader
// used mainly to prepare record befor saving in db.
func (bh BlockHeader) ToDbBlockHeader() DbBlockHeader {
	return DbBlockHeader{
		Height:           bh.Height,
		Hash:             bh.Hash.String(),
		Version:          bh.Version,
		MerkleRoot:       bh.MerkleRoot.String(),
		Timestamp:        bh.Timestamp,
		Bits:             bh.Bits,
		Nonce:            bh.Nonce,
		State:            bh.State.String(),
		Chainwork:        strconv.FormatUint(bh.Chainwork, 10),
		CumulatedWork:    bh.CumulatedWork.String(),
		PreviousBlock:    bh.PreviousBlock.String(),
		DifficultyTarget: bh.DifficultyTarget,
	}
}

// WrapWithHeaderState wraps BlockHeader with additional information creating BlockHeaderState.
func (bh *BlockHeader) WrapWithHeaderState() BlockHeaderState {
	model := BlockHeaderState{
		Header:    *bh,
		State:     bh.State.String(),
		Height:    bh.Height,
		ChainWork: bh.Chainwork,
	}

	return model
}

// CreateGenesisHeaderBlock create filled genesis block.
func CreateGenesisHeaderBlock() BlockHeader {
	// Create a new node from the genesis block and set it as the best node.
	genesisBlock := BlockHeader{
		Hash:          chaincfg.GenesisHash,
		Height:        0,
		Version:       1,
		PreviousBlock: chainhash.Hash{},           // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot:    chaincfg.GenesisMerkleRoot, // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		Timestamp:     time.Unix(0x495fab29, 0),   // 2009-01-03 18:15:05 +0000 UTC
		Bits:          0x1d00ffff,
		Nonce:         0x7c2bac1d,
		State:         LongestChain,
		CumulatedWork: big.NewInt(0),
	}

	return genesisBlock
}

// FastLog2Floor calculates the floor of the base-2 logarithm of an input 32-bit
// unsigned integer using a bitwise algorithm that masks off decreasingly lower-order bits
// of the integer until it reaches the highest order bit, and returns the resulting integer value.
func FastLog2Floor(n uint32) uint8 {
	var log2FloorMasks = []uint32{0xffff0000, 0xff00, 0xf0, 0xc, 0x2}
	rv := uint8(0)
	exponent := uint8(16)
	for i := 0; i < 5; i++ {
		if n&log2FloorMasks[i] != 0 {
			rv += exponent
			n >>= exponent
		}
		exponent >>= 1
	}
	return rv
}
