package response

import (
	"errors"
	"net/http"

	"github.com/bitcoin-sv/block-headers-service/domains"
)

// ResponseError is an object that we can return to the client if any error happens
// on the server
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error is creating an error object to send it to the client
// it also returns http statusCode
func Error(err error) (*ResponseError, int) {
	sc, cm := getCodeFromError(err)
	return parseError(err, sc, cm)
}

// ErrorFromMessage is simplified ErrorResponse when we don't want to create new error
// in code but just pass the message that will be sent to the client.
func ErrorFromMessage(errMessage string) (*ResponseError, int) {
	sc, cm := getCodeFromError(errors.New(errMessage))
	return parseError(errors.New(errMessage), sc, cm)
}

// GetCodeFromError returns error code and code message that should be returned to the client
// in a response based on the error message
func getCodeFromError(err error) (int, string) {
	var errCode domains.ErrorCode = domains.ErrGeneric

	switch {
	case errors.Is(err, domains.ErrMerklerootNotFound):
		errCode = domains.ErrMerkleRootNotFound
		return http.StatusNotFound, errCode.String()
	case errors.Is(err, domains.ErrMerklerootNotInLongestChain):
		errCode = domains.ErrMerkleRootNotInLC
		return http.StatusConflict, errCode.String()
	case errors.Is(err, domains.ErrMerklerootInvalidBatchSize):
		errCode = domains.ErrInvalidBatchSize
		return http.StatusBadRequest, errCode.String()
	default:
		return http.StatusInternalServerError, errCode.String()
	}
}

func parseError(err error, statusCode int, code string) (*ResponseError, int) {
	return &ResponseError{
		Message: err.Error(),
		Code:    code,
	}, statusCode
}
