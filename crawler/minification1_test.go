package crawler

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"golang.org/x/net/html"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	emptyHead = "<html><head> </head><body></body></html>"
)

func minificationCheck(in string, out string) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	node, err := html.Parse(bytes.NewReader([]byte(in)))
	if err == nil {
		So(RunMinification1(node), ShouldEqual, nil)
		err = html.Render(w, node)
		if err == nil {
			err = w.Flush()
			if err == nil {
				So(string(buf.Bytes()), ShouldEqual, out)
			}
		}
	}
	So(err, ShouldEqual, nil)
}

func HelperBody(name string, t *testing.T, in, out string) {
	Convey(name, t, func() {
		fin := fmt.Sprintf("<html><head></head><body>\n%s\n</body></html>", in)
		fout := fmt.Sprintf("<html><head></head><body>\n%s\n</body></html>", out)
		minificationCheck(fin, fout)
	})
}

func HelperDiv(name string, t *testing.T, in, out string) {
	Convey(name, t, func() {
		fin := fmt.Sprintf("<html><head></head><body>\n<div>%s</div>\n</body></html>", in)
		fout := fmt.Sprintf("<html><head></head><body>\n<div>%s</div>\n</body></html>", out)
		minificationCheck(fin, fout)
	})
}

func TestErrorTag(t *testing.T) {
	Convey("Test error tag", t, func() {
		in := "<html><head></head><body></body></html>"
		node, err := html.Parse(bytes.NewReader([]byte(in)))
		So(err, ShouldEqual, nil)

		node.FirstChild.Type = html.ErrorNode
		So(RunMinification1(node).Error(), ShouldEqual, ErrMinificationUnexpectedNodeType.Error())
	})
}

func TestDoctype(t *testing.T) {
	Convey("Test doctype", t, func() {
		in := "<!DOCTYPE html><html><head></head><body></body></html>"

		minificationCheck(in, in)
	})
}

func TestRemoveComments(t *testing.T) {
	Convey("Remove comment", t, func() {
		in := `<html><head>
<!-- Comment1 -->
</head><body></body></html>`

		minificationCheck(in, emptyHead)
	})

	Convey("Remove double comment", t, func() {
		in := `<html><head>
	<!-- Comment1 --><!-- Comment2 -->
	</head><body></body></html>`

		minificationCheck(in, emptyHead)
	})
}

func TestFuncRemoveAttr(t *testing.T) {
	HelperBody("Attributes are not deleted", t,
		`<div begin="begin" end="end">text</div>`,
		`<div begin="begin" end="end">text</div>`)

	HelperBody("Left and right attributes are not removed", t,
		`<div begin="begin" id="remove" end="end">text</div>`,
		`<div begin="begin" end="end">text</div>`)

	HelperBody("Left attributes are not removed", t,
		`<div begin="begin" alt="remove">text</div>`,
		`<div begin="begin">text</div>`)

	HelperBody("Right attributes are not removed", t,
		`<div alt="remove" end="end">text</div>`,
		`<div end="end">text</div>`)

	HelperBody("One attribute for remove", t,
		`<div cols="remove">text</div>`,
		`<div>text</div>`)

	HelperBody("All attributes for remove", t,
		`<div class="remove" title="remove" width="remove" disabled>text</div>`,
		`<div>text</div>`)
}

func TestRemoveAttrs(t *testing.T) {
	attrs := []string{
		"id",
		"alt",
		"cols",
		"class",
		"title",
		"width",
		"align",
		"style",
		"color",
		"valign",
		"target",
		"height",
		"border",
		"hspace",
		"vspace",
		"bgcolor",
		"onclick",
		"colspan",
		"itemprop",
		"disabled",
		"itemtype",
		"itemscope",
		"data-width",
		"cellspacing",
		"cellpadding",
		"bordercolor",
	}
	for _, attr := range attrs {
		HelperBody("Removing attribute "+attr, t,
			fmt.Sprintf(`<div %s="remove">text</div>`, attr),
			"<div>text</div>")
	}
}

