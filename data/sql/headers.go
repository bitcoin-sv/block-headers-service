package sql

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	lathos "github.com/theflyingcodr/lathos/errs"

	headers "github.com/libsv/bitcoin-hc"
	"github.com/libsv/bitcoin-hc/config"
)

const (
	insertBH = "insertblockheader"

	sqliteInsertBlockHeader = `
	INSERT INTO blockheaders(hash,confirmations, height, version, versionhex, merkleroot, time, mediantime, nonce, bits, difficulty, chainwork,previousblockhash,nextblockhash)
	VALUES(:hash, :confirmations, :height, :version, :versionhex, :merkleroot, :time, :mediantime, :nonce, :bits, :difficulty, :chainwork, :previousblockhash, :nextblockhash)
	ON CONFLICT DO NOTHING
	`

	mysqlInsertBlockHeader = `
	INSERT INTO blockheaders(hash,confirmations, height, version, versionhex, merkleroot, time, mediantime, nonce, bits, difficulty, chainwork,previousblockhash,nextblockhash)
	VALUES(:hash, :confirmations, :height, :version, :versionhex, :merkleroot, :time, :mediantime, :nonce, :bits, :difficulty, :chainwork, :previousblockhash, :nextblockhash)
	ON DUPLICATE KEY UPDATE hash=hash
	`

	sqlBlockHeader = `
	SELECT hash, confirmations, height, version, versionhex, merkleroot, time, mediantime, nonce, bits, difficulty, chainwork,previousblockhash,nextblockhash
	FROM blockheaders
	WHERE hash = ?
	`

	sqlHighestBlock = `
	SELECT COALESCE(max(height),0) as height
	FROM blockheaders
	`
)

type headersDb struct {
	dbType config.DbType
	db     *sqlx.DB
	sqls   map[config.DbType]map[string]string
}

// NewHeadersDb will setup and return a new headers store.
func NewHeadersDb(db *sqlx.DB, dbType config.DbType) *headersDb {
	return &headersDb{
		dbType: dbType,
		db:     db,
		sqls: map[config.DbType]map[string]string{
			config.DBMySql: {
				insertBH: mysqlInsertBlockHeader,
			},
			config.DBSqlite: {
				insertBH: sqliteInsertBlockHeader,
			},
			config.DBPostgres: {
				insertBH: sqliteInsertBlockHeader,
			},
		},
	}
}

func (h *headersDb) Create(ctx context.Context, req headers.BlockHeader) error {
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if _, err := tx.NamedExecContext(ctx, h.sqls[h.dbType][insertBH], req); err != nil {
		return errors.Wrap(err, "failed to insert header")
	}
	return errors.Wrap(tx.Commit(), "failed to commit tx")
}

// CreateBatch will add a batch of records to the data store.
func (h *headersDb) CreateBatch(ctx context.Context, req []*headers.BlockHeader) error {
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	for i := 0; i < len(req); i += 1000 {
		if i+1000 > len(req) {
			if _, err = tx.NamedExecContext(ctx, h.sqls[h.dbType][insertBH], req[i:]); err != nil {
				return errors.Wrap(err, "failed to bulk insert headers")
			}
			break
		}
		if _, err = tx.NamedExecContext(ctx, h.sqls[h.dbType][insertBH], req[i:i+100]); err != nil {
			return errors.Wrap(err, "failed to bulk insert headers")
		}
	}

	return errors.Wrap(tx.Commit(), "failed to commit tx")
}

// Header will return a single block header by blockhash.
func (h *headersDb) Header(ctx context.Context, args headers.HeaderArgs) (*headers.BlockHeader, error) {
	var bh headers.BlockHeader
	if err := h.db.GetContext(ctx, &bh, h.db.Rebind(sqlBlockHeader), args.Blockhash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, lathos.NewErrNotFound("N001", "could not find blockhash")
		}
		return nil, errors.Wrapf(err, "failed to get blockhash using hash %s", args.Blockhash)
	}
	return &bh, nil
}

// Height will return the current highest block height we have stored in the db.
func (h *headersDb) Height(ctx context.Context) (int, error) {
	var height int
	if err := h.db.GetContext(ctx, &height, sqlHighestBlock); err != nil {
		return 0, errors.Wrapf(err, "failed to get current block height from cache")
	}
	return height, nil
}
