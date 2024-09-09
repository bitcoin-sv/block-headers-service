package repository

import (
	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/notification"
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
	GetMerkleRoots(batchSize int, lastEvaluatedKey int) (*domains.MerkleRootsESKPagedResponse, error)
	GenesisExists() bool
	GetPreviousHeader(hash string) (*domains.BlockHeader, error)
	GetTip() (*domains.BlockHeader, error)
	GetAllTips() ([]*domains.BlockHeader, error)
	GetAncestorOnHeight(hash string, height int32) (*domains.BlockHeader, error)
	GetChainBetweenTwoHashes(low string, high string) ([]*domains.BlockHeader, error)
	GetHeadersStartHeight(hashtable []string) (int, error)
	GetHeadersByHeightRange(from int, to int) ([]*domains.BlockHeader, error)
	GetHeadersStopHeight(hashStop string) (int, error)
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
