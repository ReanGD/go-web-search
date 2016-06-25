package database

import (
	"bytes"
	"compress/zlib"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestCompressedValue ...
func TestCompressedValue(t *testing.T) {
	Convey("Check nil value", t, func() {
		val := Compressed{Data: nil}
		result, err := val.Value()
		So(result, ShouldBeNil)
		So(err, ShouldBeNil)
	})

	Convey("Check empty value", t, func() {
		val := Compressed{Data: []byte("")}
		result, err := val.Value()
		So(result, ShouldBeNil)
		So(err, ShouldBeNil)
	})

	Convey("Check not nil value", t, func() {
		msg := []byte("text text text")

		var zContent bytes.Buffer
		w, _ := zlib.NewWriterLevelDict(&zContent, 6, nil)
		_, err := w.Write(msg)
		So(err, ShouldBeNil)
		err = w.Close()
		So(err, ShouldBeNil)

		val := Compressed{Data: msg}
		result, err := val.Value()
		So(result, ShouldResemble, zContent.Bytes())
		So(err, ShouldBeNil)
	})

}

// TestCompressedScan ...
func TestCompressedScan(t *testing.T) {
	Convey("Check nil value", t, func() {
		val := Compressed{Data: []byte("text")}
		err := val.Scan(nil)
		So(err, ShouldBeNil)
		So(val.Data, ShouldBeNil)
	})

	Convey("Check not byte value", t, func() {
		val := Compressed{}
		err := val.Scan("str")
		So(err.Error(), ShouldEqual, ErrScanArgument)
	})

	Convey("Check not compressed value", t, func() {
		val := Compressed{}
		err := val.Scan([]byte("test test test"))
		So(err.Error(), ShouldEqual, "Compressed.Scan: zlib: invalid header")
	})

	Convey("Check part of compressed value", t, func() {
		msg := []byte("test test test")
		actual := Compressed{Data: msg}
		compressed := actual.Compress()

		val := Compressed{}
		err := val.Scan(compressed[:10])
		So(err.Error(), ShouldEqual, "Compressed.Scan: unexpected EOF")
	})

	Convey("Check valid value", t, func() {
		msg := []byte("test test test")
		actual := Compressed{Data: msg}
		compressed := actual.Compress()

		val := Compressed{Data: []byte("text")}
		err := val.Scan(compressed)
		So(err, ShouldBeNil)
		So(val.Data, ShouldResemble, msg)
	})
}
