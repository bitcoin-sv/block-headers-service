package auth

import (
	"errors"
	"net/http"

	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/domains"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// RequireAdmin adds wrapper to endpoint handler
// that will check if the endpoint was called with admin token.
// This verification will be skipped if authentication isn't enabled.
func RequireAdmin(handler gin.HandlerFunc) gin.HandlerFunc {
	if viper.GetBool(config.EnvHttpServerUseAuth) {
		return func(c *gin.Context) {
			token, exist := c.Get("token")
			if !exist {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			t, ok := token.(*domains.Token)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, errors.New("something went wrong"))
			}
			if !t.IsAdmin {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			handler(c)
		}
	}
	return handler
}
