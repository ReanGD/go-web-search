package crawler

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/net/html"

	. "github.com/smartystreets/goconvey/convey"
)

type errReader struct {
}

func (r errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("error")
}

// TestReadBody ...
func TestReadBody(t *testing.T) {
	Convey("gzip body", t, func() {
		msg := []byte("raw body")

		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		_, err := w.Write(msg)
		So(err, ShouldBeNil)
		err = w.Close()
		So(err, ShouldBeNil)

		body := bytes.NewReader(buf.Bytes())
		result, err := readBody("gzip", body)
		So(err, ShouldBeNil)
		So(string(result), ShouldEqual, string(msg))
	})

	Convey("gzip body open error", t, func() {
		body := bytes.NewReader([]byte("raw body"))
		_, err := readBody("gzip", body)
		So(err.Error(), ShouldEqual, ErrReadGZipResponse)
	})

	Convey("gzip body open error", t, func() {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		_, err := w.Write([]byte("raw body"))
		So(err, ShouldBeNil)
		err = w.Close()
		So(err, ShouldBeNil)

		body := bytes.NewReader(buf.Bytes()[:10])
		_, err = readBody("gzip", body)
		So(err.Error(), ShouldEqual, ErrReadGZipResponse)
	})

	Convey("raw body", t, func() {
		data := []byte("test body")
		body := bytes.NewReader(data)
		result, err := readBody("identity", body)
		So(err, ShouldBeNil)
		So(string(result), ShouldEqual, string(data))
	})

	Convey("raw body with empty content-encoding", t, func() {
		data := []byte("test body")
		body := bytes.NewReader(data)
		result, err := readBody("identity", body)
		So(err, ShouldBeNil)
		So(string(result), ShouldEqual, string(data))
	})

	Convey("raw body error", t, func() {
		_, err := readBody("identity", errReader{})
		So(err.Error(), ShouldEqual, ErrReadResponse)
	})

	Convey("unknown body error", t, func() {
		_, err := readBody("unknown", errReader{})
		So(err.Error(), ShouldEqual, ErrUnknownContentEncoding)
	})
}

func helperIsHTML(name string, t *testing.T, content string, result bool) {
	Convey(name, t, func() {
		if result {
			So(isHTML([]byte(content)), ShouldBeTrue)
		} else {
			So(isHTML([]byte(content)), ShouldBeFalse)
		}
	})
}

// TestIsHTML ...
func TestIsHTML(t *testing.T) {
	helperIsHTML("Empty content", t,
		``, false)

	Convey("Nil content", t, func() {
		So(isHTML(nil), ShouldBeFalse)
	})

	helperIsHTML("Normal HTML", t,
		`<html><body><div>text</div></body></html>`, true)

	helperIsHTML("Error token in HTML", t,
		`<html<body><div>text</div></body></html>`, false)

	helperIsHTML("Partial HTML", t,
		`<body><div>text</div></body>`, false)

	helperIsHTML("Normal with spaces at the beginning of HTML", t,
		strings.Repeat(" ", 1000)+`<html><body><div>text</div></body></html>`, true)

	helperIsHTML("Long spaces at the beginning of HTML", t,
		strings.Repeat(" ", 2000)+`<html><body><div>text</div></body></html>`, false)
}

// helperBodyToUTF8 ...
func helperBodyToUTF8(filename string, contentType string, out string) {
	base, err := os.Getwd()
	So(err, ShouldBeNil)

	path := filepath.Join(base, "../test/data/body_encoding/", filename)
	_, err = os.Stat(path)
	So(err, ShouldBeNil)

	content, err := ioutil.ReadFile(path)
	So(err, ShouldBeNil)

	reader, err := bodyToUTF8(content, contentType)
	So(err, ShouldBeNil)

	body, err := ioutil.ReadAll(reader)
	So(err, ShouldBeNil)

	So(string(body), ShouldEqual, out)
}

// TestBodyToUTF8 ...
func TestBodyToUTF8(t *testing.T) {
	Convey("windows-1251 meta", t, func() {
		helperBodyToUTF8("windows_1251_meta.html", "", `<html>
<head><meta http-equiv="Content-Type" content="text/html; charset=windows-1251"></head>
<body><div>Привет</div></body>
</html>`)
	})

	Convey("windows-1251 header", t, func() {
		helperBodyToUTF8("windows_1251.html", "text/html;charset=windows-1251", `<html>
<head></head>
<body><div>Привет</div></body>
</html>`)
	})

	Convey("koi8-r header", t, func() {
		helperBodyToUTF8("koi8-r.html", "text/html;charset=koi8-r", `<html>
<head></head>
<body><div>Привет</div></body>
</html>`)
	})

	Convey("koi8-r meta", t, func() {
		helperBodyToUTF8("koi8-r_meta.html", "", `<html>
<head><meta http-equiv="Content-Type" content="text/html; charset=koi8-r"></head>
<body><div>Привет</div></body>
</html>`)
	})

	Convey("utf-8 header", t, func() {
		helperBodyToUTF8("utf-8.html", "text/html;charset=utf-8", `<html>
<head></head>
<body><div>Привет</div></body>
</html>`)
	})

	Convey("utf-8 meta", t, func() {
		helperBodyToUTF8("utf-8_meta.html", "", `<html>
<head><meta http-equiv="Content-Type" content="text/html; charset=utf-8"></head>
<body><div>Привет</div></body>
</html>`)
	})

	Convey("not found encoding", t, func() {
		content := `<html>
<head></head>
<body><div>Привет</div></body>
</html>`
		_, err := bodyToUTF8([]byte(content), "")
		So(err.Error(), ShouldEqual, ErrEncodingNotFound)
	})

}

type IoWriterErr struct {
}

func (w IoWriterErr) Write(p []byte) (int, error) {
	return 0, errors.New("write error")
}

// TestMinification ...
func TestMinification(t *testing.T) {
	Convey("call test", t, func() {
		data := `"<html><head></head><body>test()data</body></html>"`
		node, err := html.Parse(bytes.NewReader([]byte(data)))
		So(err, ShouldBeNil)

		var buf bytes.Buffer
		err = bodyMinification(node, &buf)
		So(err, ShouldBeNil)
		So(string(buf.Bytes()), ShouldEqual, `<html><head></head><body>test data</body></html>`)
	})

	Convey("call with error buffer", t, func() {
		data := `"<html><head></head><body>test()data</body></html>"`
		node, err := html.Parse(bytes.NewReader([]byte(data)))
		So(err, ShouldBeNil)

		var buf IoWriterErr
		err = bodyMinification(node, buf)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrRenderHTML)
	})
}
