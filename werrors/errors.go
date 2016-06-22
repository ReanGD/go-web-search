package werrors

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// ErrorWithCaller - error with caller field
type ErrorWithCaller struct {
	msg    string
	Caller string
}

func (e *ErrorWithCaller) Error() string {
	return e.msg
}

// NewCaller - create ErrorWithCaller
func NewCaller(msg string) error {
	ps, file, line, ok := runtime.Caller(1)
	if !ok {
		return &ErrorWithCaller{
			msg:    msg,
			Caller: "failed to get caller",
		}
	}

	return &ErrorWithCaller{
		msg:    msg,
		Caller: fmt.Sprintf("[%s:%d] %s \n", filepath.Base(file), line, runtime.FuncForPC(ps).Name()),
	}
}
