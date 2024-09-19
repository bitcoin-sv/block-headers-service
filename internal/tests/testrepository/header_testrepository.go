package testrepository

import (
	"errors"
	"slices"
	"sort"

	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/fixtures"
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

// AddMultipleHeadersToDatabase adds multiple new headers to db.
func (r *HeaderTestRepository) AddMultipleHeadersToDatabase(headers []domains.BlockHeader) error {
	*r.db = append(*r.db, headers...)
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
func (r *HeaderTestRepository) GetAncestorOnHeight(_ string, height int32) (*domains.BlockHeader, error) {
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

// GetMerkleRoots returns ExclusiveStartKey pagination of batchSize size with merkle roots from lastEvaluatedKey which
// is the last merkleroot of the block that a client has processed
func (r *HeaderTestRepository) GetMerkleRoots(batchSize int, lastEvaluatedKey string) (*domains.MerkleRootsESKPagedResponse, error) {
	// Order headers by height ASC
	sort.Slice(*r.db, func(i, j int) bool {
		return (*r.db)[i].Height < (*r.db)[j].Height
	})

	// Check if lastEvaluatedKey is the same as the last element's MerkleRoot
	if lastEvaluatedKey != "" && (*r.db)[len(*r.db)-1].MerkleRoot.String() == lastEvaluatedKey {
		// Return empty content since we have reached the end
		return &domains.MerkleRootsESKPagedResponse{
			Page: domains.ExclusiveStartKeyPageInfo{
				TotalElements:    int32(len(*r.db)),
				Size:             0,
				LastEvaluatedKey: "",
			},
			Content: []domains.MerkleRootsResponse{},
		}, nil
	}

	// Find the starting index based on the lastEvaluatedKey
	startIdx := slices.IndexFunc(*r.db, func(c domains.BlockHeader) bool { return c.MerkleRoot.String() == lastEvaluatedKey })

	if startIdx == -1 && lastEvaluatedKey != "" {
		return nil, domains.ErrMerklerootNotFound
	}

	// Check if lastEvaluatedKey is not from the longest chain
	if startIdx != -1 && !(*r.db)[startIdx].IsLongestChain() {
		return nil, domains.ErrMerklerootNotInLongestChain
	}

	// If the lastEvaluatedKey is found, we start after it; otherwise, we start from the beginning
	if startIdx != -1 {
		startIdx++ // Start from the next element after the found key
	} else {
		startIdx = 0 // Start from the beginning if no key is provided
	}

	// Filter out headers with "STALE" state
	filteredHeaders := make([]domains.BlockHeader, 0)
	for _, header := range (*r.db)[startIdx:] {
		if header.State != "STALE" {
			filteredHeaders = append(filteredHeaders, header)
		}
	}

	// Calculate the end index for the batch based on batchSize
	endIdx := batchSize
	if endIdx > len(filteredHeaders) {
		endIdx = len(filteredHeaders) // Limit to the size of the db if the batch size exceeds available elements
	}

	merkleroots := filteredHeaders[:endIdx]

	// Determine the newLastEvaluatedKey
	newLastEvaluatedKey := ""

	if len(merkleroots) > 0 {
		newLastEvaluatedKey = merkleroots[len(merkleroots)-1].MerkleRoot.String()

		// If the newLastEvaluatedKey is equal to the tip merkleroos we have no more blocks in database
		if (*r.db)[len(*r.db)-1].MerkleRoot.String() == newLastEvaluatedKey {
			newLastEvaluatedKey = ""
		}
	}

	merkleRootsESKPagedResponse := &domains.MerkleRootsESKPagedResponse{
		Page: domains.ExclusiveStartKeyPageInfo{
			TotalElements:    int32(len(*r.db)),
			Size:             len(merkleroots),
			LastEvaluatedKey: newLastEvaluatedKey,
		},
		Content: make([]domains.MerkleRootsResponse, len(merkleroots)),
	}

	for i, mkr := range merkleroots {
		merkleRootsESKPagedResponse.Content[i] = domains.MerkleRootsResponse{
			MerkleRoot:  mkr.MerkleRoot.String(),
			BlockHeight: mkr.Height,
		}
	}

	return merkleRootsESKPagedResponse, nil
}

// GetMerkleRootsConfirmations returns a confirmation of merkle roots inclusion
// in the longest chain with hash of the block in which the merkle root is included.
func (r *HeaderTestRepository) GetMerkleRootsConfirmations(
	request []domains.MerkleRootConfirmationRequestItem,
	maxBlockHeightExcess int,
) ([]*domains.MerkleRootConfirmation, error) {
	mrcfs := make([]*domains.MerkleRootConfirmation, 0)

	topHeight := int32(0)
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

// GetHeadersStartHeight returns height of the highest header from the list of hashes.
func (r *HeaderTestRepository) GetHeadersStartHeight(hashtable []string) (int, error) {
	for i := len(*r.db) - 1; i >= 0; i-- {
		header := (*r.db)[i]
		for j := len(hashtable) - 1; j >= 0; j-- {
			if header.Hash.String() == hashtable[j] && header.State == domains.LongestChain {
				return int(header.Height), nil
			}
		}
	}
	return 0, nil
}

// GetHeadersByHeightRange returns headers from db in specified height range.
func (r *HeaderTestRepository) GetHeadersByHeightRange(from int, to int) ([]*domains.BlockHeader, error) {
	filteredHeaders := make([]*domains.BlockHeader, 0)
	for _, header := range *r.db {
		if header.Height >= int32(from) && header.Height <= int32(to) {
			headerCopy := header
			filteredHeaders = append(filteredHeaders, &headerCopy)
		}
	}
	return filteredHeaders, nil
}

// GetHeadersStopHeight returns height of hashstop header from db.
func (r *HeaderTestRepository) GetHeadersStopHeight(hashStop string) (int, error) {
	for i := len(*r.db) - 1; i >= 0; i-- {
		header := (*r.db)[i]
		if header.Hash.String() == hashStop {
			return int(header.Height), nil
		}
	}
	return 0, errors.New("could not find stop height")
}

// FillWithLongestChain fills the test header repository
// with 4 additional blocks to create a longest chain.
func (r *HeaderTestRepository) FillWithLongestChain() {
	db, _ := fixtures.AddLongestChain(*r.db)
	var filledDb []domains.BlockHeader = db
	r.db = &filledDb
}

// FillWithLongestChainWithFork fills the test header repository
// with 4 additional blocks to create a longest chain.
func (r *HeaderTestRepository) FillWithLongestChainWithFork() {
	db, _ := fixtures.LongestChainWithFork()
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
