package service

import (
	"github.com/bitcoin-sv/pulse/domains"
	"github.com/bitcoin-sv/pulse/repository"
	"github.com/dchest/uniuri"
)

// TokenService represents Token service and provide access to repositories.
type TokenService struct {
	repo       *repository.Repositories
	adminToken string
}

// NewTokenService creates and returns TokenService instance.
func NewTokenService(repo *repository.Repositories, adminToken string) *TokenService {
	return &TokenService{
		repo:       repo,
		adminToken: adminToken,
	}
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

// GetToken returns token by given value.
func (s *TokenService) GetToken(token string) (*domains.Token, error) {
	if token == s.adminToken {
		return domains.CreateAdminToken(token), nil
	}
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
