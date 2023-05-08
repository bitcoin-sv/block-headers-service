package wait

import (
	"errors"
	"fmt"
	"time"
)

// TimesOut error representing that waiting time has times out.
var TimesOut = errors.New("wait timeout")

// ForString wait for chanel to return string or finish with error when the time is out.
func ForString(sChan <-chan string, timeout time.Duration) (string, error) {
	select {
	case <-time.After(timeout):
		return "", fmt.Errorf("%w after %s", TimesOut, timeout)
	case v := <-sChan:
		return v, nil
	}
}
