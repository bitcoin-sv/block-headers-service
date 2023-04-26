package service

import (
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/repository"

	"github.com/dchest/uniuri"
)

// TokenService represents Token service and provide access to repositories.
type TokenService struct {
	repo *repository.Repositories
}

// GenerateToken generates and save new token.
func (s *TokenService) GenerateToken() (*domains.Token, error) {
	tValue := uniuri.NewLen(32)
	token := domains.CreateToken(tValue)
	err := s.repo.Tokens.AddTokenToDatabase(token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// GetToken returns token from db by given value.
func (s *TokenService) GetToken(token string) (*domains.Token, error) {
	t, err := s.repo.Tokens.GetTokenByValue(token)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// DeleteToken deletes token from db.
func (s *TokenService) DeleteToken(token string) error {
	return s.repo.Tokens.DeleteToken(token)
}

// NewTokenService creates and returns TokenService instance.
func NewTokenService(repo *repository.Repositories) *TokenService {
	return &TokenService{repo: repo}
}
