package client

import (
	LOG "log"
	"os"
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

var (
	// Logger instance with prefix
	// Loglevel default is quiet = no logging
	infoLog  *LOG.Logger = LOG.New(os.Stdout, "Status-Monitor: ", LOG.Ldate|LOG.Ltime|LOG.Lshortfile)
	logLevel LogLevel    = LogLevelQuiet
)

// Set the log level to the specified value.
// Possible values are LogLevelQuiet, LogLevelErrors, LogLevelInfo, and
// LogLevelDebug
func SetLogLevel(level LogLevel) {
	logLevel = level
}

func log(msgLogLevel LogLevel, item interface{}) {
	if logLevel >= msgLogLevel {
		infoLog.Println(item)
	}
}

func logf(msgLogLevel LogLevel, msg string, parameters ...interface{}) {
	if logLevel >= msgLogLevel {
		infoLog.Printf(msg, parameters...)
	}
}
