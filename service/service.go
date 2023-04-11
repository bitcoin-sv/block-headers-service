package service

import (
	"github.com/libsv/bitcoin-hc/configs"
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
	"github.com/libsv/bitcoin-hc/internal/wire"
	"github.com/libsv/bitcoin-hc/repository"
	peerpkg "github.com/libsv/bitcoin-hc/transports/p2p/peer"
)

// Network is an interface which represents methods required for Network service.
type Network interface {
	GetPeers() []peerpkg.PeerState
	GetPeersCount() int
}

// Headers is an interface which represents methods required for Headers service.
type Headers interface {
	FindPreviousHeader(headerHash string) *domains.BlockHeader
	LatestHeaderLocator() domains.BlockLocator
	IsCurrent() bool
	GetHeightByHash(hash *chainhash.Hash) (int32, error)
	LocateHeaders(locator domains.BlockLocator, hashStop *chainhash.Hash) []wire.BlockHeader
	GetTip() *domains.BlockHeader
	GetTipHeight() int32
	CountHeaders() int
	InsertGenesisHeaderInDatabase() error
	GetHeaderByHash(hash string) (*domains.BlockHeader, error)
	GetHeadersByHeight(height int, count int) ([]*domains.BlockHeader, error)
	GetHeaderAncestorsByHash(hash string, ancestorHash string) ([]*domains.BlockHeader, error)
	GetCommonAncestors(hashes []string) (*domains.BlockHeader, error)
	GetHeadersState(hash string) (*domains.BlockHeaderState, error)
	GetTips() ([]*domains.BlockHeader, error)
	GetPruneTip() (string, error)
	CalculateConfirmations(originHeader *domains.BlockHeader) int
}

// Chains is an interface which represents methods exposed by Chains Service.
type Chains interface {
	Add(domains.BlockHeaderSource) (*domains.BlockHeader, error)
}

// Services represents all services in app and provide access to them.
type Services struct {
	Network Network
	Headers Headers
	Chains  Chains
}

// Dept is a struct used to create Services.
type Dept struct {
	Peers        map[*peerpkg.Peer]*peerpkg.PeerSyncState
	Repositories *repository.Repositories
	Params       *chaincfg.Params
}

// NewServices creates and returns Services instance.
func NewServices(d Dept) *Services {
	return &Services{
		Network: NewNetworkService(d.Peers),
		Headers: NewHeaderService(d.Repositories),
		Chains: NewChainsService(ChainServiceDependencies{
			Repositories: d.Repositories,
			Params:       d.Params,
			Logger:       configs.Log,
			BlockHasher:  DefaultBlockHasher(),
		}),
	}
}
