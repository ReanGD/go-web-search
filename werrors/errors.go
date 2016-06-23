package werrors

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// ErrFailedGetCaller - error get caller
const ErrFailedGetCaller = "Failed to get caller"

// ErrorEx - error with extended fields
type ErrorEx struct {
	msg     string
	Caller  string
	Details string
}

func (e ErrorEx) Error() string {
	return e.msg
}

var callerSkip = 1

// NewFull - ErrorWithCaller with all params
func NewFull(skip int, msg string, details string) error {
	ps, file, line, ok := runtime.Caller(callerSkip + skip)
	if !ok {
		return &ErrorEx{
			msg:     msg,
			Caller:  ErrFailedGetCaller,
			Details: details,
		}
	}

	return &ErrorEx{
		msg:     msg,
		Caller:  fmt.Sprintf("[%s:%d] %s", filepath.Base(file), line, runtime.FuncForPC(ps).Name()),
		Details: details,
	}
}

// New - create error with caller only
func New(msg string) error {
	return NewFull(1, msg, "")
}

// NewDetails - create error with caller and details message
func NewDetails(msg string, details error) error {
	return NewFull(1, msg, details.Error())
}
