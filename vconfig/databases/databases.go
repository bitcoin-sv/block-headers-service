package databases

import (
	"fmt"

	"github.com/gignative-solutions/ba-p2p-headers/vconfig"
	"github.com/jmoiron/sqlx"
)

type dbSetupFunc func(c *vconfig.Db) (*sqlx.DB, error)
type dbSetups map[vconfig.DbType]dbSetupFunc

// NewDbSetup will load the db setup functions into a lookup map
// ready for being called in main.go.
func NewDbSetup() dbSetups {
	s := make(map[vconfig.DbType]dbSetupFunc, 3)
	s[vconfig.DBSqlite] = setupSqliteDB
	return s
}

func (d dbSetups) SetupDb(cfg *vconfig.Db) (*sqlx.DB, error) {
	fn, ok := d[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("db type %s not supported", cfg.Type)
	}
	return fn(cfg)
}
