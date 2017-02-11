package client

import (
	"log"
)

type LogLevel int

func (ll LogLevel) String() string {
	switch ll {
	case LogLevelQuiet:
		return "Quiet"
	case LogLevelErrors:
		return "Errors"
	case LogLevelInfo:
		return "Info"
	case LogLevelDebug:
		return "Debug"
	default:
		return "UNKNOWN-LOGLEVEL"
	}
}

const (
	// Possible values for logLevel, use SetLogLevel with one of these values to
	// enable the respective logging.
	// Default value is LogLevelQuiet
	LogLevelQuiet LogLevel = iota
	LogLevelErrors
	LogLevelInfo
	LogLevelDebug
)

type Logger struct {
	LogOutput *log.Logger
	LogLevel  LogLevel
}

// Set the log level to the specified value.
// Possible values are LogLevelQuiet, LogLevelErrors, LogLevelInfo, and
// LogLevelDebug
func (this *Logger) SetLogLevel(level LogLevel) {
	this.LogLevel = level
}

func (this *Logger) Log(msgLogLevel LogLevel, item interface{}) {
	if this.LogLevel >= msgLogLevel {
		this.LogOutput.Println(item)
	}
}

func (this *Logger) Logf(msgLogLevel LogLevel, msg string, parameters ...interface{}) {
	if this.LogLevel >= msgLogLevel {
		this.LogOutput.Printf(msg, parameters...)
	}
}
