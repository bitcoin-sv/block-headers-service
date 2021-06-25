package sqlite

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	lathos "github.com/theflyingcodr/lathos/errs"

	"github.com/libsv/headers-client"
)

const (
	sqlInsertBlockHeader = `
	INSERT INTO blockheaders(hash,confirmations, height, version, versionHex, merkleroot, time, mediantime, nonce, bits, difficulty, chainwork,previousblockhash,nextblockhash)
	VALUES(:hash, :confirmations, :height, :version, :versionHex, :merkleroot, :time, :mediantime, :nonce, :bits, :difficulty, :chainwork, :previousblockhash, :nextblockhash)
	ON CONFLICT DO NOTHING
	`

	sqlBlockHeader = `
	SELECT hash, confirmations, height, version, versionHex, merkleroot, time, mediantime, nonce, bits, difficulty, chainwork,previousblockhash,nextblockhash
	FROM blockheaders
	WHERE hash = :blockHash
	`

	sqlHeighestBlock = `
	SELECT COALESCE(max(height),0) as height
	FROM blockheaders
	`
)

type headersDb struct {
	db *sqlx.DB
}

func NewHeadersDb(db *sqlx.DB) *headersDb {
	return &headersDb{db: db}
}

func (h *headersDb) Create(ctx context.Context, req headers.BlockHeader) error {
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if _, err := tx.NamedExecContext(ctx, sqlInsertBlockHeader, req); err != nil {
		return errors.Wrap(err, "failed to insert header")
	}
	return errors.Wrap(tx.Commit(), "failed to commit tx")
}

// Header will return a single block header by blockhash.
func (h *headersDb) Header(ctx context.Context, args headers.HeaderArgs) (*headers.BlockHeader, error) {
	var bh headers.BlockHeader
	if err := h.db.GetContext(ctx, &bh, sqlBlockHeader, args.Blockhash); err != nil {
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
	if err := h.db.GetContext(ctx, &height, sqlHeighestBlock); err != nil {
		return 0, errors.Wrapf(err, "failed to get current block height from cache")
	}
	return height, nil
}
