package validator

import (
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	reUKPostCode = regexp.MustCompile(`^[a-zA-Z]{1,2}\d[a-zA-Z\d]?\s*\d[a-zA-Z]{2}$`)
	reZipCode    = regexp.MustCompile(`^(\d{5}(?:\-\d{4})?)$`)
)

const (
	validateEmpty      = "value cannot be empty"
	validateLength     = "value must be between %d and %d characters"
	validateExactLength     = "value should be exactly %d characters"
	validateMin        = "value %d is smaller than minimum %d"
	validateMax        = "value %d is larger than maximum %d"
	validateNumBetween = "value %d must be between %d and %d"
	validatePositive   = "value %d should be greater than 0"
	validateRegex      = "value %s failed to meet requirements"
	validateBool       = "value %v does not evaluate to %v"
	validateDateEqual  = "the date/time provided %s, does not match the expected %s"
	validateDateAfter  = "the date provided %s, must be after %s"
	validateDateBefore = "the date provided %s, must be before %s"
	validateUkPostCode = "%s is not a valid UK PostCode"
	validateIsNumeric  = "string %s is not a number"
)

// StrLength will ensure a string, val, has a length that is at least min and
// at most max.
func StrLength(val string, min, max int) ValidationFunc {
	return func() error {
		if len(val) >= min && len(val) <= max {
			return nil
		}
		return fmt.Errorf(validateLength, min, max)
	}
}

// StrLengthExact will ensure a string, val, is exactly length.
func StrLengthExact(val string,length int) ValidationFunc {
	return func() error {
		if len(val) == length {
			return nil
		}
		return fmt.Errorf(validateExactLength, length)
	}
}

// MinInt will ensure an Int, val, is at least min in value.
func MinInt(val, min int) ValidationFunc {
	return func() error {
		if val >= min {
			return nil
		}
		return fmt.Errorf(validateMin, val, min)
	}
}

// MaxInt will ensure an Int, val,  is at most Max in value.
func MaxInt(val, max int) ValidationFunc {
	return func() error {
		if val <= max {
			return nil
		}
		return fmt.Errorf(validateMax, val, max)
	}
}

// BetweenInt will ensure an int, val,  is at least min and at most max.
func BetweenInt(val, min, max int) ValidationFunc {
	return func() error {
		if val >= min && val <= max {
			return nil
		}
		return fmt.Errorf(validateNumBetween, val, min, max)
	}
}

// MinInt64 will ensure an Int64, val, is at least min in value.
func MinInt64(val, min int64) ValidationFunc {
	return func() error {
		if val >= min {
			return nil
		}
		return fmt.Errorf(validateMin, val, min)
	}
}

// MaxInt64 will ensure an Int64, val, is at most Max in value.
func MaxInt64(val, max int64) ValidationFunc {
	return func() error {
		if val <= max {
			return nil
		}
		return fmt.Errorf(validateMax, val, max)
	}
}

// BetweenInt64 will ensure an int64, val, is at least min and at most max.
func BetweenInt64(val, min, max int64) ValidationFunc {
	return func() error {
		if val >= min && val <= max {
			return nil
		}
		return fmt.Errorf(validateNumBetween, val, min, max)
	}
}

// MinUInt64 will ensure an uint64, val, is at least min in value.
func MinUInt64(val, min uint64) ValidationFunc {
	return func() error {
		if val >= min {
			return nil
		}
		return fmt.Errorf(validateMin, val, min)
	}
}

// MaxUInt64 will ensure an Int64, val, is at most Max in value.
func MaxUInt64(val, max uint64) ValidationFunc {
	return func() error {
		if val <= max {
			return nil
		}
		return fmt.Errorf(validateMax, val, max)
	}
}

// BetweenUInt64 will ensure an int64, val, is at least min and at most max.
func BetweenUInt64(val, min, max uint64) ValidationFunc {
	return func() error {
		if val >= min && val <= max {
			return nil
		}
		return fmt.Errorf(validateNumBetween, val, min, max)
	}
}

// PositiveInt will ensure an int, val, is > 0.
func PositiveInt(val int) ValidationFunc {
	return func() error {
		if val > 0 {
			return nil
		}
		return fmt.Errorf(validatePositive, val)
	}
}

// PositiveInt64 will ensure an int64, val, is > 0.
func PositiveInt64(val int64) ValidationFunc {
	return func() error {
		if val > 0 {
			return nil
		}
		return fmt.Errorf("value %d should be greater than 0", val)
	}
}

// PositiveUInt64 will ensure an uint64, val, is > 0.
func PositiveUInt64(val uint64) ValidationFunc {
	return func() error {
		if val > 0 {
			return nil
		}
		return fmt.Errorf("value %d should be greater than 0", val)
	}
}

