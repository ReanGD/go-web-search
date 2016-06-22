package crawler

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"golang.org/x/net/html"

	. "github.com/smartystreets/goconvey/convey"
)

func strEq(name string, t *testing.T, in, out string) {
	Convey(name, t, func() {
		m := minificationText{}
		So(m.processText(in), ShouldEqual, out)
	})
}

func minificationTextCheck(in string, out string) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	node, err := html.Parse(bytes.NewReader([]byte(in)))
	if err == nil {
		So(RunMinificationText(node), ShouldEqual, nil)
		err := html.Render(w, node)
		if err == nil {
			err = w.Flush()
			if err == nil {
				So(string(buf.Bytes()), ShouldEqual, out)
			}
		}
	}
	So(err, ShouldBeNil)
}

func helperTextBody(name string, t *testing.T, in, out string) {
	Convey(name, t, func() {
		fin := fmt.Sprintf("<html><head></head><body>\n%s\n</body></html>", in)
		fout := fmt.Sprintf("<html><head></head><body>%s</body></html>", out)
		minificationTextCheck(fin, fout)
	})
}

func TestMinimizeText(t *testing.T) {
	strEq("Empty string", t,
		"", "")

	strEq("Only spaces", t,
		"    ", "")

	strEq("Without spaces", t,
		"приветworld", "приветworld")

	strEq("One space", t,
		"привет world", "привет world")

	strEq("Left spaces", t,
		"  привет world", "привет world")

	strEq("Middle spaces", t,
		"привет   world", "привет world")

	strEq("Right spaces", t,
		"привет world  ", "привет world")

	strEq("Left, middle and right spaces", t,
		" привет    world ", "привет world")

	strEq("Convert separators to space",
		t, "...привет, \t\r\nworld!", "привет world")

	strEq("To lower case",
		t, "ПрИвЕт, WoRlD!", "привет world")

	strEq("Special symbols",
		t, "П&р-и@в_е+т, W'oRlD!", "п&р-и@в_е+т w'orld")

	strEq("Digits",
		t, "привет, 1234567890!", "привет 1234567890")

	strEq("Digits with comma",
		t, "12,5", "12,5")

	strEq("Digits with colon",
		t, "1:6", "1:6")

	strEq("Digits with dot 1",
		t, "123.45", "123.45")

	strEq("Digits with dot 2",
		t, "1230.045", "1230.045")

	strEq("Digits with dot 3",
		t, "1239.945", "1239.945")

	strEq("Dot after digits",
		t, "1239.", "1239")
}

func TestParseErrors(t *testing.T) {
	Convey("Test error node type", t, func() {
		in := "<html><head></head><body><!-- Comment1 --></body></html>"
		node, err := html.Parse(bytes.NewReader([]byte(in)))
		So(err, ShouldBeNil)

		So(RunMinificationText(node).Error(), ShouldEqual, ErrUnexpectedNodeType)
	})

	Convey("Test error node type", t, func() {
		in := "<html><head></head><body><b>123</b></body></html>"
		node, err := html.Parse(bytes.NewReader([]byte(in)))
		So(err, ShouldBeNil)

		So(RunMinificationText(node).Error(), ShouldEqual, ErrUnexpectedTag)
	})
}

func TestRemoveEmptyTags(t *testing.T) {
	helperTextBody("Remove empty tag between tags", t,
		"<div>pre</div><div></div><div>post</div>",
		"<div>pre</div><div>post</div>")

	helperTextBody("Remove empty tag between text nodes", t,
		"pre<div></div>post",
		"pre post")

	helperTextBody("Remove empty tag between tag and text node", t,
		"<div>pre</div><div></div>post",
		"<div>pre</div>post")

	helperTextBody("Remove empty tag between text node and tag", t,
		"pre<div></div><div>post</div>",
		"pre<div>post</div>")

	helperTextBody("Remove tag with spaces", t,
		"<div>pre</div><div>.\n\t,</div><div>post</div>",
		"<div>pre</div><div>post</div>")
}

func TestOpenTag(t *testing.T) {
	helperTextBody("Remove tag with spaces", t,
		"<div><div>text</div></div>",
		"text")
}
