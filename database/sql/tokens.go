package sql

import (
	"context"

	"github.com/bitcoin-sv/block-headers-service/bhserrors"
	"github.com/bitcoin-sv/block-headers-service/repository/dto"
)

const (
	sqlInsertToken = `
	INSERT INTO tokens(token, created_at)
	VALUES(:token, :created_at)
	ON CONFLICT DO NOTHING
	`

	//nolint:gosec
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
		return bhserrors.ErrCreateToken.Wrap(err)
	}

	if err = tx.Commit(); err != nil {
		return bhserrors.ErrCreateToken.Wrap(err)
	}

	return nil
}

// GetTokenByValue method will search and return token by value.
func (h *HeadersDb) GetTokenByValue(ctx context.Context, token string) (*dto.DbToken, error) {
	var dbToken dto.DbToken
	if err := h.db.GetContext(ctx, &dbToken, h.db.Rebind(sqlGetToken), token); err != nil {
		return nil, bhserrors.ErrTokenNotFound.Wrap(err)
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
		return bhserrors.ErrDeleteToken.Wrap(err)
	}

	if err = tx.Commit(); err != nil {
		return bhserrors.ErrDeleteToken.Wrap(err)

	}

	return nil
}
