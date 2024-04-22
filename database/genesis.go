package database

import (
	"context"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/database/sql"
	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/repository/dto"
	"github.com/rs/zerolog"
)

func insertGenesisBlock(db dbAdapter, cfg *config.AppConfig, log *zerolog.Logger) error {
	hRepository := sql.NewHeadersDb(db.getDBx(), log)

	netParams := config.GetNetParams(cfg.P2P.ChainNetType)
	genesis := domains.CreateGenesisHeaderBlock(netParams.GenesisBlock.Header)

	dbGenesis := dto.ToDbBlockHeader(genesis)
	err := hRepository.Create(context.Background(), dbGenesis)
	if err != nil {
		return err
	}

	return nil
}
