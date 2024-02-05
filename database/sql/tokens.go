package sql

import (
	"context"
	"database/sql"

	"github.com/bitcoin-sv/pulse/repository/dto"
	"github.com/pkg/errors"
)

const (
	sqlInsertToken = `
	INSERT INTO tokens(token, created_at)
	VALUES(:token, :created_at)
	ON CONFLICT DO NOTHING
	`

	// nolint:gosec
	sqlGetToken = ` 
	SELECT token, created_at
	FROM tokens
	WHERE token = ?
	`

	sqlDeleteToken = `
	DELETE FROM tokens
	WHERE token = :token
	`
)

// CreateToken method will add new record into db.
func (h *HeadersDb) CreateToken(ctx context.Context, token *dto.DbToken) error {
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.NamedExecContext(ctx, h.db.Rebind(sqlInsertToken), *token); err != nil {
		return errors.Wrap(err, "failed to insert token")
	}
	return errors.Wrap(tx.Commit(), "failed to commit tx")
}

// GetTokenByValue method will search and return token by value.
func (h *HeadersDb) GetTokenByValue(ctx context.Context, token string) (*dto.DbToken, error) {
	var dbToken dto.DbToken
	if err := h.db.GetContext(ctx, &dbToken, h.db.Rebind(sqlGetToken), token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("could not find token")
		}
		return nil, errors.Wrapf(err, "failed to get token using value %s", token)
	}
	return &dbToken, nil
}

// DeleteToken method will delete token from db.
func (h *HeadersDb) DeleteToken(ctx context.Context, token string) error {
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err = tx.NamedExecContext(ctx, h.db.Rebind(sqlDeleteToken), map[string]interface{}{"token": token}); err != nil {
		return errors.Wrap(err, "failed to delete token")
	}

	return errors.Wrap(tx.Commit(), "failed to commit tx")
}
