package testrepository

import (
	"errors"
	"fmt"

	"github.com/libsv/bitcoin-hc/domains"
)

// TokensTestRepository in memory TokensRepository representation for unit testing.
type TokensTestRepository struct {
	db *[]domains.Token
}

// AddTokenToDatabase adds new token to db.
func (r *TokensTestRepository) AddTokenToDatabase(token *domains.Token) error {
	*r.db = append(*r.db, *token)
	return nil
}

// GetTokenByValue returns token from db by given value.
func (r *TokensTestRepository) GetTokenByValue(token string) (*domains.Token, error) {
	for _, t := range *r.db {
		if t.Token == token {
			return &t, nil
		}
	}
	return nil, errors.New("could not find token")
}

// DeleteToken deletes token from db.
func (r *TokensTestRepository) DeleteToken(token string) error {
	fmt.Println("delete token")
	fmt.Println(*r.db)
	fmt.Println(token)

	for i, t := range *r.db {
		if t.Token == token {
			arr := *r.db
			// Replace the element at index i with the last element.
			arr[i] = arr[len(arr)-1]
			// Assign slice without last element.
			*r.db = arr[:len(arr)-1]
			return nil
		}
	}
	return errors.New("could not find token")
}

// NewHeadersTestRepository constructor for HeaderTestRepository.
func NewTokensTestRepository(db *[]domains.Token) *TokensTestRepository {
	return &TokensTestRepository{
		db: db,
	}
}