func TestFuncRemoveNode(t *testing.T) {
	HelperDiv("One tag inside", t,
		`<form attr="a"><div>aaa</div></form>`,
		" ")

	HelperDiv("Left text", t,
		"pre<form></form>",
		"pre ")

	HelperDiv("Left text with space", t,
		"pre \n \t \r <form></form>",
		"pre ")

	HelperDiv("Right text", t,
		"<form></form>post",
		" post")

	HelperDiv("Right text with space", t,
		"<form></form> \n \t \r post",
		" post")

	HelperDiv("Left and right text", t,
		"pre<form></form>post",
		"pre post")

	HelperDiv("Left and right text with space", t,
		"pre \n \t \r <form></form> \n \t \r post",
		"pre post")

	HelperDiv("Left tag", t,
		"<div>pre</div><form></form>",
		"<div>pre</div> ")

	HelperDiv("Right tag", t,
		"<form></form><div>post</div>",
		" <div>post</div>")

	HelperDiv("Left and right tag", t,
		"<div>pre</div><form></form><div>post</div>",
		"<div>pre</div> <div>post</div>")
}

func TestRemoveTags(t *testing.T) {
	tags := []string{
		"<script>i=0;</script>",
		"<button>text</button>",
		"<object>text</object>",
		"<select>text</select>",
		"<style>.a: 1px</style>",
		"<param name=\"a\"/>",
		"<embed a=\"1\"></embed>",
		"<form>text</form>",
		"<time>text</time>",
		"<img src=\"URL\"></img>",
		"<svg>text</svg>",
		"<canvas>text</canvas>",
		"<br/>",
		"<hr/>",
	}
	for _, tagName := range tags {
		HelperBody("Removing tag "+html.EscapeString(tagName), t,
			fmt.Sprintf("pre%spost", tagName),
			"pre post")
	}

	HelperBody("Removing tag hidden input", t,
		`pre<input type="hidden" />post`,
		"pre post")

	HelperBody("Not Removing no hidden tag input", t,
		`pre<input type="hidden1"/>post`,
		`pre<input type="hidden1"/>post`)

	HelperBody("Not Removing tag input without type", t,
		`pre<input v="val"/>post`,
		`pre<input v="val"/>post`)

	HelperBody("Removing tag wdr", t,
		"pre<wbr>post",
		"prepost")

	HelperDiv("Removing tag wdr between tags", t,
		"pre</div><wbr><div>post",
		"pre</div><div>post")
}

func TestFuncOpenNodeWithoutSeparator(t *testing.T) {
	HelperDiv("Empty", t,
		"<b></b>",
		"")

	HelperDiv("Text inside", t,
		"<b>itext</b>",
		"itext")

	HelperDiv("One tag inside", t,
		"<b><a>itext</a></b>",
		"<a>itext</a>")

	HelperDiv("Left text, right tag inside", t,
		"<b>ipre<a>itext</a></b>",
		"ipre<a>itext</a>")

	HelperDiv("Left tag, right text inside", t,
		"<b><a>itext</a>ipost</b>",
		"<a>itext</a>ipost")

	HelperDiv("Left text, right text inside", t,
		"<b>ipre<a>itext</a>ipost</b>",
		"ipre<a>itext</a>ipost")

	HelperDiv("Text {text text} text", t,
		"pre<b>ipre<a>itext</a>ipost</b>post",
		"preipre<a>itext</a>ipostpost")

	HelperDiv("Text {tag tag} text", t,
		"pre<b><a>itext1</a><a>itext2</a></b>post",
		"pre<a>itext1</a><a>itext2</a>post")

	HelperDiv("Tag {text text} tag", t,
		"<div>pre</div><b>ipre<a>itext</a>ipost</b><div>post</div>",
		"<div>pre</div>ipre<a>itext</a>ipost<div>post</div>")

	HelperDiv("Tag {tag tag} tag", t,
		"<div>pre</div><b><a>itext1</a><a>itext2</a></b><div>post</div>",
		"<div>pre</div><a>itext1</a><a>itext2</a><div>post</div>")

	HelperDiv("Tag {text} tag", t,
		"<div>pre</div><b>itext</b><div>post</div>",
		"<div>pre</div>itext<div>post</div>")

	HelperDiv("Tag {tag} tag", t,
		"<div>pre</div><b><a>itext</a></b><div>post</div>",
		"<div>pre</div><a>itext</a><div>post</div>")

	HelperDiv("One child", t,
		" <b>text</b> <script>s</script>",
		" text ")
}

