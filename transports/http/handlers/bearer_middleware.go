// Package handler provides HTTP handlers.
package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const useAuth = "http.server.useAuth"
const authToken = "http.server.authToken" //nolint:gosec
const adminEndpoints = "http.server.adminOnly"

const (
	authorizationHeader = "Authorization"
)

// tokenIdentity is a middleware which checks if the request has a valid token.
func (h *Handler) tokenIdentity(c *gin.Context) {
	if viper.GetBool(useAuth) {
		err := h.parseAuthHeader(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
		}
	}
}

// parseAuthHeader parses the Authorization header and checks if the token is valid.
func (h *Handler) parseAuthHeader(c *gin.Context) error {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		return errors.New("empty auth header")
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return errors.New("invalid auth header")
	}

	aToken := viper.GetString(authToken)
	// check if given token is the global token
	if headerParts[1] != aToken {
		path := c.Request.URL.Path

		// if endpoint is only for admin tokens, return error
		// if not check if token is valid
		if isAdminEndpoint(path) {
			return errors.New("invalid auth token")
		} else if _, err := h.services.Tokens.GetToken(headerParts[1]); err != nil {
			return errors.New("invalid access token")
		}
	}

	return nil
}

// isAdminEndpoint checks if the given path is an admin endpoint
// or if it can be authorized by generated tokens.
func isAdminEndpoint(path string) bool {
	endpoints := viper.GetStringSlice(adminEndpoints)
	prefix := viper.GetString(urlPrefix)

	for _, v := range endpoints {
		if (prefix + v) == path {
			return true
		}
	}

	return false
}
