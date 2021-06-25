package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	validator "github.com/theflyingcodr/govalidator"
	"github.com/theflyingcodr/lathos"
	"github.com/theflyingcodr/lathos/errs"
)

// ErrorHandler we can flesh this out.
func ErrorHandler(err error, c echo.Context) {
	if err == nil {
		return
	}
	var valErr validator.ErrValidation
	if errors.As(err, &valErr) {
		resp := map[string]interface{}{
			"errors": valErr,
		}
		_ = c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Internal error, log it to a system and return small detail
	if !lathos.IsClientError(err) {
		log.Error(errs.NewErrInternal(err, nil))
		_ = c.String(http.StatusInternalServerError, err.Error())
		return
	}
	var clientErr lathos.ClientError
	errors.As(err, &clientErr)
	resp := struct {
		ID      string `json:"id"`
		Code    string `json:"code"`
		Title   string `json:"title"`
		Message string `json:"message"`
	}{
		ID:      clientErr.ID(),
		Code:    clientErr.Code(),
		Title:   clientErr.Title(),
		Message: clientErr.Detail(),
	}
	if lathos.IsNotFound(err) {
		_ = c.JSON(http.StatusNotFound, resp)
		return
	}
	if lathos.IsDuplicate(err) {
		_ = c.JSON(http.StatusConflict, resp)
		return
	}
	if lathos.IsNotAuthenticated(err) {
		_ = c.JSON(http.StatusUnauthorized, resp)
		return
	}
	if lathos.IsNotAuthorised(err) {
		_ = c.JSON(http.StatusForbidden, resp)
		return
	}
}
