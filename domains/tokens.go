package domains

import (
	"time"
)

// Token represents authorization token.
type Token struct {
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"createdAt"`
}

// CreateToken creates new token.
func CreateToken(value string) *Token{
	return &Token{
		Token:     value,
		CreatedAt: time.Now(),
	}
}
