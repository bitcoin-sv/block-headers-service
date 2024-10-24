package auth

import (
	"strings"

	"github.com/bitcoin-sv/block-headers-service/bhserrors"
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
			bhserrors.AbortWithErrorResponse(c, err, nil)
			return
		}

		token, err := h.getToken(rawToken)
		if err != nil {
			bhserrors.AbortWithErrorResponse(c, err, nil)
			return
		}

		c.Set("token", token)
	}
}

func (h *TokenMiddleware) parseAuthHeader(c *gin.Context) (string, error) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		return "", bhserrors.ErrMissingAuthHeader
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", bhserrors.ErrInvalidAuthHeader
	}

	return headerParts[1], nil
}

func (h *TokenMiddleware) getToken(token string) (*domains.Token, error) {
	t, err := h.tokens.GetToken(token)
	if err != nil {
		return nil, bhserrors.ErrInvalidAccessToken
	}
	return t, nil
}
