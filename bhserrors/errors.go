package bhserrors

import (
	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// ExtendedError is an interface for errors that hold information about http status and code that should be returned
type ExtendedError interface {
	error
	GetCode() string
	GetMessage() string
	GetStatusCode() int
	StackTrace() (trace errors.StackTrace)
}

// BHSError is extended error which holds information about http status and code that should be returned from Block Header Service
type BHSError struct {
	Code       string
	Message    string
	StatusCode int
	cause      error
}

// responseError is an error which will be returned in HTTP response
type responseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error returns the error message string for BHSError, satisfying the error interface
func (e BHSError) Error() string {
	return e.Message
}

// GetCode returns the error code string for BHSError
func (e BHSError) GetCode() string {
	return e.Code
}

// GetMessage returns the error message string for BHSError
func (e BHSError) GetMessage() string {
	return e.Message
}

// GetStatusCode returns the error status code for BHSError
func (e BHSError) GetStatusCode() int {
	return e.StatusCode
}

// StackTrace returns the error's stack trace.
func (e BHSError) StackTrace() errors.StackTrace {
	err, ok := e.cause.(stackTracer)
	if !ok {
		return nil
	}

	return err.StackTrace()
}

// Unwrap returns the "cause" error
func (e BHSError) Unwrap() error {
	return e.cause
}

// Wrap sets the "cause" error
func (e BHSError) Wrap(err error) BHSError {
	e.cause = err
	return e
}

// WithTrace save the stack trace of the error
func (e BHSError) WithTrace(err error) BHSError {
	if st := stackTracer(nil); !errors.As(e.cause, &st) {
		return e.Wrap(errors.WithStack(err))
	}
	return e.Wrap(err)
}

// Is checks if the target error is the same as the current error
func (e BHSError) Is(target error) bool {
	t, ok := target.(BHSError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}
