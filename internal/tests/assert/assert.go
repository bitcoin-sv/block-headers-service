package assert

import "testing"

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
