package werrors

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/uber-go/zap"
)

func helperGetCaller() string {
	return func() string {
		return func() string {
			return getCaller()
		}()
	}()
}

// TestErrorEx ...
func TestErrorEx(t *testing.T) {
	Convey("Check Caller", t, func() {
		caller := helperGetCaller()
		So(caller, ShouldStartWith, `[errors_test.go:`)
		So(caller, ShouldEndWith, `werrors.helperGetCaller`)
	})

	Convey("Check stack error", t, func() {
		originalSkip := callerSkip
		callerSkip = 1e3
		defer func() { callerSkip = originalSkip }()

		So(helperGetCaller(), ShouldEqual, ErrFailedGetCaller)
	})

	Convey("Check NewEx", t, func() {
		callerTest = true
		defer func() { callerTest = false }()

		msg := "message"
		lvl := zap.InfoLevel

		err := NewEx(lvl, msg, zap.String("field1", "field1 data"), zap.Int("field2", 5))
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Level, ShouldEqual, lvl)
		So(werr.Error(), ShouldEqual, msg)
		So(werr.Fields, ShouldResemble,
			[]zap.Field{zap.String("caller", "<fake>"), zap.String("field1", "field1 data"), zap.Int("field2", 5)})
	})

	Convey("Check New", t, func() {
		callerTest = true
		defer func() { callerTest = false }()

		msg := "message"

		err := New(msg)
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Level, ShouldEqual, zap.ErrorLevel)
		So(werr.Error(), ShouldEqual, msg)
		So(werr.Fields, ShouldResemble, []zap.Field{zap.String("caller", "<fake>")})
	})

	Convey("Check Newlevel", t, func() {
		callerTest = true
		defer func() { callerTest = false }()

		msg := "message"
		lvl := zap.WarnLevel

		err := NewLevel(lvl, msg)
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Level, ShouldEqual, lvl)
		So(werr.Error(), ShouldEqual, msg)
		So(werr.Fields, ShouldResemble, []zap.Field{zap.String("caller", "<fake>")})
	})

	Convey("Check NewDetails", t, func() {
		callerTest = true
		defer func() { callerTest = false }()

		msg := "message"
		details := "details message"

		err := NewDetails(msg, errors.New(details))
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Level, ShouldEqual, zap.ErrorLevel)
		So(werr.Error(), ShouldEqual, msg)
		So(werr.Fields, ShouldResemble,
			[]zap.Field{zap.String("caller", "<fake>"), zap.String("details", details)})
	})

	Convey("Check NewFields", t, func() {
		callerTest = true
		defer func() { callerTest = false }()

		msg := "message"

		err := NewFields(msg, zap.String("field1", "field1 data"), zap.Int("field2", 5))
		So(err.Error(), ShouldEqual, msg)

		werr, ok := err.(*ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Level, ShouldEqual, zap.ErrorLevel)
		So(werr.Error(), ShouldEqual, msg)
		So(werr.Fields, ShouldResemble,
			[]zap.Field{zap.String("caller", "<fake>"), zap.String("field1", "field1 data"), zap.Int("field2", 5)})
	})
}
