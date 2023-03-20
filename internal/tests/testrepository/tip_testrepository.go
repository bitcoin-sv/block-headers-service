package testrepository

import (
	"github.com/libsv/bitcoin-hc/domains"
)

type TipTestRepository struct {
	db []domains.BlockHeader
}

func (r *TipTestRepository) GetConfirmedTip() (*domains.BlockHeader, error) {
	return nil, nil
}

func NewTestTipRepository(db []domains.BlockHeader) *TipTestRepository {
	return &TipTestRepository{
		db: db,
	}
}
