package dto

import (
	"time"

	"github.com/bitcoin-sv/block-headers-service/domains"
)

// DbToken represent authorization token saved in db.
type DbToken struct {
	Token     string    `db:"token"`
	CreatedAt time.Time `db:"created_at"`
}

// ToToken converts DbToken to Token.
func (dbt *DbToken) ToToken() *domains.Token {
	return &domains.Token{
		Token:     dbt.Token,
		CreatedAt: dbt.CreatedAt,
	}
}

// ToDbToken converts Token to DbToken.
func ToDbToken(t *domains.Token) *DbToken {
	return &DbToken{
		Token:     t.Token,
		CreatedAt: t.CreatedAt,
	}
}