// MatchString will check that a string, val, matches the provided regular expression.
func MatchString(val string, r *regexp.Regexp) ValidationFunc {
	return func() error {
		if r.MatchString(val) {
			return nil
		}
		return fmt.Errorf(validateRegex, val)
	}
}

// MatchBytes will check that a byte array, val, matches the provided regular expression.
func MatchBytes(val []byte, r *regexp.Regexp) ValidationFunc {
	return func() error {
		if r.Match(val) {
			return nil
		}
		return fmt.Errorf(validateRegex, val)
	}
}

// Bool is a simple check to ensure that val matches either true / false as defined by exp.
func Bool(val, exp bool) ValidationFunc {
	return func() error {
		if val == exp {
			return nil
		}
		return fmt.Errorf(validateBool, val, exp)
	}
}

// DateEqual will ensure that a date/time, val, matches exactly exp.
func DateEqual(val, exp time.Time) ValidationFunc {
	return func() error {
		if val.Equal(exp) {
			return nil
		}
		return fmt.Errorf(validateDateEqual, val, exp)
	}
}

// DateAfter will ensure that a date/time, val, occurs after exp.
func DateAfter(val, exp time.Time) ValidationFunc {
	return func() error {
		if val.After(exp) {
			return nil
		}
		return fmt.Errorf(validateDateAfter, val, exp)
	}
}

// DateBefore will ensure that a date/time, val, occurs before exp.
func DateBefore(val, exp time.Time) ValidationFunc {
	return func() error {
		if val.Before(exp) {
			return nil
		}
		return fmt.Errorf(validateDateBefore, val, exp)
	}
}

// NotEmpty will ensure that a value, val, is not empty.
// rules are:
// int: > 0
// string: != "" or whitespace
// slice: not nil and len > 0
// map: not nil and len > 0
func NotEmpty(v interface{}) ValidationFunc {
	return func() error {
		if v == nil {
			return fmt.Errorf(validateEmpty)
		}
		val := reflect.ValueOf(v)
		valid := false
		unknown := false
		// nolint:exhaustive // not supporting everything
		switch val.Kind() {
		case reflect.Array, reflect.Map, reflect.Slice:
			valid = val.Len() > 0 && !val.IsNil()
		case reflect.String:
			valid = len(strings.TrimSpace(val.String())) > 0
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			valid = val.Int() > 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			valid = val.Uint() > 0
		case reflect.Float32, reflect.Float64:
			valid = val.Float() > 0
		case reflect.Interface, reflect.Ptr:
			valid = !val.IsNil()
		default:
			unknown = true
		}
		if t, ok := v.(time.Time); ok {
			unknown = false
			valid = !t.IsZero()
		}
		// unknown type - panic - this should be false
		if unknown {
			panic(fmt.Sprintf("unsupported type %T", v))
		}
		if !valid {
			return fmt.Errorf(validateEmpty)
		}
		return nil
	}
}

// IsNumeric will pass if a string, val, is an Int.
func IsNumeric(val string) ValidationFunc {
	return func() error {
		_, err := strconv.Atoi(val)
		if err == nil {
			return nil
		}
		return fmt.Errorf(validateIsNumeric, val)
	}
}

// UKPostCode will validate that a string, val, is a valid UK PostCode.
// It does not check the postcode exists, just that it matches an agreed pattern.
func UKPostCode(val string) ValidationFunc {
	return func() error {
		if reUKPostCode.MatchString(val) {
			return nil
		}
		return fmt.Errorf(validateUkPostCode, val)
	}
}

// USZipCode will validate that a string, val, matches a US USZipCode pattern.
// It does not check the zipcode exists, just that it matches an agreed pattern.
func USZipCode(val string) ValidationFunc {
	return func() error {
		if reZipCode.MatchString(val) {
			return nil
		}
		return fmt.Errorf("%s is not a valid UK PostCode", val)
	}
}

// HasPrefix ensures string, val, has a prefix matching prefix.
func HasPrefix(val, prefix string) ValidationFunc {
	return func() error {
		if strings.HasPrefix(val, prefix) {
			return nil
		}
		return fmt.Errorf("value provided does not have a valid prefix")
	}
}

// NoPrefix ensures a string, val, does not have the supplied prefix.
func NoPrefix(val, prefix string) ValidationFunc {
	return func() error {
		if strings.HasPrefix(val, prefix) {
			return errors.New("value provided does not have a valid prefix")
		}
		return nil
	}
}

// IsHex will check that a string, val, is valid Hexadecimal.
func IsHex(val string) ValidationFunc {
	return func() error {
		if _, err := hex.DecodeString(val); err != nil {
			return errors.New("value supplied is not valid hex")
		}
		return nil
	}
}
