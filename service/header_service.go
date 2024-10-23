package service

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/bitcoin-sv/block-headers-service/bhserrors"
	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/bitcoin-sv/block-headers-service/repository"
	"github.com/rs/zerolog"
)

// HeaderService represents Header service and provide access to repositories.
type HeaderService struct {
	repo        *repository.Repositories
	checkpoints []chaincfg.Checkpoint
	timeSource  config.MedianTimeSource
	log         *zerolog.Logger
}

// NewHeaderService creates and returns HeaderService instance.
func NewHeaderService(repo *repository.Repositories, _ *config.P2PConfig, log *zerolog.Logger) *HeaderService {
	headerLogger := log.With().Str("service", "header").Logger()
	return &HeaderService{
		repo:        repo,
		checkpoints: config.Checkpoints,
		timeSource:  config.TimeSource,
		log:         &headerLogger,
	}
}

// AddHeader used to pass BlockHeader to repository which will add it to db.
func (hs *HeaderService) AddHeader(h domains.BlockHeader, _ int) error {
	return hs.repo.Headers.AddHeaderToDatabase(h)
}

// FindPreviousHeader returns previous header for the header with given hash.
func (hs *HeaderService) FindPreviousHeader(headerHash string) *domains.BlockHeader {
	h, err := hs.repo.Headers.GetPreviousHeader(headerHash)
	if err != nil {
		hs.log.Error().Msg(err.Error())
		return nil
	}
	return h
}

// BackElement returns last element from db (tip).
func (hs *HeaderService) BackElement() (domains.BlockHeader, error) {
	header, err := hs.repo.Headers.GetTip()
	if header == nil {
		return domains.BlockHeader{}, err
	}
	return *header, err
}

// IsCurrent checks if the headers are synchronized and up to date.
func (hs *HeaderService) IsCurrent() bool {
	// Not current if the latest main (best) chain height is before the
	// latest known good checkpoint (when checkpoints are enabled).
	checkpoints := hs.checkpoints
	checkpoint := &checkpoints[len(checkpoints)-1]
	tip := hs.GetTip()
	if tip == nil {
		return true
	}
	if checkpoint != nil && tip.Height < checkpoint.Height {
		return false
	}

	// Not current if the latest best block has a timestamp before 24 hours
	// ago.
	//
	// The chain appears to be current if none of the checks reported
	// otherwise.
	minus24Hours := hs.timeSource.AdjustedTime().Add(-24 * time.Hour).Unix()
	return tip.Timestamp.Unix() >= minus24Hours
}

// GetTip returns header which is the tip of the chain.
func (hs *HeaderService) GetTip() *domains.BlockHeader {
	tip, err := hs.repo.Headers.GetTip()
	if err != nil {
		return nil
	}
	return tip
}

// GetTipHeight returns height of the tip.
func (hs *HeaderService) GetTipHeight() int32 {
	tip := hs.GetTip()
	if tip != nil {
		return tip.Height
	}
	return 0
}

// GetHeaderByHash returns header with given hash.
func (hs *HeaderService) GetHeaderByHash(hash string) (*domains.BlockHeader, error) {
	header, err := hs.repo.Headers.GetHeaderByHash(hash)
	if err != nil {
		return nil, err
	}

	return header, nil
}

// GetHeadersByHeight returns the specified number of headers starting from given height.
func (hs *HeaderService) GetHeadersByHeight(height int, count int) ([]*domains.BlockHeader, error) {
	headersRange := height + count - 1
	headers, err := hs.repo.Headers.GetHeaderByHeightRange(height, headersRange)

	if err == nil {
		return headers, nil
	}
	return nil, err
}

// GetHeaderAncestorsByHash returns first ancestor for two headers specified by hash.
func (hs *HeaderService) GetHeaderAncestorsByHash(hash string, ancestorHash string) ([]*domains.BlockHeader, error) {
	// Get headers by hash
	reqHeader, err := hs.repo.Headers.GetHeaderByHash(hash)
	ancestorHeader, err2 := hs.repo.Headers.GetHeaderByHash(ancestorHash)

	// Check possible errors
	if err != nil || err2 != nil {
		return nil, bhserrors.ErrHeaderWithGivenHashes
	} else if ancestorHeader.Height > reqHeader.Height {
		return nil, bhserrors.ErrAncestorHashHigher
	} else if ancestorHeader.Height == reqHeader.Height {
		return make([]*domains.BlockHeader, 0), nil
	}

	a, err := hs.repo.Headers.GetAncestorOnHeight(reqHeader.Hash.String(), ancestorHeader.Height)
	if err != nil {
		return nil, bhserrors.ErrHeadersNotPartOfTheSameChain.Wrap(err)
	}

	if a.Hash != ancestorHeader.Hash {
		return nil, bhserrors.ErrHeadersNotPartOfTheSameChain
	}

	// Get headers from db
	headers, err := hs.repo.Headers.GetChainBetweenTwoHashes(ancestorHash, hash)

	if err == nil {
		return headers, nil
	}
	return nil, err
}

