package service

import (
	"github.com/bitcoin-sv/pulse/domains"
	customErrs "github.com/bitcoin-sv/pulse/errors"
	"github.com/bitcoin-sv/pulse/internal/chaincfg"
	"github.com/bitcoin-sv/pulse/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/pulse/metrics"
	"github.com/bitcoin-sv/pulse/repository"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// BlockHasher is an interface which is exposing BlockHash method.
type BlockHasher interface {
	// BlockHash calculates BlockHash from given header source BlockHeaderSource
	BlockHash(h *domains.BlockHeaderSource) domains.BlockHash
}

// Notification is "port" through which chain service can notify clients about important events.
type Notification interface {
	//Notify notifies about new header stored.
	Notify(any)
}

type chainService struct {
	*repository.Repositories
	chainParams  *chaincfg.Params
	log          *zerolog.Logger
	notification Notification
	BlockHasher
}

// ChainServiceDependencies is a configuration struct used to initialize a new Chains service.
type ChainServiceDependencies struct {
	*repository.Repositories
	*chaincfg.Params
	*zerolog.Logger
	Notification
	BlockHasher
}

// NewChainsService is a constructor for Chains service.
func NewChainsService(
	repos *repository.Repositories,
	params *chaincfg.Params,
	log *zerolog.Logger,
	hasher BlockHasher,
	notification Notification,
) Chains {
	serviceLogger := log.With().Str("service", "chain").Logger()
	return &chainService{
		Repositories: repos,
		chainParams:  params,
		log:          &serviceLogger,
		BlockHasher:  hasher,
		notification: notification,
	}
}

func (cs *chainService) Add(bs domains.BlockHeaderSource) (*domains.BlockHeader, error) {
	hash := cs.BlockHasher.BlockHash(&bs)

	if cs.ignoreBlockHash(&hash) {
		cs.log.Warn().Msgf("Message rejected - containing forbidden header")
		return domains.NewRejectedBlockHeader(hash), BlockRejected.error()
	}

	h, err := cs.createHeader(&hash, &bs)
	if err != nil {
		return nil, HeaderCreationFail.causedBy(&err)
	}

	isConcurrentChain := cs.hasConcurrentHeaderFromLongestChain(h)

	if isConcurrentChain {
		tip, err := cs.Repositories.Headers.GetTip()
		if err != nil {
			return nil, HeaderCreationFail.causedBy(&err)
		}

		if tip.CumulatedWork.Cmp(h.CumulatedWork) < 0 {
			h.State = domains.LongestChain
		} else {
			h.State = domains.Stale
		}
	}

	if isConcurrentChain && h.IsLongestChain() {
		err := cs.switchChainsStates(h)
		if err != nil {
			return h, err
		}
	}

	h, err = cs.insert(h)
	if err != nil {
		return nil, err
	}

	metrics.SetLatestBlock(h.Height, h.Timestamp, h.State.String())
	cs.notification.Notify(domains.HeaderAdded(h))
	return h, err
}

func (cs *chainService) hasConcurrentHeaderFromLongestChain(h *domains.BlockHeader) bool {
	if h.IsOrphan() {
		return false
	}
	if h.IsLongestChain() {
		oh, _ := cs.Headers.GetHeaderByHeight(h.Height)
		return oh != nil && oh.IsLongestChain() && !oh.Hash.IsEqual(&h.Hash)
	}
	return true
}

// switchChainsStates marking chain connected to given block as longest chain
// and concurrent part of (currently) "longest chain" as STALE.
func (cs *chainService) switchChainsStates(h *domains.BlockHeader) error {
	cs.log.Warn().Msgf("Promoting currently stale chain to be LONGEST chain ending on header %s", h.Hash)
	headerStaleChain, err := cs.stalePartOfChainOf(h)
	if err != nil {
		return ChainUpdateFail.causedBy(&err)
	}

	lh := lowestHeightOf(&headerStaleChain, h)

	concurrentChain, err := cs.longestChainFromHeight(lh)
	if err != nil {
		return ChainUpdateFail.causedBy(&err)
	}

	err = cs.Headers.UpdateState(concurrentChain.hashes(), domains.Stale)
	if err != nil {
		return ChainUpdateFail.causedBy(&err)
	}

	err = cs.Headers.UpdateState(headerStaleChain.hashes(), domains.LongestChain)
	if err != nil {
		return ChainUpdateFail.causedBy(&err)
	}
	return nil
}

func (cs *chainService) longestChainFromHeight(smallestHeight int32) (chain, error) {
	concurrentChain, err := cs.Headers.GetLongestChainHeadersFromHeight(smallestHeight)
	if err != nil {
		return concurrentChain, ChainUpdateFail.causedBy(&err)
	}
	return concurrentChain, nil
}

func (cs *chainService) stalePartOfChainOf(h *domains.BlockHeader) (chain, error) {
	headerStaleChain, err := cs.Headers.GetStaleChainHeadersBackFrom(h.PreviousBlock.String())
	if err != nil {
		return headerStaleChain, ChainUpdateFail.causedBy(&err)
	}
	return headerStaleChain, nil
}

func (cs *chainService) ignoreBlockHash(blockHash *domains.BlockHash) bool {
	bhash := chainhash.Hash(*blockHash)
	for _, hash := range cs.chainParams.HeadersToIgnore {
		if bhash.IsEqual(hash) {
			return true
		}
	}

	return false
}

func (cs *chainService) createHeader(hash *domains.BlockHash, bs *domains.BlockHeaderSource) (*domains.BlockHeader, error) {
	ph, err := cs.previousHeader(bs)
	if err != nil {
		return nil, err
	}
	bh := domains.CreateHeader(hash, bs, ph)
	return &bh, nil
}

func (cs *chainService) previousHeader(bs *domains.BlockHeaderSource) (*domains.BlockHeader, error) {
	h, err := cs.Repositories.Headers.GetHeaderByHash(bs.PrevBlock.String())
	if h == nil && err != nil && err.Error() == "could not find hash" {
		return domains.NewOrphanPreviousBlockHeader(), nil
	}
	return h, err
}

func (cs *chainService) insert(h *domains.BlockHeader) (*domains.BlockHeader, error) {
	err := cs.Repositories.Headers.AddHeaderToDatabase(*h)
	if err != nil {
		if errors.Is(err, customErrs.NewUniqueViolationError()) {
			cs.log.Warn().Msgf("Header %s already exists in the repository", h.Hash)
			return h, nil
		}

		return h, HeaderSaveFail.causedBy(&err)
	}
	return h, nil
}

type chain []*domains.BlockHeader

func lowestHeightOf(c *chain, oh *domains.BlockHeader) int32 {
	f := c.first()
	if f.Height < oh.Height {
		return f.Height
	}
	return oh.Height
}

func (c *chain) first() *domains.BlockHeader {
	hs := []*domains.BlockHeader(*c)
	if len(hs) == 0 {
		return nil
	}

	f := hs[0]
	for _, ch := range hs {
		if ch.Height < f.Height {
			f = ch
		}
	}
	return f
}

func (c *chain) hashes() []chainhash.Hash {
	hs := make([]chainhash.Hash, len(*c))
	for i, ch := range *c {
		hs[i] = ch.Hash
	}
	return hs
}

// AddBlockError errors that could occur during adding a header.
type AddBlockError struct {
	code  AddBlockErrorCode
	cause *error
}

// AddBlockErrorCode error codes that could occur during adding a header.
type AddBlockErrorCode string

const (
	//BlockRejected error code representing situation when block is on the blacklist.
	BlockRejected AddBlockErrorCode = "BlockRejected"

	//HeaderCreationFail error code representing situation when block cannot be created from source.
	HeaderCreationFail AddBlockErrorCode = "HeaderCreationFail"

	//ChainUpdateFail error code representing situation when STALE chain should become Longest chain but the update of chains failed.
	ChainUpdateFail AddBlockErrorCode = "ChainUpdateFail"

	//HeaderSaveFail error code representing situation when saving header in the repository failed.
	HeaderSaveFail AddBlockErrorCode = "HeaderSaveFail"
)

func (e *AddBlockError) Error() string {
	return string(e.code)
}

// Cause returns a Cause of the error if there is any.
func (e *AddBlockError) Cause() error {
	return *e.cause
}

func (c AddBlockErrorCode) String() string {
	return string(c)
}

func (c AddBlockErrorCode) error() error {
	return errors.New(c.String())
}

func (c AddBlockErrorCode) causedBy(cause *error) error {
	return errors.Wrap(*cause, c.String())
}

// Is checks if given error contains AddBlockErrorCode.
func (c AddBlockErrorCode) Is(err error) bool {
	return err != nil && err.Error() == c.String()
}
