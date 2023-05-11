package auth

import (
	"errors"
	"github.com/libsv/bitcoin-hc/domains"
	p2pservice "github.com/libsv/bitcoin-hc/service"
	"github.com/libsv/bitcoin-hc/vconfig"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	authorizationHeader = "Authorization"
)

// TokenMiddleware middleware that is retrieving token from Authorization header.
type TokenMiddleware struct {
	tokens p2pservice.Tokens
}

// NewAuthTokenMiddleware create Token middleware that is retrieving token from Authorization header.
func NewAuthTokenMiddleware(tokens p2pservice.Tokens) TokenMiddleware {
	return TokenMiddleware{
		tokens: tokens,
	}
}

// Apply is a middleware which checks if the request has a valid token.
func (h *TokenMiddleware) Apply(c *gin.Context) {
	if viper.GetBool(vconfig.EnvHttpServerUseAuth) {
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
	adminToken := viper.GetString(vconfig.EnvHttpServerAuthToken)
	if token == adminToken {
		return domains.CreateAdminToken(token), nil
	}
	t, err := h.tokens.GetToken(token)
	if err != nil {
		return nil, errors.New("invalid access token")
	}
	return t, nil
}
