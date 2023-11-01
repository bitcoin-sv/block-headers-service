package testrepository

import (
	"errors"

	"github.com/bitcoin-sv/pulse/domains"
	"github.com/bitcoin-sv/pulse/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/pulse/internal/tests/fixtures"
)

// HeaderTestRepository in memory HeadersRepository representation for unit testing.
type HeaderTestRepository struct {
	db *[]domains.BlockHeader
}

// AddHeaderToDatabase adds new header to db.
// If header with this same hash already exists, it will not be added.
func (r *HeaderTestRepository) AddHeaderToDatabase(header domains.BlockHeader) error {
	for _, hdb := range *r.db {
		if header.Hash == hdb.Hash {
			return nil
		}
	}
	*r.db = append(*r.db, header)
	return nil
}

// UpdateState changes state value to provided one for each of headers with provided hash.
func (r *HeaderTestRepository) UpdateState(hs []chainhash.Hash, s domains.HeaderState) error {
	for _, h := range hs {
		for i, hdb := range *r.db {
			if h == hdb.Hash {
				(*r.db)[i].State = s
			}
		}
	}
	return nil
}

// GetHeaderByHeight returns header from db by given height.
func (r *HeaderTestRepository) GetHeaderByHeight(height int32) (*domains.BlockHeader, error) {
	for _, header := range *r.db {
		if header.Height == height {
			return &header, nil
		}
	}
	return nil, errors.New("could not find height")
}

// GetHeaderByHeightRange returns headers from db in specified height range.
func (r *HeaderTestRepository) GetHeaderByHeightRange(from int, to int) ([]*domains.BlockHeader, error) {
	filteredHeaders := make([]*domains.BlockHeader, 0)

	for i, header := range *r.db {
		if header.Height >= int32(from) && header.Height <= int32(to) {
			filteredHeaders = append(filteredHeaders, &(*r.db)[i])
		}
	}

	if len(filteredHeaders) > 0 {
		return filteredHeaders, nil
	}

	return nil, errors.New("could not find headers in given range")
}

// GetLongestChainHeadersFromHeight returns from db the headers from "longest chain" starting from given height.
func (r *HeaderTestRepository) GetLongestChainHeadersFromHeight(height int32) ([]*domains.BlockHeader, error) {
	filteredHeaders := make([]*domains.BlockHeader, 0)

	for i, header := range *r.db {
		if header.Height >= height && header.State == domains.LongestChain {
			filteredHeaders = append(filteredHeaders, &(*r.db)[i])
		}
	}
	return filteredHeaders, nil
}

// GetStaleChainHeadersBackFrom returns from db all the headers with state STALE, starting from header with hash and preceding that one.
func (r *HeaderTestRepository) GetStaleChainHeadersBackFrom(hash string) ([]*domains.BlockHeader, error) {
	filteredHeaders := make([]*domains.BlockHeader, 0)

	header, _ := r.GetHeaderByHash(hash)

	for h := header; h.State == domains.Stale; {
		filteredHeaders = append(filteredHeaders, h)
		h, _ = r.GetHeaderByHash(h.PreviousBlock.String())
	}

	return filteredHeaders, nil
}

// GetPreviousHeader returns previous header from the one with given hash.
func (r *HeaderTestRepository) GetPreviousHeader(hash string) (*domains.BlockHeader, error) {
	header := findHeader(hash, *r.db)
	if header != nil {
		prevHeader := findHeader(header.Hash.String(), *r.db)
		if prevHeader != nil {
			return prevHeader, nil
		}
	}
	return nil, errors.New("could not find header")
}

// GetCurrentHeight returns current highest block hight in db.
func (r *HeaderTestRepository) GetCurrentHeight() (int, error) {
	highestHeader := domains.BlockHeader{}
	for _, header := range *r.db {
		if header.Height > highestHeader.Height {
			highestHeader = header
		}
	}
	return int(highestHeader.Height), nil
}

// GetHeadersCount returns number of headers stored in db.
func (r *HeaderTestRepository) GetHeadersCount() (int, error) {
	return len(*r.db), nil
}

// GetHeaderByHash returns header from db by given hash.
func (r *HeaderTestRepository) GetHeaderByHash(hash string) (*domains.BlockHeader, error) {
	header := findHeader(hash, *r.db)
	if header != nil {
		return header, nil
	}
	return nil, errors.New("could not find hash")
}

