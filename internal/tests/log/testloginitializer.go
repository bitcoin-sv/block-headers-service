package testlog

import (
	"github.com/libsv/bitcoin-hc/app/logger"
	"github.com/libsv/bitcoin-hc/domains/logging"
	"github.com/libsv/bitcoin-hc/transports/p2p/p2plog"
	"os"
)

// InitializeMockLogger initialize logger for tests.
func InitializeMockLogger() p2plog.Logger {
	l := NewTestLoggerFactory().NewLogger("")
	log, ok := logger.UnwrapP2plog(l)
	if !ok {
		panic("expect to unwrap P2plog from logger")
	}
	return log
}

// NewTestLoggerFactory creates new logger factory for tests.
func NewTestLoggerFactory() logging.LoggerFactory {
	return logger.NewLoggerFactory("PULSE_TEST", logging.Debug, os.Stdout.Write)
}
