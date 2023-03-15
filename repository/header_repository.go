package repository

import (
	"context"

	"github.com/libsv/bitcoin-hc/data/sql"
	"github.com/libsv/bitcoin-hc/domains"
)

type HeaderRepository struct {
	db *sql.HeadersDb
}

func (r *HeaderRepository) AddHeaderToDatabase(header domains.BlockHeader) error {
	dbHeader := header.ToDbBlockHeader()
	err := r.db.Create(context.Background(), dbHeader)
	return err
}

func (r *HeaderRepository) GetHeaderByHeight(height int32) (*domains.BlockHeader, error) {
	bh, err := r.db.GetHeaderByHeight(context.Background(), height)
	if err == nil {
		return bh.ToBlockHeader(), nil
	}
	return nil, err
}

func (r *HeaderRepository) GetBlockByHash(args domains.HeaderArgs) (*domains.BlockHeader, error) {
	bh, err := r.db.Header(context.Background(), args)
	if err == nil {
		return bh.ToBlockHeader(), nil
	}
	return nil, err
}

func (r *HeaderRepository) GetHeaderByHeightRange(from int, to int) ([]*domains.BlockHeader, error) {
	dbHeaders, err := r.db.GetHeaderByHeightRange(from, to)
	if err == nil {
		return domains.ConvertToBlockHeader(dbHeaders), nil
	}
	return nil, err
}

func (r *HeaderRepository) GetPreviousHeader(hash string) (*domains.BlockHeader, error) {
	bh, err := r.db.GetPreviousHeader(context.Background(), hash)
	if err == nil {
		return bh.ToBlockHeader(), err
	}
	return nil, err
}

func (r *HeaderRepository) GetCurrentHeight() (int, error) {
	return r.db.Height(context.Background())
}

func (r *HeaderRepository) GetHeadersCount() (int, error) {
	return r.db.Count(context.Background())
}

func (r *HeaderRepository) GetHeaderByHash(hash string) (*domains.BlockHeader, error) {
	bh, err := r.db.GetHeaderByHash(context.Background(), hash)
	if err == nil {
		return bh.ToBlockHeader(), nil
	}
	return nil, err
}

func (r *HeaderRepository) GetConfirmationsCountForBlock(hash string) (int, error) {
	return r.db.CalculateConfirmations(context.Background(), hash)
}

func (r *HeaderRepository) GenesisExists() bool {
	return r.db.GenesisExists(context.Background())
}

func NewHeadersRepository(db *sql.HeadersDb) *HeaderRepository {
	return &HeaderRepository{db: db}
}

func (r *HeaderRepository) GetTip() (*domains.BlockHeader, error) {
	tip, err := r.db.GetTip(context.Background())
	if tip == nil {
		return nil, err
	}
	header := tip.ToBlockHeader()
	return header, err
}
