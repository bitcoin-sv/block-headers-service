package logging

// Level is the level at which a logger is configured.  All messages sent
// to a level which is below the current level are filtered.
type Level uint32

// Level constants.
const (
	Trace Level = iota
	Debug
	Info
	Warn
	Error
	Critical
	Off
)

type (
	// TraceLogger logger used to log Trace messages.
	TraceLogger interface {
		// Trace formats message using the default formats for its operands
		// and writes to log with LevelTrace.
		Trace(v ...interface{})

		// Tracef formats message according to format specifier and writes
		// to log with LevelTrace.
		Tracef(format string, params ...interface{})
	}

	// DebugLogger logger used to log Debug messages.
	DebugLogger interface {
		// Debug formats message using the default formats for its operands
		// and writes to log with LevelDebug.
		Debug(v ...interface{})

		// Debugf formats message according to format specifier and writes to
		// log with LevelDebug.
		Debugf(format string, params ...interface{})
	}

	// InfoLogger logger used to log Info messages.
	InfoLogger interface {
		// Infof formats message according to format specifier and writes to
		// log with LevelInfo.
		Infof(format string, params ...interface{})

		// Info formats message using the default formats for its operands
		// and writes to log with LevelInfo.
		Info(v ...interface{})
	}

	// WarnLogger logger used to log Warn messages.
	WarnLogger interface {
		// Warnf formats message according to format specifier and writes to
		// log with LevelWarn.
		Warnf(format string, params ...interface{})

		// Warn formats message using the default formats for its operands
		// and writes to log with LevelWarn.
		Warn(v ...interface{})
	}

	// ErrorLogger logger used to log Error messages.
	ErrorLogger interface {
		// Errorf formats message according to format specifier and writes to
		// log with LevelError.
		Errorf(format string, params ...interface{})

		// Error formats message using the default formats for its operands
		// and writes to log with LevelError.
		Error(v ...interface{})
	}

	// CriticalLogger logger used to log Critical messages.
	CriticalLogger interface {
		// Criticalf formats message according to format specifier and writes to
		// log with LevelCritical.
		Criticalf(format string, params ...interface{})

		// Critical formats message using the default formats for its operands
		// and writes to log with LevelCritical.
		Critical(v ...interface{})
	}

	// CurrentLevelGetter component returning information about current logger Level.
	CurrentLevelGetter interface {
		// Level returns the current logger Level.
		Level() Level
	}

	// CurrentLevelSetter component that allows setting logger Level.
	CurrentLevelSetter interface {
		// SetLevel changes the logger level to the passed level.
		SetLevel(level Level)
	}

	// Logger is an interface which describes a level-based logger.
	Logger interface {
		TraceLogger
		DebugLogger
		InfoLogger
		WarnLogger
		ErrorLogger
		CriticalLogger
	}

	// LoggerWithLevel is a Logger with exposed information about current logger level.
	LoggerWithLevel interface {
		Logger
		CurrentLevelGetter
	}

	// ConfigurableLogger is an interface which describes a level-based logger that exposes method to set current logger level.
	ConfigurableLogger interface {
		LoggerWithLevel
		CurrentLevelSetter
	}

	// LoggerFactory used to create a logger.
	LoggerFactory interface {
		// NewLogger creates logger with component name.
		NewLogger(name string) Logger
		CurrentLevelGetter
		CurrentLevelSetter
	}
)
