package service

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/libsv/bitcoin-hc/configs"
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
	"github.com/libsv/bitcoin-hc/internal/wire"
	"github.com/libsv/bitcoin-hc/repository"
)

// HeaderService represents Header service and provide access to repositories.
type HeaderService struct {
	repo *repository.Repositories
}

// NewHeaderService creates and returns HeaderService instance.
func NewHeaderService(repo *repository.Repositories) *HeaderService {
	return &HeaderService{
		repo: repo,
	}
}

// AddHeader used to pass BlockHeader to repository which will add it to db.
func (hs *HeaderService) AddHeader(h domains.BlockHeader, blocksToConfirmFork int) error {
	return hs.repo.Headers.AddHeaderToDatabase(h)
}

// FindPreviousHeader returns previous header for the header with given hash.
func (hs *HeaderService) FindPreviousHeader(headerHash string) *domains.BlockHeader {
	h, err := hs.repo.Headers.GetPreviousHeader(headerHash)
	if err != nil {
		configs.Log.Error(err)
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

// LatestHeaderLocator returns BlockLocator for current chain.
func (hs *HeaderService) LatestHeaderLocator() domains.BlockLocator {
	return hs.blockLocator()
}

// IsCurrent checks if the headers are synchronized and up to date.
func (hs *HeaderService) IsCurrent() bool {
	// Not current if the latest main (best) chain height is before the
	// latest known good checkpoint (when checkpoints are enabled).
	checkpoints := configs.Cfg.Checkpoints
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
	minus24Hours := configs.Cfg.TimeSource.AdjustedTime().Add(-24 * time.Hour).Unix()
	return tip.Timestamp.Unix() >= minus24Hours
}

// BlockHeightByHash returns height of the header with given hash.
func (hs *HeaderService) BlockHeightByHash(hash *chainhash.Hash) (int32, error) {
	bh, err := hs.repo.Headers.GetBlockByHash(domains.HeaderArgs{Blockhash: hash.String()})
	if err != nil {
		str := fmt.Sprintf("block %s is not in the main chain", hash)
		return 0, errors.New(str)
	}

	return bh.Height, nil
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
	headers_range := height + count - 1
	headers, err := hs.repo.Headers.GetHeaderByHeightRange(height, headers_range)

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
	if err != nil && err2 != nil {
		return nil, errors.New("error during etting headers with given hashes")
	} else if ancestorHeader.Height > reqHeader.Height {
		return nil, errors.New("ancestor header height can not be higher than requested header heght")
	} else if ancestorHeader.Height == reqHeader.Height {
		return make([]*domains.BlockHeader, 0), nil
	}

	// Get headers from db
	headers, err := hs.repo.Headers.GetHeaderByHeightRange(int(ancestorHeader.Height), int(reqHeader.Height))

	if err == nil {
		return headers, nil
	}
	return nil, nil
}

// GetCommonAncestors returns first ancestor for given slice of hashes.
func (hs *HeaderService) GetCommonAncestors(hashes []string) (*domains.BlockHeader, error) {
	headers := make([]*domains.BlockHeader, 0)

	for _, hash := range hashes {
		header, err := hs.repo.Headers.GetHeaderByHash(hash)
		if err != nil {
			return nil, err
		}

		headers = append(headers, header)
	}

	// sort slice to get the lowest height
	sort.SliceStable(headers, func(i, j int) bool {
		return headers[i].Height < headers[j].Height
	})

	header, err := hs.repo.Headers.GetHeaderByHeight(headers[0].Height - 1)
	if err == nil {
		return header, nil
	}
	return nil, err
}

// GetHeadersState returns state of the header with given hash.
func (hs *HeaderService) GetHeadersState(hash string) (*domains.BlockHeaderState, error) {
	header, err := hs.repo.Headers.GetHeaderByHash(hash)

	if err != nil {
		return nil, err
	}

	state := domains.BlockHeaderState{
		Header:        *header,
		State:         header.State.String(),
		Height:        header.Height,
		ChainWork:     header.Chainwork,
		Confirmations: hs.CalculateConfirmations(header),
	}
	return &state, nil
}

func (hs *HeaderService) blockLocator() domains.BlockLocator {
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


// LocateHeaders fetches headers for a number of blocks after the most recent known block
// in the best chain, based on the provided block locator and stop hash, and defaults to the
// genesis block if the locator is unknown.
func (hs *HeaderService) LocateHeaders(locator domains.BlockLocator, hashStop *chainhash.Hash) []wire.BlockHeader {
	headers := hs.locateHeaders(locator, hashStop, wire.MaxBlockHeadersPerMsg)
	return headers
}

func (hs *HeaderService) locateHeaders(locator domains.BlockLocator, hashStop *chainhash.Hash, maxHeaders uint32) []wire.BlockHeader {
	// Find the node after the first known block in the locator and the
	// total number of nodes after it needed while respecting the stop hash
	// and max entries.
	node, total := hs.locateInventory(locator, hashStop, maxHeaders)
	if total == 0 {
		return nil
	}

	// Populate and return the found headers.
	headers := make([]wire.BlockHeader, 0, total)
	for i := uint32(0); i < total; i++ {
		header := wire.BlockHeader{
			Version:    node.Version,
			PrevBlock:  node.PreviousBlock,
			MerkleRoot: node.MerkleRoot,
			Timestamp:  node.Timestamp,
			Bits:       node.Bits,
			Nonce:      node.Nonce,
		}
		headers = append(headers, header)
		node = hs.nodeByHeight(node.Height + 1)
	}
	return headers
}

func (hs *HeaderService) locateInventory(locator domains.BlockLocator, hashStop *chainhash.Hash, maxEntries uint32) (*domains.BlockHeader, uint32) {
	// There are no block locators so a specific block is being requested
	// as identified by the stop hash.
	stopNode := hs.LookupNode(hashStop)
	if len(locator) == 0 {
		if stopNode == nil {
			// No blocks with the stop hash were found so there is
			// nothing to do.
			return nil, 0
		}
		return stopNode, 1
	}

	// Find the most recent locator block hash in the main chain.  In the
	// case none of the hashes in the locator are in the main chain, fall
	// back to the genesis block.
	startNode, _ := hs.repo.Headers.GetHeaderByHeight(0)
	for _, hash := range locator {
		node := hs.LookupNode(hash)
		if node != nil && hs.Contains(node) {
			startNode = node
			break
		}
	}

	// Start at the block after the most recently known block.  When there
	// is no next block it means the most recently known block is the tip of
	// the best chain, so there is nothing more to do.
	next := hs.Next(startNode)
	if next == nil {
		return nil, 0
	}
	startNode = next

	// Calculate how many entries are needed.
	total := uint32((hs.GetTipHeight() - startNode.Height) + 1)
	if stopNode != nil && hs.Contains(stopNode) &&
		stopNode.Height >= startNode.Height {

		total = uint32((stopNode.Height - startNode.Height) + 1)
	}
	if total > maxEntries {
		total = maxEntries
	}

	return startNode, total
}

// LookupNode return header by given Hash.
func (hs *HeaderService) LookupNode(hash *chainhash.Hash) *domains.BlockHeader {
	node, err := hs.repo.Headers.GetBlockByHash(domains.HeaderArgs{Blockhash: hash.String()})
	if err != nil {
		return nil
	}

	return node
}

// Contains checks if given header is stored in db.
func (hs *HeaderService) Contains(node *domains.BlockHeader) bool {
	contains := hs.contains(node)
	return contains
}

func (hs *HeaderService) contains(node *domains.BlockHeader) bool {
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
	if node == nil || !hs.contains(node) {
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

// InsertGenesisHeaderInDatabase adds a genesis header (height=0) to db.
func (hs *HeaderService) InsertGenesisHeaderInDatabase() error {
	genesisBlock := domains.CreateGenesisHeaderBlock()
	if hs.repo.Headers.GenesisExists() {
		return nil
	}

	err := hs.repo.Headers.AddHeaderToDatabase(genesisBlock)

	return err
}

// CalculateConfirmations returns number of confirmations for given header.
func (hs *HeaderService) CalculateConfirmations(originHeader *domains.BlockHeader) int {
	conf, err := hs.repo.Headers.
		GetConfirmationsCountForBlock(originHeader.Hash.String())
	if err != nil {
		configs.Log.Errorf("%v", err.Error())
		return conf
	}

	return conf
}
