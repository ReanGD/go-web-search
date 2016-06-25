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
	msg    string
	Level  LogLevel
	Fields map[string]string
}

func (e ErrorEx) Error() string {
	return e.msg
}

var callerSkip = 2

func newImpl(logLevel LogLevel, msg string, fields ...string) error {
	var caller string
	ps, file, line, ok := runtime.Caller(callerSkip)
	if !ok {
		caller = ErrFailedGetCaller
	} else {
		caller = fmt.Sprintf("[%s:%d] %s", filepath.Base(file), line, runtime.FuncForPC(ps).Name())
	}

	result := &ErrorEx{
		msg:    msg,
		Level:  logLevel,
		Fields: make(map[string]string, 1+len(fields)/2)}
	result.Fields["caller"] = caller
	for i := 0; i != len(fields)/2; i++ {
		result.Fields[fields[i*2]] = fields[i*2+1]
	}

	return result
}

// New - create error with caller only
func New(msg string) error {
	return newImpl(ErrorLevel, msg)
}

// NewLevel - create error with caller and log level
func NewLevel(msg string, logLevel LogLevel) error {
	return newImpl(logLevel, msg)
}

// NewDetails - create error with caller and details message
func NewDetails(msg string, details error) error {
	return newImpl(ErrorLevel, msg, "details", details.Error())
}

// NewFields - create error with custom fields
func NewFields(msg string, fields ...string) error {
	return newImpl(ErrorLevel, msg, fields...)
}
