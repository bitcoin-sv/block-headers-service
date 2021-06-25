package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/libsv/headers-client"
)

type header struct {
	svc headers.BlockheaderService
}

// NewHeader will setup a new headers http transport.
func NewHeader(svc headers.BlockheaderService) *header {
	return &header{svc: svc}
}

// Routes will setup the routes with the echo group.
func (h *header) Routes(g *echo.Group) {
	g.GET(urlHeader, h.Header)
	g.GET(urlHeight, h.Height)
}

// Header will return a header based on the blockhash provided.
func (h *header) Header(e echo.Context) error {
	var args headers.HeaderArgs
	if err := e.Bind(&args); err != nil {
		return err
	}
	resp, err := h.svc.Header(e.Request().Context(), args)
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusOK, resp)
}

// Height will return current block height.
func (h *header) Height(e echo.Context) error {
	resp, err := h.svc.Height(e.Request().Context())
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusOK, resp)
}
