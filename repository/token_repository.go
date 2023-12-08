package repository

import (
	"context"

	"github.com/bitcoin-sv/pulse/database/sql"
	"github.com/bitcoin-sv/pulse/domains"
	dto "github.com/bitcoin-sv/pulse/repository/dto"
)

// TokenRepository provide access to repositories and implements methods for token.
type TokenRepository struct {
	db *sql.HeadersDb
}

// AddTokenToDatabase adds new token to db.
func (r *TokenRepository) AddTokenToDatabase(token *domains.Token) error {
	dbToken := dto.ToDbToken(token)
	err := r.db.CreateToken(context.Background(), dbToken)
	return err
}

// GetTokenByValue returns token from db by given value.
func (r *TokenRepository) GetTokenByValue(token string) (*domains.Token, error) {
	t, err := r.db.GetTokenByValue(context.Background(), token)
	if err != nil {
		return nil, err
	}
	dbt := t.ToToken()
	return dbt, err
}

// DeleteToken deletes token from db.
func (r *TokenRepository) DeleteToken(token string) error {
	err := r.db.DeleteToken(context.Background(), token)
	return err
}

// NewTokensRepository creates and returns TokenRepository instance.
func NewTokensRepository(db *sql.HeadersDb) *TokenRepository {
	return &TokenRepository{db: db}
}
