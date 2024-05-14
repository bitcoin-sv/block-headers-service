package sql

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/bitcoin-sv/block-headers-service/repository/dto"
)

const (
	HeadersTableName = "headers"

	sqlInsertHeader = `
	INSERT INTO headers(hash, height, version, merkleroot, nonce, bits, header_state, chainwork, previous_block, timestamp , cumulated_work)
	VALUES(:hash, :height, :version, :merkleroot, :nonce, :bits, :header_state, :chainwork, :previous_block, :timestamp, :cumulated_work)
	ON CONFLICT DO NOTHING
	`

	sqlUpdateState = `
	UPDATE headers
	SET header_state = ?
	WHERE hash IN (?)
	`

	sqlHeader = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, header_state, cumulated_work
	FROM headers
	WHERE hash = ?
	`

	sqlHeaderByHeight = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, header_state, cumulated_work
	FROM headers
	WHERE height = ? AND header_state = ?
	`

	sqlHeaderByHeightRange = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, header_state, cumulated_work
	FROM headers
	WHERE height BETWEEN ? AND ?
	`

	sqlLongestChainHeadersFromHeight = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, header_state, cumulated_work
	FROM headers
	WHERE height >= ? AND header_state = 'LONGEST_CHAIN'
	`

	sqlStaleHeadersFrom = `
	WITH RECURSIVE recur(hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, header_state, cumulated_work) as (
		select hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, header_state, cumulated_work
		from headers 
		where hash = ?
		UNION ALL
		SELECT h.hash, h.height, h.version, h.merkleroot, h.nonce, h.bits, h.chainwork, h.previous_block, h.timestamp, h.header_state, h.cumulated_work
		FROM headers h JOIN recur r
		  ON h.hash = r.previous_block
	)
	select hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, header_state, cumulated_work
	from recur
	where header_state = 'STALE';
	`

	sqlHighestBlock = `
	SELECT COALESCE(max(height),0) as height
	FROM headers
	`

	sqlHeadersCount = `
	SELECT COUNT(1)
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
		   prev.previous_block,
		   prev.timestamp,
		   prev.header_state,
		   prev.cumulated_work
	FROM headers h,
		 headers prev
	WHERE h.hash = ?
	  AND h.previous_block = prev.hash
  	`

	sqlSelectTip = `
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, header_state, cumulated_work
	FROM headers
	WHERE height = (SELECT max(height) FROM headers where header_state = 'LONGEST_CHAIN')
	`

	sqlSelectAncestorOnHeight = `
    WITH RECURSIVE ancestors(hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, cumulated_work, level) AS (
        SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, cumulated_work, 0 level
        FROM headers
        WHERE hash = ?
        UNION ALL
        SELECT h.hash, h.height, h.version, h.merkleroot, h.nonce, h.bits, h.chainwork, h.previous_block, h.timestamp, h.cumulated_work, a.level + 1 level
        FROM headers h JOIN ancestors a
          ON h.hash = a.previous_block AND h.height >= ?
      )
    SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, cumulated_work
    FROM ancestors
    WHERE height = ?
    `

	sqlSelectTips = `
	with mainTip as (
	select hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, header_state, cumulated_work
	from headers
	where header_state = 'LONGEST_CHAIN'
	order by height desc
	limit 1
	)
	select hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, header_state, cumulated_work
	from mainTip
	union
	select hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, header_state, cumulated_work
	from headers
	where header_state != 'LONGEST_CHAIN' and
			hash not in (select previous_block from headers where header_state != 'LONGEST_CHAIN')
				   `

	sqlChainBetweenTwoHashes = `
	WITH RECURSIVE ancestors(hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, cumulated_work, level) AS (
		SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, cumulated_work, 0 level
		FROM headers
		WHERE hash = ?
		UNION ALL
		SELECT h.hash, h.height, h.version, h.merkleroot, h.nonce, h.bits, h.chainwork, h.previous_block, h.timestamp, h.cumulated_work, a.level + 1 level
		FROM headers h JOIN ancestors a
			ON h.hash = a.previous_block AND h.hash != ?
		)
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, cumulated_work
	FROM ancestors
	UNION ALL
	SELECT hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, cumulated_work
	FROM headers
	WHERE hash = ?
	`

	sqlTipOfChainHeight = `SELECT MAX(height) FROM headers WHERE header_state = 'LONGEST_CHAIN'`

	sqlVerifyHash = `SELECT hash FROM headers WHERE merkleroot = $1 AND height = $2 AND header_state = 'LONGEST_CHAIN'`

	sqlGetHeadersBeginning = `
	WITH AllHashes (hash) AS (
		VALUES 
		(?)
	`

	sqlGetHeadersHash = `,
		(?)`

	sqlGetHeadersEnd = `
	)
	SELECT 
    	COALESCE(h.height, '0') AS height,
    	a.hash
	FROM AllHashes a
	LEFT JOIN headers h ON a.hash = h.hash;
	`

	sqlHeaderByHeightRangeLongestChain = `
	SELECT 
		hash, height, version, merkleroot, nonce, bits, chainwork, previous_block, timestamp, header_state, cumulated_work
	FROM headers
	WHERE height BETWEEN ? AND ? AND header_state = 'LONGEST_CHAIN';
	`
)

// HeadersDb represents a database connection and map of related sql queries.
type HeadersDb struct {
	db  *sqlx.DB
	log *zerolog.Logger
}

// NewHeadersDb will setup and return a new headers store.
func NewHeadersDb(db *sqlx.DB, log *zerolog.Logger) *HeadersDb {
	headerLogger := log.With().Str("subservice", "headers-db").Logger()
	return &HeadersDb{
		db:  db,
		log: &headerLogger,
	}
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
	if _, err := tx.NamedExecContext(ctx, sqlInsertHeader, req); err != nil {
		return errors.Wrap(err, "failed to insert header")
	}
	return errors.Wrap(tx.Commit(), "failed to commit tx")
}

