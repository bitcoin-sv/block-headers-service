package assert

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

// Equal return error if two given variables are not equal.
// Used mainly in tests.
func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper()
	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}

// NotEqual return error if two given variables are equal.
// Used mainly in tests.
func NotEqual[T comparable](t *testing.T, actual, expected T) {
	t.Helper()
	if actual == expected {
		t.Errorf("The same variables were received. Expected different")
	}
}

// NoError finishes test with t.Fatal when there is an error.
func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Doesn't expect to receive error, but get one %s", err)
	}
}

// IsError expects that error exists and its Error() is returning expected string.
func IsError(t *testing.T, err error, expected string) {
	t.Helper()
	if err == nil {
		t.Errorf("Expect to receive error %s but got nothing", expected)
	} else {
		assert.Equal(t, err.Error(), expected)
	}
}