// GetCommonAncestor returns first ancestor for given slice of hashes.
func (hs *HeaderService) GetCommonAncestor(hashes []string) (*domains.BlockHeader, error) {
	headers := make([]*domains.BlockHeader, 0, len(hashes)+1)
	height := int32(math.MaxInt32)

	for _, hash := range hashes {
		header, err := hs.repo.Headers.GetHeaderByHash(hash)
		if err != nil {
			return nil, err
		}

		headers = append(headers, header)
		if header.Height < height {
			height = header.Height
		}
	}

	if height < 1 {
		return nil, nil
	}
	height--

	for i, h := range headers {
		a, err := hs.repo.Headers.GetAncestorOnHeight(h.Hash.String(), height)
		if err != nil {
			return nil, err
		}
		headers[i] = a
	}

	for height >= 0 {
		if areAllElementsEqual(headers) {
			return headers[0], nil
		}
		for i := range headers {
			h, err := hs.repo.Headers.GetPreviousHeader(headers[i].Hash.String())
			if err != nil {
				return nil, err
			}
			headers[i] = h
		}
		height--
	}

	return nil, nil
}

// GetHeadersState returns state of the header with given hash.
func (hs *HeaderService) GetHeadersState(hash string) (*domains.BlockHeaderState, error) {
	header, err := hs.repo.Headers.GetHeaderByHash(hash)
	if err != nil {
		return nil, err
	}

	state := domains.BlockHeaderState{
		Header:    *header,
		State:     header.State.String(),
		Height:    header.Height,
		ChainWork: header.Chainwork,
	}
	return &state, nil
}

// LatestHeaderLocator returns BlockLocator for current chain.
func (hs *HeaderService) LatestHeaderLocator() domains.BlockLocator {
	tip := hs.GetTip()
	if tip == nil {
		return nil
	}

	// Calculate the max number of entries that will ultimately be in the
	// block locator.  See the description of the algorithm for how these
	// numbers are derived.
	var maxEntries uint8
	if tip.Height <= 12 {
		maxEntries = uint8(tip.Height) + 1
	} else {
		// Requested hash itself + previous 10 entries + genesis block.
		// Then floor(log2(height-10)) entries for the skip portion.
		adjustedHeight := uint32(tip.Height) - 10
		maxEntries = 12 + domains.FastLog2Floor(adjustedHeight)
	}
	locator := make(domains.BlockLocator, 0, maxEntries)

	step := int32(1)
	for tip != (&domains.BlockHeader{}) {
		locator = append(locator, &tip.Hash)

		// Nothing more to add once the genesis block has been added.
		if tip.Height == 0 {
			break
		}

		// Calculate height of previous node to include ensuring the
		// final node is the genesis block.
		height := tip.Height - step
		if height < 0 {
			height = 0
		}

		v, _ := hs.repo.Headers.GetHeaderByHeight(height)
		if v == nil {
			return locator
		}

		tip = v
		// Once 11 entries have been included, start doubling the
		// distance between included hashes.
		if len(locator) > 10 {
			step *= 2
		}
	}

	return locator
}

// GetHeightByHash calculates height by hash.
func (hs *HeaderService) GetHeightByHash(hash *chainhash.Hash) (int32, error) {
	bh, err := hs.repo.Headers.GetHeaderByHash(hash.String())
	if err != nil {
		str := fmt.Sprintf("block %s is not in the main chain", hash)
		return 0, errors.New(str)
	}

	return bh.Height, nil
}

// LocateHeadersGetHeaders returns headers with given hashes.
func (hs *HeaderService) LocateHeadersGetHeaders(locators []*chainhash.Hash, hashstop *chainhash.Hash) ([]*wire.BlockHeader, error) {
	headers, err := hs.locateHeadersGetHeaders(locators, hashstop)
	if err != nil {
		return nil, err
	}
	return headers, nil
}

