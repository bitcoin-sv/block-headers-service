package service

import (
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
	"github.com/libsv/bitcoin-hc/repository"
	"github.com/libsv/bitcoin-hc/transports/p2p/p2plog"
)

type BlockHasher interface {
	BlockHash(h *BlockHeaderSource) domains.BlockHash
}

type chainService struct {
	*repository.Repositories
	chainParams *chaincfg.Params
	log         p2plog.Logger
	BlockHasher
}

// ChainServiceDependencies is a configuration struct used to initialize a new Chains sevice
type ChainServiceDependencies struct {
	*repository.Repositories
	*chaincfg.Params
	p2plog.Logger
	BlockHasher
}

func NewChainsService(deps ChainServiceDependencies) Chains {
	return &chainService{
		Repositories: deps.Repositories,
		chainParams:  deps.Params,
		log:          deps.Logger,
		BlockHasher:  deps.BlockHasher,
	}
}

func (cs *chainService) Add(bs BlockHeaderSource) (*domains.BlockHeader, error) {
	hash := cs.BlockHasher.BlockHash(&bs)

	if cs.ignoreBlockHash(&hash) {
		cs.log.Warnf("Message rejected - containing forbidden header")
		return domains.NewRejectedBlockHeader(hash), BlockRejected.error()
	}

	h := cs.createHeader(&hash, &bs)
	return cs.insert(&h)
}

func (cs *chainService) insert(h *domains.BlockHeader) (*domains.BlockHeader, error) {
	err := cs.Repositories.Headers.AddHeaderToDatabase(*h)
	if err != nil {
		return h, BlockSaveFail.causedBy(&err)
	}
	return h, nil
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

func (cs *chainService) createHeader(hash *domains.BlockHash, bs *BlockHeaderSource) domains.BlockHeader {
	ph := cs.previousHeader(bs)

	cw := domains.CalculateWork(bs.Bits)
	ccw := domains.CumulatedChainWorkOf(*ph.CumulatedWork).Add(cw)

	var state domains.HeaderState
	if ph.IsOrphan() {
		cs.log.Infof("Header %s is considered an orphan", hash)
		state = domains.Orphan
	} else {
		state = domains.LongestChain
	}

	return domains.BlockHeader{
		Height:        ph.Height + 1,
		Hash:          hash.ChainHash(),
		Version:       bs.Version,
		MerkleRoot:    bs.MerkleRoot,
		Timestamp:     bs.Timestamp,
		Bits:          bs.Bits,
		Nonce:         bs.Nonce,
		State:         state,
		Chainwork:     cw.Uint64(),
		CumulatedWork: ccw.BigInt(),
		PreviousBlock: ph.PreviousBlock,
	}
}

func (cs *chainService) previousHeader(bs *BlockHeaderSource) *domains.BlockHeader {
	h, _ := cs.Repositories.Headers.GetHeaderByHash(bs.PrevBlock.String())
	if h == nil {
		return domains.NewOrphanPreviousBlockHeader()
	} else {
		return h
	}
}

type AddBlockError struct {
	code  AddBlockErrorCode
	cause *error
}

type AddBlockErrorCode string

const (
	BlockRejected AddBlockErrorCode = "BlockRejected"
	BlockSaveFail AddBlockErrorCode = "BlockSaveFail"
)

func (e *AddBlockError) Error() string {
	return string(e.code)
}

func (e *AddBlockError) Cause() error {
	return *e.cause
}

func (c AddBlockErrorCode) String() string {
	return string(c)
}

func (c AddBlockErrorCode) error() *AddBlockError {
	return &AddBlockError{
		code: c,
	}
}

func (c AddBlockErrorCode) causedBy(cause *error) *AddBlockError {
	return &AddBlockError{
		code:  c,
		cause: cause,
	}
}

func (c AddBlockErrorCode) Is(err error) bool {
	return err != nil && err.Error() == c.String()
}
