package logger

import (
	"fmt"

	"github.com/rs/zerolog"
)

// CustomZerolog zerolog adapter struct.
type CustomZerolog struct {
	entry zerolog.Logger
}

func newZerologAdapter(logger zerolog.Logger) *CustomZerolog {
	return &CustomZerolog{entry: logger}
}

// Trace formats message using the default formats for its operands, prepends
// the prefix as necessary, and writes to log with LevelTrace.
//
// This is part of the Logger interface implementation.
func (cl *CustomZerolog) Trace(v ...interface{}) {
	cl.entry.Trace().Msg(fmt.Sprint(v...))
}

// Tracef formats message according to format specifier, prepends the prefix as
// necessary, and writes to log with LevelTrace.
//
// This is part of the Logger interface implementation.
func (cl *CustomZerolog) Tracef(format string, params ...interface{}) {
	cl.entry.Trace().Msg(fmt.Sprintf(format, params...))
}

// Debug formats message using the default formats for its operands, prepends
// the prefix as necessary, and writes to log with LevelDebug.
//
// This is part of the Logger interface implementation.
func (cl *CustomZerolog) Debug(v ...interface{}) {
	cl.entry.Debug().Msg(fmt.Sprint(v...))
}

// Debugf formats message according to format specifier, prepends the prefix as
// necessary, and writes to log with LevelDebug.
//
// This is part of the Logger interface implementation.
func (cl *CustomZerolog) Debugf(format string, params ...interface{}) {
	cl.entry.Debug().Msg(fmt.Sprintf(format, params...))
}

// Info formats message using the default formats for its operands, prepends
// the prefix as necessary, and writes to log with LevelInfo.
//
// This is part of the Logger interface implementation.
func (cl *CustomZerolog) Info(v ...interface{}) {
	cl.entry.Info().Msg(fmt.Sprint(v...))
}

// Infof formats message according to format specifier, prepends the prefix as
// necessary, and writes to log with LevelInfo.
//
// This is part of the Logger interface implementation.
func (cl *CustomZerolog) Infof(format string, params ...interface{}) {
	cl.entry.Info().Msg(fmt.Sprintf(format, params...))
}

// Warn formats message using the default formats for its operands, prepends
// the prefix as necessary, and writes to log with LevelWarn.
//
// This is part of the Logger interface implementation.
func (cl *CustomZerolog) Warn(v ...interface{}) {
	cl.entry.Warn().Msg(fmt.Sprint(v...))
}

// Warnf formats message according to format specifier, prepends the prefix as
// necessary, and writes to log with LevelWarn.
//
// This is part of the Logger interface implementation.
func (cl *CustomZerolog) Warnf(format string, params ...interface{}) {
	cl.entry.Warn().Msg(fmt.Sprintf(format, params...))
}

// Error formats message using the default formats for its operands, prepends
// the prefix as necessary, and writes to log with LevelError.
//
// This is part of the Logger interface implementation.
func (cl *CustomZerolog) Error(v ...interface{}) {
	cl.entry.Error().Msg(fmt.Sprint(v...))
}

// Errorf formats message according to format specifier, prepends the prefix as
// necessary, and writes to log with LevelError.
//
// This is part of the Logger interface implementation.
func (cl *CustomZerolog) Errorf(format string, params ...interface{}) {
	cl.entry.Error().Msg(fmt.Sprintf(format, params...))
}

// Critical formats message using the default formats for its operands, prepends
// the prefix as necessary, and writes to log with LevelCritical.
//
// This is part of the Logger interface implementation.
func (cl *CustomZerolog) Critical(v ...interface{}) {
	cl.entry.Fatal().Msg(fmt.Sprint(v...))
}

// Criticalf formats message according to format specifier, prepends the prefix
// as necessary, and writes to log with LevelCritical.
//
// This is part of the Logger interface implementation.
func (cl *CustomZerolog) Criticalf(format string, params ...interface{}) {
	cl.entry.Fatal().Msg(fmt.Sprintf(format, params...))
}
