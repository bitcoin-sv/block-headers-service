package repository

import (
	"context"

	"github.com/libsv/bitcoin-hc/data/sql"
	"github.com/libsv/bitcoin-hc/domains"
)

type TipRepository struct {
	db *sql.HeadersDb
}

func (r *TipRepository) GetConfirmedTip() (*domains.BlockHeader, error) {
	bh, err := r.db.GetCurrentTip(context.Background())
	return bh.ToBlockHeader(), err
}

func NewTipRepository(db *sql.HeadersDb) *TipRepository {
	return &TipRepository{db: db}
}