// GenesisExists check if genesis header is in db.
func (r *HeaderTestRepository) GenesisExists() bool {
	for _, header := range *r.db {
		if header.Height == 0 {
			return true
		}
	}
	return false
}

// GetTip returns tip from db.
func (r *HeaderTestRepository) GetTip() (*domains.BlockHeader, error) {
	highestHeader := domains.BlockHeader{}
	for _, header := range *r.db {
		if header.Height > highestHeader.Height {
			highestHeader = header
		}
	}
	return &highestHeader, nil
}

// GetAncestorOnHeight returns ancestor header from db on given height.
func (r *HeaderTestRepository) GetAncestorOnHeight(hash string, height int32) (*domains.BlockHeader, error) {
	for _, header := range *r.db {
		if header.Height == height {
			return &header, nil
		}
	}
	return nil, errors.New("could not find height")
}

// GetAllTips returns all tips from db.
func (r *HeaderTestRepository) GetAllTips() ([]*domains.BlockHeader, error) {
	prevHashes := make([]string, 0)
	tips := make([]*domains.BlockHeader, 0)

	for _, h := range *r.db {
		prevHashes = append(prevHashes, h.PreviousBlock.String())
	}

	for i, h := range *r.db {
		if !contains(prevHashes, h.Hash.String()) {
			tips = append(tips, &(*r.db)[i])
		}
	}

	return tips, nil
}

// GetChainBetweenTwoHashes returns all headers between two hashes.
func (r *HeaderTestRepository) GetChainBetweenTwoHashes(low string, high string) ([]*domains.BlockHeader, error) {
	hLow := findHeader(low, *r.db)
	hHigh := findHeader(high, *r.db)
	headers, err := r.GetHeaderByHeightRange(int(hLow.Height), int(hHigh.Height))
	if err != nil {
		return nil, err
	}
	return headers, nil
}

// GetMerkleRootsConfirmations returns a confirmation of merkle roots inclusion
// in the longest chain with hash of the block in which the merkle root is included.
func (r *HeaderTestRepository) GetMerkleRootsConfirmations(
	request []domains.MerkleRootConfirmationRequestItem,
	maxBlockHeightExcess int,
) ([]*domains.MerkleRootConfirmation, error) {
	mrcfs := make([]*domains.MerkleRootConfirmation, 0)

	var topHeight int32 = 0
	for _, h := range *r.db {
		if h.Height > topHeight {
			topHeight = h.Height
		}
	}

	for _, rq := range request {
		found := false
		confm := &domains.MerkleRootConfirmation{
			MerkleRoot:  rq.MerkleRoot,
			BlockHeight: rq.BlockHeight,
		}

		for _, h := range *r.db {
			if h.MerkleRoot.String() == rq.MerkleRoot && h.Height == rq.BlockHeight {
				found = true
				confm.Hash = h.Hash.String()
				confm.Confirmation = domains.Confirmed
				break
			}
		}

		if !found {
			if rq.BlockHeight > topHeight && (rq.BlockHeight-topHeight) < int32(maxBlockHeightExcess) {
				confm.Confirmation = domains.UnableToVerify
			} else {
				confm.Confirmation = domains.Invalid
			}
		}

		mrcfs = append(mrcfs, confm)
	}

	return mrcfs, nil
}

// FillWithLongestChain fills the test header repository
// with 4 additional blocks to create a longest chain.
func (r *HeaderTestRepository) FillWithLongestChain() {
	db, _ := fixtures.AddLongestChain(*r.db)
	var filledDb []domains.BlockHeader = db
	r.db = &filledDb
}

func contains(hashes []string, hash string) bool {
	for _, h := range hashes {
		if h == hash {
			return true
		}
	}
	return false
}

func findHeader(hash string, headers []domains.BlockHeader) *domains.BlockHeader {
	for _, header := range headers {
		if header.Hash.String() == hash {
			return &header
		}
	}
	return nil
}

// NewHeadersTestRepository constructor for HeaderTestRepository.
func NewHeadersTestRepository(db *[]domains.BlockHeader) *HeaderTestRepository {
	return &HeaderTestRepository{
		db: db,
	}
}
