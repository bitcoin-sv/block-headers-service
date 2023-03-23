package testlog

import (
	"os"

	"github.com/libsv/bitcoin-hc/transports/p2p/p2plog"
)

type logWriter struct{}

func (logWriter) Write(p []byte) (n int, err error) {
	_, err = os.Stdout.Write(p)
	if err != nil {
		return len(p), err
	}
	return len(p), nil
}

// InitializeMockLogger initialize logger for tests.
func InitializeMockLogger() p2plog.Logger {
	backendLog := p2plog.NewBackend(logWriter{})
	logger := backendLog.Logger("TEST_LOGGER")
	return logger
}
