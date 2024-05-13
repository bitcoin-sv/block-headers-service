package database

import (
	"context"
	"time"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/database/sql"
	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/bitcoin-sv/block-headers-service/repository/dto"
	"github.com/rs/zerolog"
)

func insertGenesisBlock(db dbAdapter, cfg *config.AppConfig, log *zerolog.Logger) error {
	hRepository := sql.NewHeadersDb(db.getDBx(), log)
	netParams := cfg.P2P.GetNetParams()

	genesis := createGenesisHeaderBlock(netParams.GenesisBlock.Header)

	err := hRepository.Create(context.Background(), genesis)
	if err != nil {
		return err
	}

	return nil
}

// CreateGenesisHeaderBlock create filled genesis block based on the chosen chain net header block.
func createGenesisHeaderBlock(genesisBlockHeader wire.BlockHeader) dto.DbBlockHeader {
	longestChain := domains.LongestChain
	genesisBlock := dto.DbBlockHeader{
		Hash:          genesisBlockHeader.BlockHash().String(),
		Height:        0,
		Version:       1,
		PreviousBlock: chainhash.Hash{}.String(),              // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot:    genesisBlockHeader.MerkleRoot.String(), // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		Timestamp:     time.Unix(genesisBlockHeader.Timestamp.Unix(), 0),
		Bits:          genesisBlockHeader.Bits,
		Nonce:         genesisBlockHeader.Nonce,
		State:         longestChain.String(),
		Chainwork:     domains.CalculateWork(genesisBlockHeader.Bits).BigInt().String(),
		CumulatedWork: domains.CalculateWork(genesisBlockHeader.Bits).BigInt().String(),
	}

	return genesisBlock
}
