package werrors

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func helperErrorWithCaller(msg string) error {
	return NewCaller(msg)
}

// TestErrorWithCaller ...
func TestErrorWithCaller(t *testing.T) {
	Convey("Check Caller", t, func() {
		msg := "message"

		err := helperErrorWithCaller(msg)
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorWithCaller)
		So(ok, ShouldBeTrue)
		So(werr.Error(), ShouldEqual, msg)
		So(werr.Caller, ShouldStartWith, `[errors_test.go:`)
		So(werr.Caller, ShouldEndWith, `werrors.helperErrorWithCaller`)
	})

	Convey("Check stack error", t, func() {
		msg := "message"

		originalSkip := callerSkip
		callerSkip = 1e3
		defer func() { callerSkip = originalSkip }()

		err := helperErrorWithCaller(msg)
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorWithCaller)
		So(ok, ShouldBeTrue)
		So(werr.Error(), ShouldEqual, msg)
		So(werr.Caller, ShouldEqual, ErrFailedGetCaller)
	})
}
