package logger

import (
	"os"
	"strings"

	"github.com/libsv/bitcoin-hc/domains/logging"
	"github.com/libsv/bitcoin-hc/transports/p2p/p2plog"
)

// DefaultLoggerFactory creates default factory with default system tag, level and writing to std out.
func DefaultLoggerFactory() logging.LoggerFactory {
	return NewLoggerFactory("HEADERS", logging.Info, os.Stdout)
}

// SetLevelFromString sets logger level based on string.
// Defaults to Info if string doesn't match expected level string.
func SetLevelFromString(target interface{}, level string) {
	l, _ := ParseLevel(level)
	SetLevel(target, l)
}

// SetLevel tries to set a logging level.
// If target is logging.CurrentLevelSetter then it is setting a logging level and returning true,
// otherwise returning false.
func SetLevel(target interface{}, l logging.Level) (ok bool) {
	t, ok := target.(logging.CurrentLevelSetter)
	if ok {
		t.SetLevel(l)
	}
	return
}

// ToLevel utility to convert from p2plog.Level to logger.Level.
func ToLevel(level p2plog.Level) logging.Level {
	return logging.Level(level)
}

// ParseLevel returns a level based on the input string s.  If the input
// can't be interpreted as a valid log level, the info level and false is
// returned.
func ParseLevel(s string) (l logging.Level, ok bool) {
	switch strings.ToLower(s) {
	case "trace", "trc":
		return logging.Trace, true
	case "debug", "dbg":
		return logging.Debug, true
	case "info", "inf":
		return logging.Info, true
	case "warn", "wrn":
		return logging.Warn, true
	case "error", "err":
		return logging.Error, true
	case "critical", "crt":
		return logging.Critical, true
	case "off":
		return logging.Off, true
	default:
		return logging.Info, false
	}
}

// DirectionString returns string direction.
func DirectionString(inbound bool) string {
	if inbound {
		return "inbound"
	}
	return "outbound"
}

// LogClosure is a closure that can be printed with %v to be used to
// generate expensive-to-create data for a detailed log level and avoid doing
// the work if the data isn't printed.
type logClosure func() string

// String String() realization.
func (c logClosure) String() string {
	return c()
}

// NewLogClosure returns logClosure.
func NewLogClosure(c func() string) logClosure {
	return logClosure(c)
}
