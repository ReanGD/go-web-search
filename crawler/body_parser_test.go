package crawler

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func helperIsHTML(name string, t *testing.T, content string, result bool) {
	Convey(name, t, func() {
		if result {
			So(isHTML([]byte(content)), ShouldBeTrue)
		} else {
			So(isHTML([]byte(content)), ShouldBeFalse)
		}
	})
}

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

func helperBodyToUTF8(filename string, contentType []string, out string) {
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

func TestBodyToUTF8(t *testing.T) {
	Convey("windows-1251 meta", t, func() {
		helperBodyToUTF8("windows_1251_meta.html", []string{}, `<html>
<head><meta http-equiv="Content-Type" content="text/html; charset=windows-1251"></head>
<body><div>Привет</div></body>
</html>`)
	})

	Convey("windows-1251 header", t, func() {
		helperBodyToUTF8("windows_1251.html", []string{"text/html;charset=windows-1251"}, `<html>
<head></head>
<body><div>Привет</div></body>
</html>`)
	})

	Convey("koi8-r header", t, func() {
		helperBodyToUTF8("koi8-r.html", []string{"text/html;charset=koi8-r"}, `<html>
<head></head>
<body><div>Привет</div></body>
</html>`)
	})

	Convey("koi8-r meta", t, func() {
		helperBodyToUTF8("koi8-r_meta.html", []string{}, `<html>
<head><meta http-equiv="Content-Type" content="text/html; charset=koi8-r"></head>
<body><div>Привет</div></body>
</html>`)
	})

	Convey("utf-8 header", t, func() {
		helperBodyToUTF8("utf-8.html", []string{"text/html;charset=utf-8"}, `<html>
<head></head>
<body><div>Привет</div></body>
</html>`)
	})

	Convey("utf-8 meta", t, func() {
		helperBodyToUTF8("utf-8_meta.html", []string{}, `<html>
<head><meta http-equiv="Content-Type" content="text/html; charset=utf-8"></head>
<body><div>Привет</div></body>
</html>`)
	})

	Convey("not found encoding", t, func() {
		content := `<html>
<head></head>
<body><div>Привет</div></body>
</html>`
		_, err := bodyToUTF8([]byte(content), []string{})
		So(err, ShouldEqual, ErrBodyParserEncodingNotFound)
	})

}
