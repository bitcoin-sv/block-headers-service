package repository

import (
	"context"

	"github.com/gignative-solutions/ba-p2p-headers/data/sql"
	"github.com/gignative-solutions/ba-p2p-headers/domains"
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
