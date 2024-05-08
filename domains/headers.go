package domains

import (
	"math/big"
	"time"

	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
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
	Height        int32          `json:"-"`
	Hash          chainhash.Hash `json:"hash"`
	Version       int32          `json:"version"`
	MerkleRoot    chainhash.Hash `json:"merkleRoot"`
	Timestamp     time.Time      `json:"creationTimestamp"`
	Bits          uint32         `json:"-"`
	Nonce         uint32         `json:"nonce"`
	State         HeaderState    `json:"-"`
	Chainwork     *big.Int       `json:"-"`
	CumulatedWork *big.Int       `json:"work"`
	PreviousBlock chainhash.Hash `json:"prevBlockHash"`
}

// MerkleRootConfirmationRequestItem is a request type for verification
// of Merkle Roots inclusion in the longest chain.
type MerkleRootConfirmationRequestItem struct {
	MerkleRoot  string `json:"merkleRoot"`
	BlockHeight int32  `json:"blockHeight"`
}

// MerkleRootConfirmationState represents the state of each Merkle Root verification
// process and can be one of three values: Confirmed, Invalid and UnableToVerify.
type MerkleRootConfirmationState string

const (
	// Confirmed state occurs when Merkle Root is found in the longest chain.
	Confirmed MerkleRootConfirmationState = "CONFIRMED"
	// UnableToVerify state occurs when Block Headers Service is behind in synchronization with the longest chain.
	UnableToVerify MerkleRootConfirmationState = "UNABLE_TO_VERIFY"
	// Invalid state occurs when Merkle Root is not found in the longest chain.
	Invalid MerkleRootConfirmationState = "INVALID"
)

// MerkleRootConfirmation is used to confirm the inclusion of
// Merkle Roots in the longest chain.
type MerkleRootConfirmation struct {
	MerkleRoot   string                      `json:"merkleRoot"`
	BlockHeight  int32                       `json:"blockHeight"`
	Hash         string                      `json:"hash,omitempty"`
	Confirmation MerkleRootConfirmationState `json:"confirmation"`
}

// HeaderArgs are used to retrieve a single block header.
type HeaderArgs struct {
	Blockhash string `param:"blockhash" db:"blockHash"`
}

// BlockHeaderState is an extended version of the BlockHeader
// that has more important informations. Mostly used in http server endpoints.
type BlockHeaderState struct {
	Header    BlockHeader `json:"header"`
	State     string      `json:"state"`
	ChainWork *big.Int    `json:"chainWork" swaggertype:"string"`
	Height    int32       `json:"height"`
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
	} else if ph.IsLongestChain() {
		state = LongestChain
	} else {
		state = Stale
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
		Chainwork:     cw.BigInt(),
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

// FastLog2Floor calculates the floor of the base-2 logarithm of an input 32-bit
// unsigned integer using a bitwise algorithm that masks off decreasingly lower-order bits
// of the integer until it reaches the highest order bit, and returns the resulting integer value.
func FastLog2Floor(n uint32) uint8 {
	log2FloorMasks := []uint32{0xffff0000, 0xff00, 0xf0, 0xc, 0x2}
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
