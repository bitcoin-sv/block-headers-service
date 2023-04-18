package dto

import (
	"math/big"
	"strconv"
	"time"

	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
)

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

// ToBlockHeader converts work from string to big.Int and return BlockHeader.
func (dbh *DbBlockHeader) ToBlockHeader() *domains.BlockHeader {
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

	return &domains.BlockHeader{
		Height:           dbh.Height,
		Hash:             *hash,
		Version:          dbh.Version,
		MerkleRoot:       *merkleTree,
		Timestamp:        dbh.Timestamp,
		Bits:             dbh.Bits,
		Nonce:            dbh.Nonce,
		Chainwork:        chainWork,
		CumulatedWork:    cumulatedWork,
		State:            domains.HeaderState(dbh.State),
		PreviousBlock:    *prevBlock,
		DifficultyTarget: dbh.DifficultyTarget,
	}
}

// ConvertToBlockHeader converts one or whole slice of DbBlockHeaders to BlockHeaders
// used after getting records from db.
func ConvertToBlockHeader(dbBlockHeaders []*DbBlockHeader) []*domains.BlockHeader {
	if dbBlockHeaders != nil {
		var blockHeaders []*domains.BlockHeader

		for _, header := range dbBlockHeaders {
			h := header.ToBlockHeader()
			blockHeaders = append(blockHeaders, h)
		}
		return blockHeaders
	}
	return nil
}

// ConvertToBlockHeaderState converts one or whole slice of DbBlockHeaders to BlockHeaderState
// used after getting records from db.
func ConvertToBlockHeaderState(dbBlockHeaders []*DbBlockHeader) []*domains.BlockHeaderState {
	if dbBlockHeaders != nil {
		var blockHeaders []*domains.BlockHeaderState

		for _, header := range dbBlockHeaders {
			h := header.ToBlockHeader()
			hs := &domains.BlockHeaderState{
				Header:        *h,
				State:         h.State.String(),
				Height:        h.Height,
				ChainWork:     h.Chainwork,
				Confirmations: 1,
			}
			blockHeaders = append(blockHeaders, hs)
		}
		return blockHeaders
	}
	return nil
}

// ToDbBlockHeader converts BlockHeader to DbBlockHeader
// used mainly to prepare record befor saving in db.
func ToDbBlockHeader(bh domains.BlockHeader) DbBlockHeader {
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
