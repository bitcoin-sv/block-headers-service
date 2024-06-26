package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/service"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeader = "Authorization"
)

// TokenMiddleware middleware that is retrieving token from Authorization header.
type TokenMiddleware struct {
	tokens service.Tokens
	cfg    *config.HTTPConfig
}

// NewMiddleware create Token middleware that is retrieving token from Authorization header.
func NewMiddleware(s *service.Services, cfg *config.HTTPConfig) *TokenMiddleware {
	return &TokenMiddleware{
		tokens: s.Tokens,
		cfg:    cfg,
	}
}

// ApplyToAPI is a middleware which checks if the request has a valid token.
func (h *TokenMiddleware) ApplyToAPI(c *gin.Context) {
	if h.cfg.UseAuth {
		rawToken, err := h.parseAuthHeader(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
			return
		}

		token, err := h.getToken(rawToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
			return
		}

		c.Set("token", token)
	}
}

func (h *TokenMiddleware) parseAuthHeader(c *gin.Context) (string, error) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		return "", errors.New("empty auth header")
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", errors.New("invalid auth header")
	}

	return headerParts[1], nil
}

func (h *TokenMiddleware) getToken(token string) (*domains.Token, error) {
	t, err := h.tokens.GetToken(token)
	if err != nil {
		return nil, errors.New("invalid access token")
	}
	return t, nil
}
