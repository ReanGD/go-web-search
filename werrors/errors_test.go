package werrors

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func helperErrorEx(msg string) error {
	return New(msg)
}

// TestErrorEx ...
func TestErrorEx(t *testing.T) {
	Convey("Check Caller", t, func() {
		msg := "message"

		err := helperErrorEx(msg)
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Error(), ShouldEqual, msg)
		So(werr.Caller, ShouldStartWith, `[errors_test.go:`)
		So(werr.Caller, ShouldEndWith, `werrors.helperErrorEx`)
		So(werr.Details, ShouldEqual, "")
		So(werr.Level, ShouldEqual, ErrorLevel)
	})

	Convey("Check Log level", t, func() {
		msg := "message"
		lvl := WarningLevel

		err := NewLevel(msg, lvl)
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Error(), ShouldEqual, msg)
		So(werr.Details, ShouldEqual, "")
		So(werr.Level, ShouldEqual, lvl)
	})

	Convey("Check Details", t, func() {
		msg := "message"
		details := "details"

		err := NewDetails(msg, errors.New(details))
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Error(), ShouldEqual, msg)
		So(werr.Details, ShouldEqual, details)
		So(werr.Level, ShouldEqual, ErrorLevel)
	})

	Convey("Check stack error", t, func() {
		msg := "message"

		originalSkip := callerSkip
		callerSkip = 1e3
		defer func() { callerSkip = originalSkip }()

		err := helperErrorEx(msg)
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Error(), ShouldEqual, msg)
		So(werr.Caller, ShouldEqual, ErrFailedGetCaller)
		So(werr.Details, ShouldEqual, "")
		So(werr.Level, ShouldEqual, ErrorLevel)
	})
}
