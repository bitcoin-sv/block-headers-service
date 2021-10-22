package errs

import (
	"fmt"

	"github.com/google/uuid"
)

// ErrInternal implements InternalError and can be used
// to create server errors. You can also implement your own version
// by implementing the methods on the InternalError interface.
type ErrInternal struct {
	id       string
	// err is the original error that triggered the error
	err error
	code string
	message  string
	stack    string
	metadata map[string]interface{}
}

// NewErrInternal will create and return a new ErrInternal.
// This implementation will print a stack trace using the %+v verb
// so assumes you are using the /pkg/errors library to wrap
// your errors.
// You can implement your own.
func NewErrInternal(err error, code string) *ErrInternal {
	return &ErrInternal{
		id:       uuid.New().String(),
		message:  err.Error(),
		err: err,
		code: code,
		stack:    fmt.Sprintf("%+v", err),
		metadata: make(map[string]interface{}),
	}
}

// AddField assumes the underlying metadata map has been created and appends fields to it
// in a fluent manner.
//   internalError.AddField("key","my value").AddField("number",1234)
func (e *ErrInternal) AddField(key string, value interface{}) *ErrInternal{
	e.metadata[key] = value
	return e
}

// ID returns the ID for this instance of an error.
func (e ErrInternal) ID() string {
	return e.id
}

// Message will return the human readable message.
func (e ErrInternal) Message() string {
	return e.message
}

// Stack will return the stacktrace.
func (e ErrInternal) Stack() string {
	return e.stack
}

// Metadata is a data bag and can contain headers,
// method, status code, uri etc.
func (e ErrInternal) Metadata() map[string]interface{} {
	return e.metadata
}

// Error implements the error interface.
func (e ErrInternal) Error() string {
	return e.message
}

// Code returns the error code if there is one.
func (e ErrInternal) Code() string {
	return e.code
}
