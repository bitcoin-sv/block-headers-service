package repository

import (
	"github.com/libsv/bitcoin-hc/data/sql"
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
)

// Headers is a interface which represents methods performed on defined storage.
type Headers interface {
	AddHeaderToDatabase(header domains.BlockHeader) error
	UpdateState([]chainhash.Hash, domains.HeaderState) error
	GetHeaderByHeight(height int32) (*domains.BlockHeader, error)
	GetHeaderByHeightRange(from int, to int) ([]*domains.BlockHeader, error)
	GetLongestChainHeadersFromHeight(height int32) ([]*domains.BlockHeader, error)
	GetStaleChainHeadersBackFrom(hash string) ([]*domains.BlockHeader, error)
	GetCurrentHeight() (int, error)
	GetHeadersCount() (int, error)
	GetHeaderByHash(hash string) (*domains.BlockHeader, error)
	GenesisExists() bool
	GetPreviousHeader(hash string) (*domains.BlockHeader, error)
	GetTip() (*domains.BlockHeader, error)
	GetAllTips() ([]*domains.BlockHeader, error)
	GetAncestorOnHeight(hash string, height int32) (*domains.BlockHeader, error)
	GetChainBetweenTwoHashes(low string, high string) ([]*domains.BlockHeader, error)
}

// Repositories represents all repositories in app and provide access to them.
type Repositories struct {
	Headers Headers
}
// NewRepositories creates and returns Repositories instance.
func NewRepositories(db *sql.HeadersDb) *Repositories {
	return &Repositories{
		Headers: NewHeadersRepository(db),
	}
}
