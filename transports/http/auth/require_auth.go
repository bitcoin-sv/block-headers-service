package auth

import (
	"github.com/bitcoin-sv/block-headers-service/bhserrors"
	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/gin-gonic/gin"
)

// RequireAdmin adds wrapper to endpoint handler
// that will check if the endpoint was called with admin token.
// This verification will be skipped if authentication isn't enabled.
func RequireAdmin(handler gin.HandlerFunc, requireAdmin bool) gin.HandlerFunc {
	if requireAdmin {
		return func(c *gin.Context) {
			if err := validateToken(c); err == nil {
				handler(c)
			} else {
				bhserrors.AbortWithErrorResponse(c, err, nil)
			}
		}
	}
	return handler
}

func validateToken(c *gin.Context) error {
	token, exist := c.Get("token")
	if !exist {
		return bhserrors.ErrAdminTokenNotFound
	}
	t, ok := token.(*domains.Token)
	if !ok {
		return bhserrors.ErrGeneric
	}
	if !t.IsAdmin {
		return bhserrors.ErrUnauthorized
	}
	return nil // the token is valid
}
