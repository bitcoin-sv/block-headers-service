package sql

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/libsv/bitcoin-hc/configs"
	"github.com/libsv/bitcoin-hc/repository/dto"
	"github.com/libsv/bitcoin-hc/vconfig"
	"github.com/pkg/errors"
)

const (
	insertBH = "insertheader"

	sqliteInsertHeader = `
	INSERT INTO headers(hash, height, version, merkleroot, nonce, bits, header_state, chainwork, previousblock, timestamp , cumulatedWork)
	VALUES(:hash, :height, :version, :merkleroot, :nonce, :bits, :header_state, :chainwork, :previousblock, :timestamp, :cumulatedWork)
	ON CONFLICT DO NOTHING
	`

	sqlUpdateState = `
	UPDATE headers
	SET header_state = ?
	WHERE hash IN (?)
	`

	sqlHeader = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, header_state, cumulatedWork
	FROM headers
	WHERE hash = ?
	`

	sqlHeaderByHeight = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, header_state, cumulatedWork
	FROM headers
	WHERE height = ? AND header_state = ?
	`

	sqlHeaderByHeightRange = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, header_state, cumulatedWork
	FROM headers
	WHERE height BETWEEN ? AND ?
	`

	sqlLongestChainHeadersFromHeight = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, header_state, cumulatedWork
	FROM headers
	WHERE height >= ? AND header_state = 'LONGEST_CHAIN'
	`

	sqlStaleHeadersFrom = `
	WITH RECURSIVE recur(hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, header_state, cumulatedWork) as (
		select hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, header_state, cumulatedWork
		from headers 
		where hash = ?
		UNION ALL
		SELECT h.hash, h.height, h.version, h.merkleroot, h.nonce, h.bits, h.chainwork, h.previousblock, h.timestamp, h.header_state, h.cumulatedWork
		FROM headers h JOIN recur r
		  ON h.hash = r.previousblock
	)
	select hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, header_state, cumulatedWork
	from recur
	where header_state = 'STALE';
	`

	sqlHighestBlock = `
	SELECT COALESCE(max(height),0) as height
	FROM headers
	`

	sqlHeadersCount = `
	SELECT max(RowId) 
	FROM headers;
	`

	sqlVerifyIfGenesisPresent = `
	SELECT hash 
	FROM headers 
	WHERE height = 0
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
		   prev.header_state,
		   prev.cumulatedWork
	FROM headers h,
		 headers prev
	WHERE h.hash = ?
	  AND h.previousblock = prev.hash
  	`

	sqlSelectTip = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, header_state, cumulatedWork
	FROM headers
	WHERE height = (SELECT max(height) FROM headers where header_state = 'LONGEST_CHAIN')
	`

	sqlSelectAncestorOnHeight = `
    WITH RECURSIVE ancestors(hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, cumulatedWork, level) AS (
        SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, cumulatedWork, 0 level
        FROM headers
        WHERE hash = ?
        UNION ALL
        SELECT h.hash, h.height, h.version, h.merkleroot, h.nonce, h.bits, h.chainwork, h.previousblock, h.timestamp, h.cumulatedWork, a.level + 1 level
        FROM headers h JOIN ancestors a
          ON h.hash = a.previousblock AND h.height >= ?
      )
    SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, cumulatedWork
    FROM ancestors
    WHERE height = ?
    `

	sqlSelectHighestHundred = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, header_state, cumulatedWork
	FROM headers
	ORDER BY height DESC
	LIMIT 100
	`

	sqlSelectTips = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, cumulatedWork, header_state
	FROM headers
	WHERE hash NOT IN (SELECT previousblock
					   FROM headers)
				   `

	sqlChainBetweenTwoHashes = `
	WITH RECURSIVE ancestors(hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, cumulatedWork, level) AS (
		SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, cumulatedWork, 0 level
		FROM headers
		WHERE hash = ?
		UNION ALL
		SELECT h.hash, h.height, h.version, h.merkleroot, h.nonce, h.bits, h.chainwork, h.previousblock, h.timestamp, h.cumulatedWork, a.level + 1 level
		FROM headers h JOIN ancestors a
			ON h.hash = a.previousblock AND h.hash != ?
		)
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, cumulatedWork
	FROM ancestors
	UNION ALL
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previousblock, timestamp, cumulatedWork
	FROM headers
	WHERE hash = ?
	`
)

// HeadersDb represents a database connection and map of related sql queries.
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

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

// Create method will add new record into db.
func (h *HeadersDb) Create(ctx context.Context, req dto.DbBlockHeader) error {
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

// UpdateState will update state of headers of hashes to given state.
func (h *HeadersDb) UpdateState(ctx context.Context, hashes []string, state string) error {
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query, args, err := sqlx.In(sqlUpdateState, state, hashes)
	if err != nil {
		return errors.Wrapf(err, "failed to update headers state to %s", state)
	}
	if _, err := tx.ExecContext(ctx, h.db.Rebind(query), args...); err != nil {
		return errors.Wrapf(err, "failed to update headers state to %s", state)
	}
	return errors.Wrap(tx.Commit(), "failed to commit tx")
}

// Height will return the current highest block height we have stored in the db.
func (h *HeadersDb) Height(ctx context.Context) (int, error) {
	var height int
	if err := h.db.GetContext(ctx, &height, sqlHighestBlock); err != nil {
		return 0, errors.Wrapf(err, "failed to get current block height from cache")
	}
	return height, nil
}

// Count will return the current number of headers in db.
func (h *HeadersDb) Count(ctx context.Context) (int, error) {
	var count int
	if err := h.db.GetContext(ctx, &count, sqlHeadersCount); err != nil {
		return 0, errors.Wrapf(err, "failed to get headers count")
	}

	return count, nil
}

// GetHeaderByHash will return header from db with given hash.
func (h *HeadersDb) GetHeaderByHash(ctx context.Context, hash string) (*dto.DbBlockHeader, error) {
	var bh dto.DbBlockHeader
	if err := h.db.GetContext(ctx, &bh, h.db.Rebind(sqlHeader), hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("could not find hash")
		}
		return nil, errors.Wrapf(err, "failed to get blockhash using hash %s", hash)
	}
	return &bh, nil
}

// GetHeaderByHeight will return header from db with given height and in given state.
func (h *HeadersDb) GetHeaderByHeight(ctx context.Context, height int32, state string) (*dto.DbBlockHeader, error) {
	var bh dto.DbBlockHeader
	if err := h.db.GetContext(ctx, &bh, h.db.Rebind(sqlHeaderByHeight), height, state); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("could not find height")
		}
		return nil, errors.Wrapf(err, "failed to get blockhash using height %d", height)
	}
	return &bh, nil
}

// GetHeaderByHeightRange will return headers from db for given height range (including sended height).
func (h *HeadersDb) GetHeaderByHeightRange(from int, to int) ([]*dto.DbBlockHeader, error) {
	var bh []*dto.DbBlockHeader
	if err := h.db.Select(&bh, h.db.Rebind(sqlHeaderByHeightRange), from, to); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("could not find headers in given range")
		}
		return nil, errors.Wrapf(err, "failed to get headers using given range from: %d to: %d", from, to)
	}
	return bh, nil
}

// GetLongestChainHeadersFromHeight returns from db the headers from "longest chain" starting from given height.
func (h *HeadersDb) GetLongestChainHeadersFromHeight(height int32) ([]*dto.DbBlockHeader, error) {
	var bh []*dto.DbBlockHeader
	if err := h.db.Select(&bh, h.db.Rebind(sqlLongestChainHeadersFromHeight), height); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Errorf("could not find headers in longest chain from height %d", height)
		}
		return nil, errors.Wrapf(err, "failed to get headers in longest chain from height %d", height)
	}
	return bh, nil
}

// GetStaleHeadersBackFrom returns from db all the headers with state STALE, starting from header with hash and preceding that one.
func (h *HeadersDb) GetStaleHeadersBackFrom(hash string) ([]*dto.DbBlockHeader, error) {
	var bh []*dto.DbBlockHeader
	if err := h.db.Select(&bh, h.db.Rebind(sqlStaleHeadersFrom), hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Errorf("header with %s hash does not exist", hash)
		}
		return nil, errors.Wrapf(err, "failed to get headers in stale chain from hash %s", hash)
	}
	return bh, nil
}

// GenesisExists check if genesis header is present in db.
func (h *HeadersDb) GenesisExists(ctx context.Context) bool {
	err := h.db.QueryRow(sqlVerifyIfGenesisPresent)
	return err == nil
}

// GetPreviousHeader will return previous header for this with given hash.
func (h *HeadersDb) GetPreviousHeader(ctx context.Context, hash string) (*dto.DbBlockHeader, error) {
	var bh dto.DbBlockHeader
	if err := h.db.GetContext(ctx, &bh, h.db.Rebind(sqlSelectPreviousBlock), hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("could not find header")
		}
		return nil, errors.Wrapf(err, "failed to get prev header using hash %s", hash)
	}
	return &bh, nil
}

// GetTip will return highest header from db.
func (h *HeadersDb) GetTip(ctx context.Context) (*dto.DbBlockHeader, error) {
	var tip []dto.DbBlockHeader
	if err := h.db.Select(&tip, sqlSelectTip); err != nil {
		configs.Log.Error("sql error", err)
		return nil, errors.Wrap(err, "failed to get tip")
	}
	return &tip[0], nil
}

// GetAncestorOnHeight provides ancestor for a hash on a specified height.
func (h *HeadersDb) GetAncestorOnHeight(hash string, height int32) (*dto.DbBlockHeader, error) {
	var bh []*dto.DbBlockHeader
	if err := h.db.Select(&bh, h.db.Rebind(sqlSelectAncestorOnHeight), hash, int(height), int(height)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("could not find ancestors for a providen hash")
		}
		return nil, errors.Wrapf(err, "failed to get ancestors using given hash: %s ", hash)
	}
	if bh == nil {
		return nil, errors.New("could not find ancestors for a providen hash")
	}
	return bh[0], nil
}

// GetAllTips returns headers whose hash does not appear as previous block hashes in any header within the highest 100 headers.
func (h *HeadersDb) GetAllTips() ([]*dto.DbBlockHeader, error) {
	var bh []*dto.DbBlockHeader
	if err := h.db.Select(&bh, sqlSelectHighestHundred); err != nil {
		return nil, errors.Wrapf(err, "failed to get top hundred headers by height")
	}

	previousBlocks := make([]string, 0, len(bh))
	for _, header := range bh {
		previousBlocks = append(previousBlocks, header.PreviousBlock)
	}

	var tips []*dto.DbBlockHeader
	for _, header := range bh {
		if !contains(previousBlocks, header.Hash) {
			tips = append(tips, header)
		}
	}
	return tips, nil
}

// GetChainBetweenTwoHashes calculates and returnes chain between 2 hashes.
func (h *HeadersDb) GetChainBetweenTwoHashes(low string, high string) ([]*dto.DbBlockHeader, error) {
	var bh []*dto.DbBlockHeader
	if err := h.db.Select(&bh, h.db.Rebind(sqlChainBetweenTwoHashes), high, low, low); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("could not find headers in given range")
		}
		return nil, errors.Wrapf(err, "failed to get headers using given range from: %s to: %s", low, high)
	}
	return bh, nil
}
