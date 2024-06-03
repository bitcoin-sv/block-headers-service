package service

import (
	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/bitcoin-sv/block-headers-service/notification"
	"github.com/bitcoin-sv/block-headers-service/repository"
	"github.com/bitcoin-sv/block-headers-service/transports/http/client"
	peerpkg "github.com/bitcoin-sv/block-headers-service/transports/p2p/peer"
	"github.com/rs/zerolog"
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
	GetHeaderByHash(hash string) (*domains.BlockHeader, error)
	GetHeadersByHeight(height int, count int) ([]*domains.BlockHeader, error)
	GetMerkleRootsConfirmations(request []domains.MerkleRootConfirmationRequestItem) ([]*domains.MerkleRootConfirmation, error)
	GetHeaderAncestorsByHash(hash string, ancestorHash string) ([]*domains.BlockHeader, error)
	GetCommonAncestor(hashes []string) (*domains.BlockHeader, error)
	GetHeadersState(hash string) (*domains.BlockHeaderState, error)
	GetTips() ([]*domains.BlockHeader, error)
	LocateHeadersGetHeaders(locators []*chainhash.Hash, hashstop *chainhash.Hash) ([]*wire.BlockHeader, error)
}

// Chains is an interface which represents methods exposed by Chains Service.
type Chains interface {
	Add(domains.BlockHeaderSource) (*domains.BlockHeader, error)
}

// Tokens is an interface which represents methods required for Tokens service.
type Tokens interface {
	GenerateToken() (*domains.Token, error)
	GetToken(token string) (*domains.Token, error)
	DeleteToken(token string) error
}

// Services represents all services in app and provide access to them.
type Services struct {
	Network  Network
	Headers  Headers
	Chains   Chains
	Tokens   Tokens
	Notifier *notification.Notifier
	Webhooks *notification.WebhooksService
}

// Dept is a struct used to create Services.
type Dept struct {
	Peers        map[*peerpkg.Peer]*peerpkg.PeerSyncState
	Repositories *repository.Repositories
	AdminToken   string
	Logger       *zerolog.Logger
	Config       *config.AppConfig
}

// NewServices creates and returns Services instance.
func NewServices(d Dept) *Services {
	notifier := newNotifier()

	return &Services{
		Network:  NewNetworkService(d.Peers),
		Headers:  NewHeaderService(d.Repositories, d.Config.P2P, d.Config.MerkleRoot, d.Logger),
		Notifier: notifier,
		Chains:   newChainService(d, notifier),
		Tokens:   NewTokenService(d.Repositories, d.AdminToken),
		Webhooks: newWebhooks(d),
	}
}

func newChainService(d Dept, notifier *notification.Notifier) Chains {
	return NewChainsService(
		d.Repositories,
		d.Config.P2P.GetNetParams(),
		d.Logger,
		DefaultBlockHasher(),
		notifier,
	)
}

func newWebhooks(d Dept) *notification.WebhooksService {
	return notification.NewWebhooksService(
		d.Repositories.Webhooks,
		client.NewWebhookTargetClient(),
		d.Logger,
		d.Config.Webhook,
	)
}

func newNotifier() *notification.Notifier {
	return notification.NewNotifier()
}
