package repository

import (
	"context"
	"github.com/bitcoin-sv/pulse/internal/chaincfg/chainhash"

	"github.com/bitcoin-sv/pulse/data/sql"
	"github.com/bitcoin-sv/pulse/domains"
	dto "github.com/bitcoin-sv/pulse/repository/dto"
)

// HeaderRepository provide access to repositories and implements methods for headers.
type HeaderRepository struct {
	db *sql.HeadersDb
}

// AddHeaderToDatabase adds new header to db.
// If header with given hash already exists, it will be omitted.
func (r *HeaderRepository) AddHeaderToDatabase(header domains.BlockHeader) error {
	dbHeader := dto.ToDbBlockHeader(header)
	err := r.db.Create(context.Background(), dbHeader)
	return err
}

// UpdateState changes state value to provided one for each of headers with provided hash.
func (r *HeaderRepository) UpdateState(hashes []chainhash.Hash, state domains.HeaderState) error {
	hs := make([]string, len(hashes))
	for i, h := range hashes {
		hs[i] = h.String()
	}

	err := r.db.UpdateState(context.Background(), hs, state.String())
	return err
}

// GetHeaderByHeight returns header from db by given height.
func (r *HeaderRepository) GetHeaderByHeight(height int32) (*domains.BlockHeader, error) {
	bh, err := r.db.GetHeaderByHeight(context.Background(), height, string(domains.LongestChain))
	if err == nil {
		return bh.ToBlockHeader(), nil
	}
	return nil, err
}

// GetHeaderByHeightRange returns headers from db in specified height range.
func (r *HeaderRepository) GetHeaderByHeightRange(from int, to int) ([]*domains.BlockHeader, error) {
	dbHeaders, err := r.db.GetHeaderByHeightRange(from, to)
	if err == nil {
		return dto.ConvertToBlockHeader(dbHeaders), nil
	}
	return nil, err
}

// GetLongestChainHeadersFromHeight returns from db the headers from "longest chain" starting from given height.
func (r *HeaderRepository) GetLongestChainHeadersFromHeight(height int32) ([]*domains.BlockHeader, error) {
	dbHeaders, err := r.db.GetLongestChainHeadersFromHeight(height)
	if err == nil {
		return dto.ConvertToBlockHeader(dbHeaders), nil
	}
	return nil, err
}

// GetStaleChainHeadersBackFrom returns from db all the headers with state STALE, starting from header with hash and preceding that one.
func (r *HeaderRepository) GetStaleChainHeadersBackFrom(hash string) ([]*domains.BlockHeader, error) {
	dbHeaders, err := r.db.GetStaleHeadersBackFrom(hash)
	if err == nil {
		return dto.ConvertToBlockHeader(dbHeaders), nil
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

// GetMerkleRootsConfirmations returns confirmation of merkle roots inclusion in the longest chain.
func (r *HeaderRepository) GetMerkleRootsConfirmations(merkleroots []string) ([]*domains.MerkleRootConfirmation, error) {
	mrcs, err := r.db.GetMerkleRootsConfirmations(merkleroots)
	if err != nil {
		return nil, err
	}
	return dto.ConvertToMerkleRootsConfirmations(mrcs), nil
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

// GetAncestorOnHeight provides ancestor for a hash on a specified height.
func (r *HeaderRepository) GetAncestorOnHeight(hash string, height int32) (*domains.BlockHeader, error) {
	bh, err := r.db.GetAncestorOnHeight(hash, height)
	if err == nil {
		return bh.ToBlockHeader(), nil
	}
	return nil, err
}

// GetAllTips returns all tips from db.
func (r *HeaderRepository) GetAllTips() ([]*domains.BlockHeader, error) {
	tips, err := r.db.GetAllTips()
	if err == nil {
		return dto.ConvertToBlockHeader(tips), nil
	}
	return nil, err
}

// GetChainBetweenTwoHashes calculates and returnes chain between 2 hashes.
func (r *HeaderRepository) GetChainBetweenTwoHashes(low string, high string) ([]*domains.BlockHeader, error) {
	dbHeaders, err := r.db.GetChainBetweenTwoHashes(low, high)
	if err == nil {
		return dto.ConvertToBlockHeader(dbHeaders), nil
	}
	return nil, err
}
