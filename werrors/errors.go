package werrors

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// ErrFailedGetCaller - error get caller
const ErrFailedGetCaller = "Failed to get caller"

// ErrorWithCaller - error with caller field
type ErrorWithCaller struct {
	msg    string
	Caller string
}

func (e ErrorWithCaller) Error() string {
	return e.msg
}

var callerSkip = 1

// NewCaller - create ErrorWithCaller
func NewCaller(msg string) error {
	ps, file, line, ok := runtime.Caller(callerSkip)
	if !ok {
		return &ErrorWithCaller{
			msg:    msg,
			Caller: ErrFailedGetCaller,
		}
	}

	return &ErrorWithCaller{
		msg:    msg,
		Caller: fmt.Sprintf("[%s:%d] %s", filepath.Base(file), line, runtime.FuncForPC(ps).Name()),
	}
}
