package sql

import (
	"context"
	"database/sql"

	"github.com/gignative-solutions/ba-p2p-headers/configs"
	"github.com/gignative-solutions/ba-p2p-headers/domains"
	"github.com/gignative-solutions/ba-p2p-headers/vconfig"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	insertBH = "insertheader"

	sqliteInsertHeader = `
	INSERT INTO headers(hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, isorphan, isconfirmed, cumulatedWork)
	VALUES(:hash, :height, :version, :merkleroot, :nonce, :bits, :chainwork, :previousblock, :timestamp, :isorphan, :isconfirmed, :cumulatedWork)
	ON CONFLICT DO NOTHING
	`

	sqlHeader = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, isorphan, isconfirmed, cumulatedWork
	FROM headers
	WHERE hash = ?
	`

	sqlHeaderByHeight = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, isorphan, isconfirmed, cumulatedWork
	FROM headers
	WHERE height = ?
	`

	sqlHeaderByHeightRange = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, isorphan, isconfirmed, cumulatedWork
	FROM headers
	WHERE height BETWEEN ? AND ?
	`

	sqlHighestBlock = `
	SELECT COALESCE(max(height),0) as height
	FROM headers
	`

	sqlHeadersCount = `
	SELECT max(RowId) 
	FROM headers;
	`

	sqlGetTip = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, isorphan, isconfirmed, cumulatedWork
	FROM headers 
	ORDER BY height DESC LIMIT 1;
	`

	sqlVerifyIfGenesisPresent = `
	SELECT hash 
	FROM headers 
	WHERE height = 0
	`

	sqlCalculateConfirmations = `
	WITH RECURSIVE recur(hash, height, cumulatedwork, confirmations) AS (
		SELECT hash, height, cumulatedwork, 1 confirmations
		FROM headers
		WHERE hash = ?
		UNION ALL
		SELECT h.hash, h.height, h.cumulatedwork, confirmations + 1
		FROM headers h JOIN recur r
		  ON h.previousblock = r.hash
	  )
	  SELECT MAX(confirmations)
	  FROM recur
	  WHERE CAST(cumulatedwork AS INTEGER) = (SELECT MAX(CAST(cumulatedwork AS INTEGER)) FROM recur)
	`

	sqlSelectPreviousBlock = `
	SELECT prev.hash,
		   prev.height,
		   prev.version,
		   prev.merkleroot,
		   prev.nonce,
		   prev.bits,
		   prev.chainwork,
		   prev.previousblock,
		   prev.timestamp,
		   prev.isorphan,
		   prev.isconfirmed,
		   prev.cumulatedWork
	FROM headers h,
		 headers prev
	WHERE h.hash = ?
	  AND h.previousblock = prev.hash
  	`

	sqlSelectTip = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, isorphan, isconfirmed, cumulatedWork
	FROM headers
	WHERE height = (SELECT max(height)
					FROM headers)
	`
)

type HeadersDb struct {
	dbType vconfig.DbType
	db     *sqlx.DB
	sqls   map[vconfig.DbType]map[string]string
}

// NewHeadersDb will setup and return a new headers store.
func NewHeadersDb(db *sqlx.DB, dbType vconfig.DbType) *HeadersDb {
	return &HeadersDb{
		dbType: dbType,
		db:     db,
		sqls: map[vconfig.DbType]map[string]string{
			vconfig.DBSqlite: {
				insertBH: sqliteInsertHeader,
			},
		},
	}
}

func (h *HeadersDb) Create(ctx context.Context, req domains.DbBlockHeader) error {
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
func (h *HeadersDb) CreateBatch(ctx context.Context, req []*domains.BlockHeader) error {
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
func (h *HeadersDb) Header(ctx context.Context, args domains.HeaderArgs) (*domains.DbBlockHeader, error) {
	var bh domains.DbBlockHeader
	if err := h.db.GetContext(ctx, &bh, h.db.Rebind(sqlHeader), args.Blockhash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("could not find blockhash")
		}
		return nil, errors.Wrapf(err, "failed to get blockhash using hash %s", args.Blockhash)
	}
	return &bh, nil
}

// Height will return the current highest block height we have stored in the db.
func (h *HeadersDb) Height(ctx context.Context) (int, error) {
	var height int
	if err := h.db.GetContext(ctx, &height, sqlHighestBlock); err != nil {
		return 0, errors.Wrapf(err, "failed to get current block height from cache")
	}
	return height, nil
}

func (h *HeadersDb) Count(ctx context.Context) (int, error) {
	var count int
	if err := h.db.GetContext(ctx, &count, sqlHeadersCount); err != nil {
		return 0, errors.Wrapf(err, "failed to get headers count")
	}

	return count, nil
}

func (h *HeadersDb) GetHeaderByHash(ctx context.Context, hash string) (*domains.DbBlockHeader, error) {
	var bh domains.DbBlockHeader
	if err := h.db.GetContext(ctx, &bh, h.db.Rebind(sqlHeader), hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("could not find hash")
		}
		return nil, errors.Wrapf(err, "failed to get blockhash using hash %s", hash)
	}
	return &bh, nil
}

func (h *HeadersDb) GetHeaderByHeight(ctx context.Context, height int32) (*domains.DbBlockHeader, error) {
	var bh domains.DbBlockHeader
	if err := h.db.GetContext(ctx, &bh, h.db.Rebind(sqlHeaderByHeight), height); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("could not find height")
		}
		return nil, errors.Wrapf(err, "failed to get blockhash using height %s", height)
	}
	return &bh, nil
}

func (h *HeadersDb) GetHeaderByHeightRange(from int, to int) ([]*domains.DbBlockHeader, error) {
	var bh []*domains.DbBlockHeader
	if err := h.db.Select(&bh, h.db.Rebind(sqlHeaderByHeightRange), from, to); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("could not find headers in given range")
		}
		return nil, errors.Wrapf(err, "failed to get headers using given range from: %s to: %s", from, to)
	}
	return bh, nil
}

func (h *HeadersDb) GetCurrentTip(ctx context.Context) (*domains.DbBlockHeader, error) {
	var bh domains.DbBlockHeader
	if err := h.db.GetContext(ctx, &bh, h.db.Rebind(sqlGetTip)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no records in db")
		}
		return nil, errors.Wrapf(err, "failed to get tip block %s")
	}

	return &bh, nil
}

func (h *HeadersDb) GenesisExists(ctx context.Context) bool {
	err := h.db.QueryRow(sqlVerifyIfGenesisPresent)
	return err == nil
}

func (h *HeadersDb) CalculateConfirmations(ctx context.Context, hash string) (int, error) {
	var amount int
	if err := h.db.Select(&amount, h.db.Rebind(sqlCalculateConfirmations), hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.Errorf("header with %s hash does not exist", hash)
		}
		return 0, errors.Wrapf(err, "failed to calculate confirmations for %s hash", hash)
	}
	return amount, nil
}

func (h *HeadersDb) GetPreviousHeader(ctx context.Context, hash string) (*domains.DbBlockHeader, error) {
	var bh domains.DbBlockHeader
	if err := h.db.GetContext(ctx, &bh, h.db.Rebind(sqlSelectPreviousBlock), hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("could not find header")
		}
		return nil, errors.Wrapf(err, "failed to get prev header using hash %s", hash)
	}
	return &bh, nil
}

func (h *HeadersDb) GetTip(ctx context.Context) (*domains.DbBlockHeader, error) {
	var tip []domains.DbBlockHeader
	if err := h.db.Select(&tip, sqlSelectTip); err != nil {
		configs.Log.Error("sql error", err)
		return nil, errors.Wrap(err, "failed to get tip")
	}
	return &tip[0], nil
}
