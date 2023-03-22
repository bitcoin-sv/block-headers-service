package assert

import (
	"github.com/magiconair/properties/assert"
	"testing"
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

func NoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Doesn't expect to receive error, but get one %s", err)
	}
}

func IsError(t *testing.T, err error, expected string) {
	if err == nil {
		t.Error("Expect to receive error BlockRejected")
	} else {
		assert.Equal(t, err.Error(), expected)
	}
}
