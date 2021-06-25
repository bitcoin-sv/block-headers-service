// Package validator allows simple boundary checking of user input
// either against a struct or by checking values using the provided
// fluent API.
//
// It stores the output of any errors found in a usable structure that lists
// the field names and all validation errors associated with it.
//
// This allows the output to be serialised and returned in 400 responses or equivalent
// with readable and useful messages.
//
// It is designed to be simple to use and comes with a series of built in validation
// functions. You can add your own simply by wrapping the provided ValidationFunc.
package validator

import (
	"fmt"
	"sort"
	"strings"
)

// Validator is an interface that can be implemented on a struct
// to run validation checks against its properties.
//
// This is designed to be used in http handlers or middleware where
// all requests can be checked for this behaviour and evaluated.
//
//  type MyStruct struct{
//    MyProp string
//    ShouldBeTrue bool
//  }
//
//  func(m *MyStruct) Validate() error{
//      return validator.New().
//          Validate("myProp", validator.Length(m.MyProp,10,20)).
//          Validate("shouldBeTrue", validator.Bool(m.ShouldBeTrue, true))
//  }
//
type Validator interface {
	Validate() ErrValidation
}

// ValidationFunc defines a simple function that can be wrapped
// and supplied with arguments.
//
// Typical usage is shown:
//  func Length(val string, min, max int) ValidationFunc {
//	    return func() error {
//		    if len(val) >= min && len(val) <= max {
//			    return nil
//		    }
//		    return fmt.Errorf(validateLength, val, min, max)
//	    }
//   }
type ValidationFunc func() error

// String satisfies the String interface and returns the underlying error
// string that is returned by evaluating the function.
func (v ValidationFunc) String() string {
	return v().Error()
}

// ErrValidation contains a list of field names and a list of errors
// found against each. This can then be converted for output to a user.
type ErrValidation map[string][]string

// New will create and return a new ErrValidation which can have Validate functions chained.
func New() ErrValidation {
	return map[string][]string{}
}

// Validate will log any errors found when evaluating the list of validation functions
// supplied to it.
func (e ErrValidation) Validate(field string, fns ...ValidationFunc) ErrValidation {
	out := make([]string, 0)
	for _, fn := range fns {
		if err := fn(); err != nil {
			out = append(out, err.Error())
		}
	}
	if len(out) > 0 {
		e[field] = out
	}
	return e
}

// Err will return nil if no errors are found, ie all validators return valid
// or ErrValidation if an error has been found.
func (e ErrValidation) Err() error {
	if len(e) > 0 {
		return e
	}
	return nil
}

// String implements the Stringer interface and
// will return a string based representation
// of any errors found.
func (e ErrValidation) String() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	errs := make([]string, 0)
	for k, vv := range e {
		errs = append(errs, fmt.Sprintf("[%s: %s]", k, strings.Join(vv, ", ")))
	}
	sort.Strings(errs)
	return strings.Join(errs, ", ")
}

// Error implements the Error interface and ensure that ErrValidation
// can be passed as an error as well and being printable.
func (e ErrValidation) Error() string {
	return e.String()
}

// BadRequest implements the err BadRequest behaviour
// from the https://github.com/theflyingcodr/lathos package.
func (e ErrValidation) BadRequest() bool {
	return true
}
