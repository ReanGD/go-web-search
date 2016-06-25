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
		So(werr.Level, ShouldEqual, ErrorLevel)
		So(werr.Error(), ShouldEqual, msg)
		So(len(werr.Fields), ShouldEqual, 1)
		So(werr.Fields["caller"], ShouldStartWith, `[errors_test.go:`)
		So(werr.Fields["caller"], ShouldEndWith, `werrors.helperErrorEx`)
	})

	Convey("Check Newlevel", t, func() {
		msg := "message"
		lvl := WarningLevel

		err := NewLevel(msg, lvl)
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Level, ShouldEqual, lvl)
		So(werr.Error(), ShouldEqual, msg)
		So(len(werr.Fields), ShouldEqual, 1)
	})

	Convey("Check NewDetails", t, func() {
		msg := "message"
		details := "details message"

		err := NewDetails(msg, errors.New(details))
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Level, ShouldEqual, ErrorLevel)
		So(werr.Error(), ShouldEqual, msg)
		So(len(werr.Fields), ShouldEqual, 2)
		So(werr.Fields["details"], ShouldEqual, details)
	})

	Convey("Check NewFields", t, func() {
		msg := "message"

		err := NewFields(msg, "field1", "field1 data", "field2", "field2 data")
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Level, ShouldEqual, ErrorLevel)
		So(werr.Error(), ShouldEqual, msg)
		So(len(werr.Fields), ShouldEqual, 3)
		So(werr.Fields["field1"], ShouldEqual, "field1 data")
		So(werr.Fields["field2"], ShouldEqual, "field2 data")
	})

	Convey("Check NewFields with an odd number of arguments", t, func() {
		msg := "message"

		err := NewFields(msg, "field1", "field1 data", "field err")
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Level, ShouldEqual, ErrorLevel)
		So(werr.Error(), ShouldEqual, msg)
		So(len(werr.Fields), ShouldEqual, 2)
		So(werr.Fields["field1"], ShouldEqual, "field1 data")
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
		So(werr.Level, ShouldEqual, ErrorLevel)
		So(werr.Error(), ShouldEqual, msg)
		So(werr.Fields, ShouldResemble, map[string]string{"caller": ErrFailedGetCaller})
	})
}