func (hs *HeaderService) locateHeadersGetHeaders(locators []*chainhash.Hash, hashstop *chainhash.Hash) ([]*wire.BlockHeader, error) {

	if len(locators) == 0 {
		return nil, errors.New("no locators provided")
	}

	hashes := make([]string, len(locators))
	for i, v := range locators {
		hashes[i] = v.String()
	}

	startHeight, err := hs.repo.Headers.GetHeadersStartHeight(hashes)
	if err != nil {
		return nil, fmt.Errorf("error getting headers of locators: %v", err)
	}
	var stopHeight int
	if hashstop.IsEqual(&chainhash.Hash{}) {
		stopHeight = startHeight + wire.MaxCFHeadersPerMsg
	} else {
		stopHeight, err = hs.repo.Headers.GetHeadersStopHeight(hashstop.String())
		if err != nil {
			return nil, fmt.Errorf("error getting hashstop height: %v", err)
		}
	}

	if stopHeight == 0 {
		stopHeight = startHeight + wire.MaxCFHeadersPerMsg
	}

	if stopHeight <= startHeight {
		return nil, errors.New("hashStop is lower than first valid height")
	}

	// Check if peer requested number of headers is higher than the maximum number of headers per message
	if wire.MaxCFHeadersPerMsg < stopHeight-startHeight {
		stopHeight = startHeight + wire.MaxCFHeadersPerMsg
	}

	dbHeaders, err := hs.repo.Headers.GetHeadersByHeightRange(startHeight+1, stopHeight)
	if err != nil {
		return nil, fmt.Errorf("error getting headers between heights: %v", err)
	}

	headers := make([]*wire.BlockHeader, 0, len(dbHeaders))
	for _, dbHeader := range dbHeaders {
		header := &wire.BlockHeader{
			Version:    dbHeader.Version,
			PrevBlock:  dbHeader.PreviousBlock,
			MerkleRoot: dbHeader.MerkleRoot,
			Timestamp:  dbHeader.Timestamp,
			Bits:       dbHeader.Bits,
			Nonce:      dbHeader.Nonce,
		}
		headers = append(headers, header)
	}

	return headers, nil
}

// LocateHeaders fetches headers for a number of blocks after the most recent known block
// in the best chain, based on the provided block locator and stop hash, and defaults to the
// genesis block if the locator is unknown.
func (hs *HeaderService) LocateHeaders(locator domains.BlockLocator, hashStop *chainhash.Hash) []wire.BlockHeader {
	headers, err := hs.locateHeadersGetHeaders(locator, hashStop)
	if err != nil {
		hs.log.Error().Msg(err.Error())
		return nil
	}

	result := make([]wire.BlockHeader, 0, len(headers))
	for _, header := range headers {
		result = append(result, *header)
	}

	return result
}

// Contains checks if given header is stored in db.
func (hs *HeaderService) Contains(node *domains.BlockHeader) bool {
	return hs.nodeByHeight(node.Height) == node
}

func (hs *HeaderService) nodeByHeight(height int32) *domains.BlockHeader {
	if height < 0 || height >= int32(hs.HeadersCount()) {
		return nil
	}

	header, err := hs.repo.Headers.GetHeaderByHeight(height)
	if err != nil {
		return nil
	}

	return header
}

// HeadersCount return current number of stored headers.
func (hs *HeaderService) HeadersCount() int {
	count, err := hs.repo.Headers.GetHeadersCount()
	if err != nil {
		return 0
	}

	return count
}

// Next returns next header for the given one.
func (hs *HeaderService) Next(node *domains.BlockHeader) *domains.BlockHeader {
	if node == nil || !hs.Contains(node) {
		return nil
	}

	return hs.nodeByHeight(node.Height + 1)
}

// CountHeaders return current number of stored headers.
func (hs *HeaderService) CountHeaders() int {
	count, err := hs.repo.Headers.GetHeadersCount()
	if err != nil {
		return 0
	}
	return count
}

// GetTips returns slice with current tips.
func (hs *HeaderService) GetTips() ([]*domains.BlockHeader, error) {
	return hs.repo.Headers.GetAllTips()
}

func areAllElementsEqual(slice []*domains.BlockHeader) bool {
	for _, val := range slice {
		if val.Hash != slice[0].Hash {
			return false
		}
	}
	return true
}
