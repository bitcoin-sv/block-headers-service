package wait

import (
	"errors"
	"fmt"
	"time"
)

// ErrTimesOut error representing that waiting time has times out.
var ErrTimesOut = errors.New("wait timeout")

// ForString wait for chanel to return string or finish with error when the time is out.
func ForString(sChan <-chan string, timeout time.Duration) (string, error) {
	select {
	case <-time.After(timeout):
		return "", fmt.Errorf("%w after %s", ErrTimesOut, timeout)
	case v := <-sChan:
		return v, nil
	}
}
