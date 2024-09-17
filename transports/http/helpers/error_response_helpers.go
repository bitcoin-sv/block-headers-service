package helpers

import (
	"errors"
	"net/http"

	"github.com/bitcoin-sv/block-headers-service/domains"
)

// ResponseError is an object that we can return to the client if any error happens
// on the server
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is creating an error object to send it to the client
func ErrorResponse(err error, statusCode int) ResponseError {
	return parseError(err, statusCode)
}

// ErrorResponseFromMessage is simplified ErrorResponse when we don't want to create new error
// in code but just pass the message that will be sent to the client.
func ErrorResponseFromMessage(errMessage string, statusCode int) ResponseError {
	return parseError(errors.New(errMessage), statusCode)
}

// GetCodeFromError returns error code that should be returned to the client
// in a response based on the error message
func GetCodeFromError(err error) int {
	errorMessage := err.Error()

	switch errorMessage {
	case domains.MerklerootNotFoundError:
		return http.StatusNotFound
	case domains.MerklerootNotInLongestChainError:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func parseError(err error, statusCode int) ResponseError {
	return ResponseError{
		Message: err.Error(),
		Code:    statusCode,
	}
}
