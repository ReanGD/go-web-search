package werrors

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/uber-go/zap"
)

// LogLevel level for logging
type LogLevel int32

// ErrFailedGetCaller - error get caller
const ErrFailedGetCaller = "Failed to get caller"

// ErrorEx - error with extended fields
type ErrorEx struct {
	msg    string
	Level  zap.Level
	Fields []zap.Field
}

func (e ErrorEx) Error() string {
	return e.msg
}

var callerSkip = 3
var callerTest = false

func getCaller() string {
	if callerTest {
		return "<fake>"
	}
	ps, file, line, ok := runtime.Caller(callerSkip)
	if !ok {
		return ErrFailedGetCaller
	}

	return fmt.Sprintf("[%s:%d] %s", filepath.Base(file), line, runtime.FuncForPC(ps).Name())
}

func newError(logLevel zap.Level, msg string, fields ...zap.Field) error {
	result := &ErrorEx{
		msg:    msg,
		Level:  logLevel,
		Fields: make([]zap.Field, len(fields)+1)}

	result.Fields[0] = zap.String("caller", getCaller())
	for i, field := range fields {
		result.Fields[i+1] = field
	}

	return result
}

// NewEx - create error with all fields
func NewEx(logLevel zap.Level, msg string, fields ...zap.Field) error {
	return newError(logLevel, msg, fields...)
}

// New - create error with caller only
func New(msg string) error {
	return newError(zap.ErrorLevel, msg)
}

// NewLevel - create error with caller and log level
func NewLevel(logLevel zap.Level, msg string) error {
	return newError(logLevel, msg)
}

// NewDetails - create error with caller and details message
func NewDetails(msg string, details error) error {
	return newError(zap.ErrorLevel, msg, zap.String("details", details.Error()))
}

// NewFields - create error with custom fields
func NewFields(msg string, fields ...zap.Field) error {
	return newError(zap.ErrorLevel, msg, fields...)
}

// LogError - write error to zap log
func LogError(logger zap.Logger, err error) {
	werr, ok := err.(*ErrorEx)
	if !ok {
		logger.Error(err.Error())
	} else {
		logger.Log(werr.Level, werr.Error(), werr.Fields...)
	}
}
