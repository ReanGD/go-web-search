package crawler

import (
	"net/http"
	"testing"

	"github.com/ReanGD/go-web-search/werrors"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/uber-go/zap"
)

// TestCheckStatusCode ...
func TestCheckStatusCode(t *testing.T) {
	Convey("success", t, func() {
		So(checkStatusCode(200), ShouldBeNil)
	})

	Convey("error result", t, func() {
		err := checkStatusCode(300)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrStatusCode)
		werr, ok := err.(*werrors.ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Level, ShouldEqual, zap.ErrorLevel)
	})

	Convey("status 401", t, func() {
		err := checkStatusCode(401)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrStatusCode)
		werr, ok := err.(*werrors.ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Level, ShouldEqual, zap.WarnLevel)
	})

	Convey("status 404", t, func() {
		err := checkStatusCode(404)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrStatusCode)
		werr, ok := err.(*werrors.ErrorEx)
		So(ok, ShouldBeTrue)
		So(werr.Level, ShouldEqual, zap.WarnLevel)
	})
}

// TestCheckContentType ...
func TestCheckContentType(t *testing.T) {
	Convey("not found content-type", t, func() {
		header := &http.Header{"Content-Type-Err": []string{"val2"}}
		_, err := checkContentType(header)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrNotFountContentType)
	})

	Convey("wrong content-type", t, func() {
		header := &http.Header{"Content-Type": []string{"error value"}}
		_, err := checkContentType(header)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrParseContentType)
	})

	Convey("content-type not text/html", t, func() {
		header := &http.Header{"Content-Type": []string{"application/pdf; charset=UTF-8"}}
		_, err := checkContentType(header)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, InfoUnsupportedMimeFormat)
	})

	Convey("correct content-type", t, func() {
		contentType := "text/html; charset=UTF-8"
		header := &http.Header{"Content-Type": []string{contentType}}
		contentTypeResult, err := checkContentType(header)
		So(err, ShouldBeNil)
		So(contentTypeResult, ShouldEqual, contentType)
	})
}

// TestGetContentEncoding ...
func TestGetContentEncoding(t *testing.T) {
	Convey("one value", t, func() {
		header := &http.Header{
			"Content-Encoding1": []string{"val0", "val1"},
			"Content-Encoding":  []string{"val2"},
			"Content-Encoding2": []string{"val3", "val4"}}
		So(getContentEncoding(header), ShouldEqual, "val2")
	})

	Convey("two value", t, func() {
		header := &http.Header{
			"Content-Encoding1": []string{"val0", "val1"},
			"Content-Encoding":  []string{"val2", "val3"},
			"Content-Encoding2": []string{"val4", "val5"}}
		So(getContentEncoding(header), ShouldEqual, "val2")
	})

	Convey("empty header", t, func() {
		header := &http.Header{}
		So(getContentEncoding(header), ShouldEqual, "")
	})
}
