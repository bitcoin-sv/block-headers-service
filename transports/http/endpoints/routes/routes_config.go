package router

import "github.com/gin-gonic/gin"

// RootEndpointsFunc wrapping type for function to mark it as implementation of RootEndpoints.
type RootEndpointsFunc func(router *gin.RouterGroup)

// RootEndpoints registrar which will register routes in root context of application.
type RootEndpoints interface {
	// RegisterEndpoints register root endpoints.
	RegisterEndpoints(router *gin.RouterGroup)
}

// ApiEndpointsFunc wrapping type for function to mark it as implementation of ApiEndpoints.
type ApiEndpointsFunc func(router *gin.RouterGroup)

// ApiEndpoints registrar which will register routes in API routes group.
type ApiEndpoints interface {
	// RegisterApiEndpoints register API endpoints.
	RegisterApiEndpoints(router *gin.RouterGroup)
}

// RegisterEndpoints register root endpoints by registrar RootEndpointsFunc.
func (f RootEndpointsFunc) RegisterEndpoints(router *gin.RouterGroup) {
	f(router)
}

// RegisterApiEndpoints register API endpoints by registrar ApiEndpointsFunc.
func (f ApiEndpointsFunc) RegisterApiEndpoints(router *gin.RouterGroup) {
	f(router)
}

// ApiMiddleware middleware that should handle API requests.
type ApiMiddleware interface {
	//ApplyToApi handle API request by middleware.
	ApplyToApi(c *gin.Context)
}

// ApiMiddlewareFunc wrapping type for function to mark it as implementation of ApiMiddleware.
type ApiMiddlewareFunc func(c *gin.Context)

// ApplyToApi handle API request by middleware function.
func (f ApiMiddlewareFunc) ApplyToApi(c *gin.Context) {
	f(c)
}
