package repository

import (
	"context"

	"github.com/libsv/bitcoin-hc/data/sql"
	"github.com/libsv/bitcoin-hc/domains"
)

// HeaderRepository provide access to repositories and implements methods for headers.
type HeaderRepository struct {
	db *sql.HeadersDb
}

// AddHeaderToDatabase adds new header to db.
func (r *HeaderRepository) AddHeaderToDatabase(header domains.BlockHeader) error {
	dbHeader := header.ToDbBlockHeader()
	err := r.db.Create(context.Background(), dbHeader)
	return err
}

// GetHeaderByHeight returns header from db by given height.
func (r *HeaderRepository) GetHeaderByHeight(height int32) (*domains.BlockHeader, error) {
	bh, err := r.db.GetHeaderByHeight(context.Background(), height)
	if err == nil {
		return bh.ToBlockHeader(), nil
	}
	return nil, err
}

// GetBlockByHash returns header from db by given arguments.
func (r *HeaderRepository) GetBlockByHash(args domains.HeaderArgs) (*domains.BlockHeader, error) {
	bh, err := r.db.Header(context.Background(), args)
	if err == nil {
		return bh.ToBlockHeader(), nil
	}
	return nil, err
}

// GetHeaderByHeightRange returns headers from db in specified height range.
func (r *HeaderRepository) GetHeaderByHeightRange(from int, to int) ([]*domains.BlockHeader, error) {
	dbHeaders, err := r.db.GetHeaderByHeightRange(from, to)
	if err == nil {
		return domains.ConvertToBlockHeader(dbHeaders), nil
	}
	return nil, err
}

// GetPreviousHeader returns previous header from the one with given hash.
func (r *HeaderRepository) GetPreviousHeader(hash string) (*domains.BlockHeader, error) {
	bh, err := r.db.GetPreviousHeader(context.Background(), hash)
	if err == nil {
		return bh.ToBlockHeader(), nil
	}
	return nil, err
}

// GetCurrentHeight returns current highest block hight in db.
func (r *HeaderRepository) GetCurrentHeight() (int, error) {
	return r.db.Height(context.Background())
}

// GetHeadersCount returns number of headers stored in db.
func (r *HeaderRepository) GetHeadersCount() (int, error) {
	return r.db.Count(context.Background())
}

// GetHeaderByHash returns header from db by given hash.
func (r *HeaderRepository) GetHeaderByHash(hash string) (*domains.BlockHeader, error) {
	bh, err := r.db.GetHeaderByHash(context.Background(), hash)
	if err == nil {
		return bh.ToBlockHeader(), nil
	}
	return nil, err
}

// GetConfirmationsCountForBlock returns number of confirmations for header with given hash.
func (r *HeaderRepository) GetConfirmationsCountForBlock(hash string) (int, error) {
	return r.db.CalculateConfirmations(context.Background(), hash)
}

// GenesisExists check if genesis header is in db.
func (r *HeaderRepository) GenesisExists() bool {
	return r.db.GenesisExists(context.Background())
}

// NewHeadersRepository creates and returns HeaderRepository instance.
func NewHeadersRepository(db *sql.HeadersDb) *HeaderRepository {
	return &HeaderRepository{db: db}
}

// GetTip returns tip from db.
func (r *HeaderRepository) GetTip() (*domains.BlockHeader, error) {
	tip, err := r.db.GetTip(context.Background())
	if tip == nil {
		return nil, err
	}
	header := tip.ToBlockHeader()
	return header, err
}
