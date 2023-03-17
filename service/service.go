package service

import (
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
	"github.com/libsv/bitcoin-hc/internal/wire"
	"github.com/libsv/bitcoin-hc/repository"
	peerpkg "github.com/libsv/bitcoin-hc/transports/p2p/peer"
)

// Network is a interface which represents methods required for Network service.
type Network interface {
	GetPeers() []peerpkg.PeerState
	GetPeersCount() int
}


// Headers is a interface which represents methods required for Headers service.
type Headers interface {
	AddHeader(h domains.BlockHeader, blocksToConfirmFork int) error
	FindPreviousHeader(headerHash string) *domains.BlockHeader
	BackElement() (domains.BlockHeader, error)
	LatestHeaderLocator() domains.BlockLocator
	IsCurrent() bool
	BlockHeightByHash(hash *chainhash.Hash) (int32, error)
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
}

// Tip is a interface which represents methods required for Tip service.
type Tip interface {
	GetTips() ([]domains.BlockHeaderState, error)
	PruneTip() (string, error)
	GetAllTips() []domains.BlockHeader
}

// Services represents all services in app and provide access to them.
type Services struct {
	Network Network
	Headers Headers
	Tip     Tip
}

// Dept is a struct used to create Services. 
type Dept struct {
	Peers        map[*peerpkg.Peer]*peerpkg.PeerSyncState
	Repositories *repository.Repositories
}

// NewServices creates and returns Services instance.
func NewServices(d Dept) *Services {
	return &Services{
		Network: NewNetworkService(d.Peers),
		Headers: NewHeaderService(d.Repositories),
		Tip:     NewTipService(d.Repositories),
	}

}
