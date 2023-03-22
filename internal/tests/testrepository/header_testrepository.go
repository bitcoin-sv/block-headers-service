package testrepository

import (
	"errors"
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
)

type HeaderTestRepository struct {
	db []domains.BlockHeader
}

func (r *HeaderTestRepository) AddHeaderToDatabase(header domains.BlockHeader) error {
	r.db = append(r.db, header)
	return nil
}

func (r *HeaderTestRepository) UpdateState(hs []chainhash.Hash, s domains.HeaderState) error {
	for _, h := range hs {
		for i, hdb := range r.db {
			if h == hdb.Hash {
				r.db[i].State = s
			}
		}
	}
	return nil
}

func (r *HeaderTestRepository) GetHeaderByHeight(height int32) (*domains.BlockHeader, error) {
	for _, header := range r.db {
		if header.Height == height {
			return &header, nil
		}
	}
	return nil, errors.New("could not find height")
}

func (r *HeaderTestRepository) GetBlockByHash(args domains.HeaderArgs) (*domains.BlockHeader, error) {
	header := findHeader(args.Blockhash, r.db)
	if header != nil {
		return header, nil
	}
	return nil, errors.New("could not find blockhash")
}

func (r *HeaderTestRepository) GetHeaderByHeightRange(from int, to int) ([]*domains.BlockHeader, error) {
	filteredHeaders := make([]*domains.BlockHeader, 0)

	for i, header := range r.db {
		if header.Height >= int32(from) && header.Height <= int32(to) {
			filteredHeaders = append(filteredHeaders, &r.db[i])
		}
	}
	return filteredHeaders, nil
}

func (r *HeaderTestRepository) GetHeaderFromHeightToTip(height int32) ([]*domains.BlockHeader, error) {
	tip, err := r.GetTip()
	if err != nil {
		return nil, err
	}

	filteredHeaders := make([]*domains.BlockHeader, 0)

	for i, header := range r.db {
		if header.Height >= height && header.Height <= tip.Height {
			filteredHeaders = append(filteredHeaders, &r.db[i])
		}
	}
	return filteredHeaders, nil
}

func (r *HeaderTestRepository) GetLongestChainHeadersFromHeight(height int32) ([]*domains.BlockHeader, error) {
	filteredHeaders := make([]*domains.BlockHeader, 0)

	for i, header := range r.db {
		if header.Height >= height && header.State == domains.LongestChain {
			filteredHeaders = append(filteredHeaders, &r.db[i])
		}
	}
	return filteredHeaders, nil
}

func (r *HeaderTestRepository) GetStaleChainHeadersBackFrom(hash string) ([]*domains.BlockHeader, error) {
	filteredHeaders := make([]*domains.BlockHeader, 0)

	header, _ := r.GetHeaderByHash(hash)

	for h := header; h.State == domains.Stale; {
		filteredHeaders = append(filteredHeaders, h)
		h, _ = r.GetHeaderByHash(h.PreviousBlock.String())
	}

	return filteredHeaders, nil
}

func (r *HeaderTestRepository) GetPreviousHeader(hash string) (*domains.BlockHeader, error) {
	header := findHeader(hash, r.db)
	if header != nil {
		prevHeader := findHeader(header.Hash.String(), r.db)
		if prevHeader != nil {
			return prevHeader, nil
		}
	}
	return nil, errors.New("could not find header")
}

func (r *HeaderTestRepository) GetCurrentHeight() (int, error) {
	highestHeader := domains.BlockHeader{}
	for _, header := range r.db {
		if header.Height > highestHeader.Height {
			highestHeader = header
		}
	}
	return int(highestHeader.Height), nil
}

func (r *HeaderTestRepository) GetHeadersCount() (int, error) {
	return len(r.db), nil
}

func (r *HeaderTestRepository) GetHeaderByHash(hash string) (*domains.BlockHeader, error) {
	header := findHeader(hash, r.db)
	if header != nil {
		return header, nil
	}
	return nil, errors.New("could not find hash")
}

func (r *HeaderTestRepository) GetConfirmationsCountForBlock(hash string) (int, error) {
	header := findHeader(hash, r.db)
	if header == nil {
		return 0, errors.New("could not find blockhash")
	}

	count := 0
	for header != nil {
		header := findHeader(hash, r.db)
		if header != nil {
			count++
		}
	}
	return count, nil
}

func (r *HeaderTestRepository) GenesisExists() bool {
	for _, header := range r.db {
		if header.Height == 0 {
			return true
		}
	}
	return false
}

func (r *HeaderTestRepository) GetTip() (*domains.BlockHeader, error) {
	highestHeader := domains.BlockHeader{}
	for _, header := range r.db {
		if header.Height > highestHeader.Height {
			highestHeader = header
		}
	}
	return &highestHeader, nil
}

func findHeader(hash string, headers []domains.BlockHeader) *domains.BlockHeader {
	for _, header := range headers {
		if header.Hash.String() == hash {
			return &header
		}
	}
	return nil
}

func NewHeadersTestRepository(db []domains.BlockHeader) *HeaderTestRepository {
	return &HeaderTestRepository{
		db: db,
	}
}
