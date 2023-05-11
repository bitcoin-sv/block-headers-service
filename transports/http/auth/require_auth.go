package auth

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/vconfig"
	"github.com/spf13/viper"
	"net/http"
)

// RequireAdmin adds wrapper to endpoint handler
// that will check if the endpoint was called with admin token.
// This verification will be skipped if authentication isn't enabled.
func RequireAdmin(handler gin.HandlerFunc) gin.HandlerFunc {
	if viper.GetBool(vconfig.EnvHttpServerUseAuth) {
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
