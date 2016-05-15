package crawler

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func strEq(name string, t *testing.T, in, out string) {
	Convey(name, t, func() {
		m := minification2{}
		So(m.processText(in), ShouldEqual, out)
	})
}

func TestSpaces(t *testing.T) {
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