// CreateMultiple method will add multiple new records into db.
func (h *HeadersDb) CreateMultiple(ctx context.Context, headers []dto.DbBlockHeader) error {
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for _, record := range headers {
		if _, err := tx.NamedExecContext(ctx, sqlInsertHeader, record); err != nil {
			return errors.Wrap(err, "failed to insert header")
		}
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
		h.log.Error().Msgf("sql error: %v", err)
		return nil, errors.Wrap(err, "failed to get tip")
	}
	if len(tip) == 0 {
		return nil, errors.New("could not find tip")
	}

	return &tip[0], nil
}

// GetAncestorOnHeight provides ancestor for a hash on a specified height.
func (h *HeadersDb) GetAncestorOnHeight(hash string, height int32) (*dto.DbBlockHeader, error) {
	var bh []*dto.DbBlockHeader
	if err := h.db.Select(&bh, h.db.Rebind(sqlSelectAncestorOnHeight), hash, int(height), int(height)); err != nil {
		return nil, errors.Wrapf(err, "failed to get ancestors using given hash: %s ", hash)
	}
	if len(bh) == 0 {
		return nil, errors.New("could not find ancestors for a providen hash")
	}
	return bh[0], nil
}

// GetAllTips returns all tips from db.
func (h *HeadersDb) GetAllTips() ([]*dto.DbBlockHeader, error) {
	var bh []*dto.DbBlockHeader
	if err := h.db.Select(&bh, sqlSelectTips); err != nil {
		return nil, errors.Wrapf(err, "failed to get tips")
	}
	return bh, nil
}

// GetChainBetweenTwoHashes calculates and returnes chain between 2 hashes.
func (h *HeadersDb) GetChainBetweenTwoHashes(low string, high string) ([]*dto.DbBlockHeader, error) {
	var bh []*dto.DbBlockHeader
	if err := h.db.Select(&bh, h.db.Rebind(sqlChainBetweenTwoHashes), high, low, low); err != nil {
		return nil, errors.Wrapf(err, "failed to get headers using given range from: %s to: %s", low, high)
	}
	if len(bh) == 0 {
		return nil, errors.New("could not find headers in given range")
	}
	return bh, nil
}

// GetMerkleRootsConfirmations returns confirmation of merkle roots inclusion in the longest chain.
func (h *HeadersDb) GetMerkleRootsConfirmations(
	request []domains.MerkleRootConfirmationRequestItem,
) ([]*dto.DbMerkleRootConfirmation, error) {
	confirmations := make([]*dto.DbMerkleRootConfirmation, 0)
	tipHeight, err := h.getChainTipHeight()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get chain tip height")
	}

	for _, item := range request {
		confirmation, err := h.getMerkleRootConfirmation(item, tipHeight)
		if err != nil {
			continue
		}
		confirmations = append(confirmations, confirmation)
	}

	return confirmations, nil
}

// GetHeadersHeightOfLocators returns hash and height from db with given locators.
func (h *HeadersDb) GetHeadersHeightOfLocators(hashtable []interface{}, hashStop *chainhash.Hash) (bh []*dto.DbBlockHeader, err error) {
	query := sqlGetHeadersBeginning
	for i := 1; i < len(hashtable); i++ {
		query += sqlGetHeadersHash
	}
	query += sqlGetHeadersEnd

	if err := h.db.Select(&bh, h.db.Rebind(query), hashtable...); err != nil {
		h.log.Error().Err(err).Msg("Failed to get headers by height range")
		return nil, err
	}
	return bh, nil
}

// GetHashStopHeight will return header from db with given hash.
func (h *HeadersDb) GetHashStopHeight(hashStop string) (int32, error) {
	var dbHashStop dto.DbBlockHeader
	if err := h.db.Get(&dbHashStop, h.db.Rebind(sqlHeader), hashStop); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, errors.Wrapf(err, "failed to get stophash %s", hashStop)
		}
		dbHashStop.Height = wire.MaxCFHeadersPerMsg
	}
	return dbHashStop.Height, nil
}

// GetHeadersByHeightRange returns headers from db in specified height range.
func (h *HeadersDb) GetHeadersByHeightRange(from int, to int) ([]*dto.DbBlockHeader, error) {
	var listOfHeaders []*dto.DbBlockHeader
	if err := h.db.Select(&listOfHeaders, h.db.Rebind(sqlHeaderByHeightRangeLongestChain), from, to); err != nil {
		return nil, errors.Wrapf(err, "failed to get headers using given range from: %d to: %d", from, to)
	}
	return listOfHeaders, nil
}

func (h *HeadersDb) getChainTipHeight() (int32, error) {
	var tipHeight int32
	err := h.db.Get(&tipHeight, sqlTipOfChainHeight)
	return tipHeight, err
}

func (h *HeadersDb) getMerkleRootConfirmation(item domains.MerkleRootConfirmationRequestItem, tipHeight int32) (*dto.DbMerkleRootConfirmation, error) {
	confirmation := &dto.DbMerkleRootConfirmation{
		MerkleRoot:  item.MerkleRoot,
		BlockHeight: item.BlockHeight,
		TipHeight:   tipHeight,
	}

	var hash sql.NullString
	err := h.db.Get(&hash, sqlVerifyHash, item.MerkleRoot, item.BlockHeight)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
		} else {
			return nil, err
		}
	} else {
		confirmation.Hash = hash
	}

	return confirmation, nil
}
