package werrors

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// LogLevel level for logging
type LogLevel int32

// ErrFailedGetCaller - error get caller
const ErrFailedGetCaller = "Failed to get caller"

const (
	// DebugLevel - debug level
	DebugLevel LogLevel = iota - 1
	// InfoLevel - info level
	InfoLevel
	// WarningLevel - warning level
	WarningLevel
	// ErrorLevel - error level
	ErrorLevel
)

// ErrorEx - error with extended fields
type ErrorEx struct {
	msg     string
	Caller  string
	Details string
	Level   LogLevel
}

func (e ErrorEx) Error() string {
	return e.msg
}

var callerSkip = 1

// NewFull - ErrorWithCaller with all params
func NewFull(skip int, msg string, details string, logLevel LogLevel) error {
	ps, file, line, ok := runtime.Caller(callerSkip + skip)
	if !ok {
		return &ErrorEx{
			msg:     msg,
			Caller:  ErrFailedGetCaller,
			Details: details,
			Level:   logLevel,
		}
	}

	return &ErrorEx{
		msg:     msg,
		Caller:  fmt.Sprintf("[%s:%d] %s", filepath.Base(file), line, runtime.FuncForPC(ps).Name()),
		Details: details,
		Level:   logLevel,
	}
}

// New - create error with caller only
func New(msg string) error {
	return NewFull(1, msg, "", ErrorLevel)
}

// NewLevel - create error with caller and log level
func NewLevel(msg string, logLevel LogLevel) error {
	return NewFull(1, msg, "", logLevel)
}

// NewDetails - create error with caller and details message
func NewDetails(msg string, details error) error {
	return NewFull(1, msg, details.Error(), ErrorLevel)
}
