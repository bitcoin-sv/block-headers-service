package repository

import (
	"github.com/bitcoin-sv/pulse/database/sql"
	"github.com/bitcoin-sv/pulse/domains"
	"github.com/bitcoin-sv/pulse/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/pulse/notification"
)

// Headers is a interface which represents methods performed on header table in defined storage.
type Headers interface {
	AddHeaderToDatabase(domains.BlockHeader) error
	AddMultipleHeadersToDatabase([]domains.BlockHeader) error
	UpdateState([]chainhash.Hash, domains.HeaderState) error
	GetHeaderByHeight(height int32) (*domains.BlockHeader, error)
	GetHeaderByHeightRange(from int, to int) ([]*domains.BlockHeader, error)
	GetLongestChainHeadersFromHeight(height int32) ([]*domains.BlockHeader, error)
	GetStaleChainHeadersBackFrom(hash string) ([]*domains.BlockHeader, error)
	GetCurrentHeight() (int, error)
	GetHeadersCount() (int, error)
	GetHeaderByHash(hash string) (*domains.BlockHeader, error)
	GetMerkleRootsConfirmations(request []domains.MerkleRootConfirmationRequestItem, maxBlockHeightExcess int) ([]*domains.MerkleRootConfirmation, error)
	GenesisExists() bool
	GetPreviousHeader(hash string) (*domains.BlockHeader, error)
	GetTip() (*domains.BlockHeader, error)
	GetAllTips() ([]*domains.BlockHeader, error)
	GetAncestorOnHeight(hash string, height int32) (*domains.BlockHeader, error)
	GetChainBetweenTwoHashes(low string, high string) ([]*domains.BlockHeader, error)
}

// Tokens is a interface which represents methods performed on tokens table in defined storage.
type Tokens interface {
	AddTokenToDatabase(token *domains.Token) error
	GetTokenByValue(token string) (*domains.Token, error)
	DeleteToken(token string) error
}

// Repositories represents all repositories in app and provide access to them.
type Repositories struct {
	Headers  Headers
	Tokens   Tokens
	Webhooks notification.Webhooks
}

// NewRepositories creates and returns Repositories instance.
func NewRepositories(db *sql.HeadersDb) *Repositories {
	return &Repositories{
		Headers:  NewHeadersRepository(db),
		Tokens:   NewTokensRepository(db),
		Webhooks: NewWebhooksRepository(db),
	}
}
