package service

import (
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/repository"
)

type TipService struct {
	repo *repository.Repositories
}

func (s *TipService) GetTips() ([]domains.BlockHeaderState, error) {
	blockHeaders := make([]domains.BlockHeaderState, 0)
	tips := s.GetAllTips()
	for _, tip := range tips {
		v := tip.WrapWithHeaderState()
		blockHeaders = append(blockHeaders, v)
	}
	return blockHeaders, nil
}

func (s *TipService) PruneTip() (string, error) {
	return "", nil
}

func NewTipService(repo *repository.Repositories) *TipService {
	return &TipService{
		repo: repo,
	}
}

func (ts *TipService) GetAllTips() []domains.BlockHeader {
	var headersSlice []domains.BlockHeader

	// This method needs to be rewritten

	// if ts.headerChain.Forks == nil || len(ts.headerChain.Forks) == 0 {
	// 	lh, _ := ts.repo.Tips.GetConfirmedTip()
	// 	return []domains.Header{domains.MapHeaderSvFacadeToHeader(*lh)}
	// }

	// tipIndex := int(ts.headerChain.Tip.Height) + 1
	// forkMap, ok := ts.headerChain.Forks[tipIndex]
	// tips := forkMap

	// for ok {
	// 	for hash, header := range forkMap {
	// 		_, ok2 := tips[header.PrevBlock.String()]

	// 		if ok2 {
	// 			delete(tips, header.PrevBlock.String())
	// 		}
	// 		tips[hash] = header
	// 	}
	// 	tipIndex++
	// 	forkMap, ok = ts.headerChain.Forks[tipIndex]
	// }

	// for _, header := range tips {
	// 	headersSlice = append(headersSlice, *header)
	// }

	return headersSlice
}
