package domains

import (
	"math/big"
	"strconv"
	"time"

	"github.com/libsv/bitcoin-hc/internal/chaincfg"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
)

// BlockHeader defines a single block header, used in SPV validations.
type BlockHeader struct {
	Height           int32          `json:"-"`
	Hash             chainhash.Hash `json:"hash"`
	Version          int32          `json:"version"`
	MerkleRoot       chainhash.Hash `json:"merkleRoot"`
	Timestamp        time.Time      `json:"creationTimestamp"`
	Bits             uint32         `json:"-"`
	Nonce            uint32         `json:"nonce"`
	Chainwork        uint64         `json:"-"`
	CumulatedWork    *big.Int       `json:"work"`
	IsOrphan         bool           `json:"-"`
	IsConfirmed      bool           `json:"-"`
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
	Chainwork        string    `db:"chainwork"`
	CumulatedWork    string    `db:"cumulatedWork"`
	IsOrphan         bool      `db:"isorphan"`
	IsConfirmed      bool      `db:"isconfirmed"`
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
func (dbBlockHeader *DbBlockHeader) ToBlockHeader() *BlockHeader {
	if dbBlockHeader.CumulatedWork == "" {
		dbBlockHeader.CumulatedWork = "0"
	}
	cumulatedWork, ok := new(big.Int).SetString(dbBlockHeader.CumulatedWork, 10)
	if !ok {
		cumulatedWork = big.NewInt(0)
	}

	chainWork, err := strconv.ParseUint(dbBlockHeader.Chainwork, 10, 64)
	if err != nil {
		chainWork = 0
	}

	hash, _ := chainhash.NewHashFromStr(dbBlockHeader.Hash)
	merkleTree, _ := chainhash.NewHashFromStr(dbBlockHeader.MerkleRoot)
	prevBlock, _ := chainhash.NewHashFromStr(dbBlockHeader.PreviousBlock)

	return &BlockHeader{
		Height:           dbBlockHeader.Height,
		Hash:             *hash,
		Version:          dbBlockHeader.Version,
		MerkleRoot:       *merkleTree,
		Timestamp:        dbBlockHeader.Timestamp,
		Bits:             dbBlockHeader.Bits,
		Nonce:            dbBlockHeader.Nonce,
		Chainwork:        chainWork,
		CumulatedWork:    cumulatedWork,
		IsOrphan:         dbBlockHeader.IsOrphan,
		IsConfirmed:      dbBlockHeader.IsConfirmed,
		PreviousBlock:    *prevBlock,
		DifficultyTarget: dbBlockHeader.DifficultyTarget,
	}
}

// ToDbBlockHeader converts BlockHeader to DbBlockHeader
// used mainly to prepare record befor saving in db.
func (blockHeader BlockHeader) ToDbBlockHeader() DbBlockHeader {
	return DbBlockHeader{
		Height:           blockHeader.Height,
		Hash:             blockHeader.Hash.String(),
		Version:          blockHeader.Version,
		MerkleRoot:       blockHeader.MerkleRoot.String(),
		Timestamp:        blockHeader.Timestamp,
		Bits:             blockHeader.Bits,
		Nonce:            blockHeader.Nonce,
		Chainwork:        strconv.FormatUint(blockHeader.Chainwork, 10),
		CumulatedWork:    blockHeader.CumulatedWork.String(),
		IsOrphan:         blockHeader.IsOrphan,
		IsConfirmed:      blockHeader.IsConfirmed,
		PreviousBlock:    blockHeader.PreviousBlock.String(),
		DifficultyTarget: blockHeader.DifficultyTarget,
	}
}

// CumulateWork sums up cumulatedWork from previous header with chainwork from new header.
func (blockHeader *BlockHeader) CumulateWork(prevWork *big.Int) {
	work := prevWork
	if work == nil {
		work = big.NewInt(0)
	}
	blockHeader.CumulatedWork = work.Add(new(big.Int).SetUint64(blockHeader.Chainwork), work)
}

// WrapWithHeaderState wraps BlockHeader with additional information creating BlockHeaderState.
func (blockHeader *BlockHeader) WrapWithHeaderState() BlockHeaderState {
	model := BlockHeaderState{
		Header:    *blockHeader,
		State:     "LONGEST_CHAIN",
		Height:    blockHeader.Height,
		ChainWork: blockHeader.Chainwork,
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
		CumulatedWork: big.NewInt(0),
	}

	return genesisBlock
}

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