func TestFuncOpenNodeWithSpaces(t *testing.T) {
	HelperDiv("Text space {text} space text", t,
		"pre <b>text</b> post",
		"pre text post")

	HelperDiv("space {text} space", t,
		" <b>text</b> ",
		" text ")

	HelperDiv("{tag} space text", t,
		"<b><a>text</a></b> post",
		"<a>text</a> post")

	HelperDiv("text space {tag}", t,
		"pre <b><a>text</a></b>",
		"pre <a>text</a>")
}

func TestFuncOpenNodeWithSeparator(t *testing.T) {
	HelperDiv("Empty", t,
		"<abbr></abbr>",
		" ")

	HelperDiv("Text inside", t,
		"<abbr>itext</abbr>",
		" itext ")

	HelperDiv("One tag inside", t,
		"<abbr><a>itext</a></abbr>",
		" <a>itext</a> ")

	HelperDiv("Left text, right tag inside", t,
		"<abbr>ipre<a>itext</a></abbr>",
		" ipre<a>itext</a> ")

	HelperDiv("Left tag, right text inside", t,
		"<abbr><a>itext</a>ipost</abbr>",
		" <a>itext</a>ipost ")

	HelperDiv("Left text, right text inside", t,
		"<abbr>ipre<a>itext</a>ipost</abbr>",
		" ipre<a>itext</a>ipost ")

	HelperDiv("Text {text text} text", t,
		"pre<abbr>ipre<a>itext</a>ipost</abbr>post",
		"pre ipre<a>itext</a>ipost post")

	HelperDiv("Text {tag tag} text", t,
		"pre<abbr><a>itext1</a><a>itext2</a></abbr>post",
		"pre <a>itext1</a><a>itext2</a> post")

	HelperDiv("Tag {text text} tag", t,
		"<div>pre</div><abbr>ipre<a>itext</a>ipost</abbr><div>post</div>",
		"<div>pre</div> ipre<a>itext</a>ipost <div>post</div>")

	HelperDiv("Tag {tag tag} tag", t,
		"<div>pre</div><abbr><a>itext1</a><a>itext2</a></abbr><div>post</div>",
		"<div>pre</div> <a>itext1</a><a>itext2</a> <div>post</div>")

	HelperDiv("Tag {text} tag", t,
		"<div>pre</div><abbr>itext</abbr><div>post</div>",
		"<div>pre</div> itext <div>post</div>")

	HelperDiv("Tag {tag} tag", t,
		"<div>pre</div><abbr><a>itext</a></abbr><div>post</div>",
		"<div>pre</div> <a>itext</a> <div>post</div>")
}

func TestOpenTags(t *testing.T) {
	HelperDiv("Open tag abbr with title", t,
		"<abbr title=\"title value\">text</abbr>", " title value text ")

	HelperDiv("Open tag abbr without title", t,
		"<abbr>text</abbr>", " text ")

	HelperDiv("Open tag abbr without text", t,
		"<abbr title=\"title value\"></abbr>", " title value ")

	HelperDiv("Open tag caption", t,
		"<table><caption>text</caption></table>", "<table> text </table>")

	// without add space
	tags := []string{
		"b",
		"i",
		"del",
		"em",
		"font",
		"ins",
		"q",
		"s",
		"small",
		"span",
	}
	for _, tagName := range tags {
		HelperDiv("Open tag "+tagName, t,
			fmt.Sprintf("pre<%s>text</%s>post", tagName, tagName),
			"pretextpost")
	}

	// with add space
	tags = []string{
		"blockquote",
	}
	for _, tagName := range tags {
		HelperDiv("Open tag "+tagName, t,
			fmt.Sprintf("pre<%s>text</%s>post", tagName, tagName),
			"pre text post")
	}
}
