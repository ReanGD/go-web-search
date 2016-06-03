package crawler

import (
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
