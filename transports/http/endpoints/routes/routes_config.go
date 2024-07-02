package router

import (
	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/gin-gonic/gin"
)

// RootEndpointsFunc wrapping type for function to mark it as implementation of RootEndpoints.
type RootEndpointsFunc func(router *gin.RouterGroup)

// RootEndpoints registrar which will register routes in root context of application.
type RootEndpoints interface {
	// RegisterEndpoints register root endpoints.
	RegisterEndpoints(router *gin.RouterGroup)
}

// APIEndpointsFunc wrapping type for function to mark it as implementation of APIEndpoints.
type APIEndpointsFunc func(router *gin.RouterGroup)

// APIEndpoints registrar which will register routes in API routes group.
type APIEndpoints interface {
	// RegisterAPIEndpoints register API endpoints.
	RegisterAPIEndpoints(router *gin.RouterGroup, cfg *config.HTTPConfig)
}

// RegisterEndpoints register root endpoints by registrar RootEndpointsFunc.
func (f RootEndpointsFunc) RegisterEndpoints(router *gin.RouterGroup) {
	f(router)
}

// RegisterAPIEndpoints register API endpoints by registrar APIEndpointsFunc.
func (f APIEndpointsFunc) RegisterAPIEndpoints(router *gin.RouterGroup) {
	f(router)
}

// APIMiddleware middleware that should handle API requests.
type APIMiddleware interface {
	// ApplyToAPI handle API request by middleware.
	ApplyToAPI(c *gin.Context)
}

// APIMiddlewareFunc wrapping type for function to mark it as implementation of APIMiddleware.
type APIMiddlewareFunc func(c *gin.Context)

// ApplyToAPI handle API request by middleware function.
func (f APIMiddlewareFunc) ApplyToAPI(c *gin.Context) {
	f(c)
}
