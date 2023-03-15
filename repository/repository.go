package repository

import (
	"github.com/gignative-solutions/ba-p2p-headers/data/sql"
	"github.com/gignative-solutions/ba-p2p-headers/domains"
)

type Headers interface {
	AddHeaderToDatabase(header domains.BlockHeader) error
	GetHeaderByHeight(height int32) (*domains.BlockHeader, error)
	GetBlockByHash(args domains.HeaderArgs) (*domains.BlockHeader, error)
	GetHeaderByHeightRange(from int, to int) ([]*domains.BlockHeader, error)
	GetCurrentHeight() (int, error)
	GetHeadersCount() (int, error)
	GetHeaderByHash(hash string) (*domains.BlockHeader, error)
	GenesisExists() bool
	GetPreviousHeader(hash string) (*domains.BlockHeader, error)
	GetTip() (*domains.BlockHeader, error)
	GetConfirmationsCountForBlock(hash string) (int, error)
}

type Tips interface {
	GetConfirmedTip() (*domains.BlockHeader, error)
}

type Repositories struct {
	Headers Headers
	Tips    Tips
}

func NewRepositories(db *sql.HeadersDb) *Repositories {
	return &Repositories{
		Tips:    NewTipRepository(db),
		Headers: NewHeadersRepository(db),
	}
}
