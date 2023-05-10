package domains

import (
	"time"
)

// Token represents authorization token.
type Token struct {
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"createdAt"`
	IsAdmin   bool      `json:"isAdmin"`
}

// CreateToken creates new token.
func CreateToken(value string) *Token {
	return &Token{
		Token:     value,
		CreatedAt: time.Now(),
	}
}

// CreateAdminToken creates admin token.
func CreateAdminToken(value string) *Token {
	return &Token{
		Token:   value,
		IsAdmin: true,
	}
}
