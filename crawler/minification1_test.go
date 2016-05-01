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
	Convey("Attributes are not deleted", t, func() {
		in := `<html><head></head><body>
	<div begin="begin" end="end">text</div>
	</body></html>`
		out := `<html><head></head><body>
	<div begin="begin" end="end">text</div>
	</body></html>`
		minificationCheck(in, out)
	})

	Convey("Left and right attributes are not removed", t, func() {
		in := `<html><head></head><body>
	<div begin="begin" id="remove" end="end">text</div>
	</body></html>`
		out := `<html><head></head><body>
	<div begin="begin" end="end">text</div>
	</body></html>`
		minificationCheck(in, out)
	})

	Convey("Left attributes are not removed", t, func() {
		in := `<html><head></head><body>
	<div begin="begin" alt="remove">text</div>
	</body></html>`
		out := `<html><head></head><body>
	<div begin="begin">text</div>
	</body></html>`
		minificationCheck(in, out)
	})

	Convey("Right attributes are not removed", t, func() {
		in := `<html><head></head><body>
	<div alt="remove" end="end">text</div>
	</body></html>`
		out := `<html><head></head><body>
	<div end="end">text</div>
	</body></html>`
		minificationCheck(in, out)
	})

	Convey("One attribute for remove", t, func() {
		in := `<html><head></head><body>
	<div cols="remove">text</div>
	</body></html>`
		out := `<html><head></head><body>
	<div>text</div>
	</body></html>`
		minificationCheck(in, out)
	})

	Convey("All attributes for remove", t, func() {
		in := `<html><head></head><body>
	<div class="remove" title="remove" width="remove" disabled>text</div>
	</body></html>`
		out := `<html><head></head><body>
	<div>text</div>
	</body></html>`
		minificationCheck(in, out)
	})
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
		Convey("Removing attribute "+attr, t, func() {
			out := `<html><head></head><body>
<div>text</div>
</body></html>`
			in := fmt.Sprintf(`<html><head></head><body>
<div %s="remove">text</div>
</body></html>`, attr)
			minificationCheck(in, out)
		})
	}
}

func TestFuncRemoveNode(t *testing.T) {
	Convey("One tag inside", t, func() {
		in := `<html><head></head><body>
<div><form attr="a"><div>aaa</div></form></div>
</body></html>`
		out := `<html><head></head><body>
<div> </div>
</body></html>`

		minificationCheck(in, out)
	})

	Convey("Left text", t, func() {
		in := `<html><head></head><body>
<div>pre<form></form></div>
</body></html>`
		out := `<html><head></head><body>
<div>pre </div>
</body></html>`

		minificationCheck(in, out)
	})

	Convey("Left text with space", t, func() {
		in := "<html><head></head><body>\n" +
			"<div>pre \n \t \r <form></form></div>\n" +
			"</body></html>"
		out := `<html><head></head><body>
<div>pre </div>
</body></html>`

		minificationCheck(in, out)
	})

	Convey("Right text", t, func() {
		in := `<html><head></head><body>
<div><form></form>post</div>
</body></html>`
		out := `<html><head></head><body>
<div> post</div>
</body></html>`

		minificationCheck(in, out)
	})

	Convey("Right text with space", t, func() {
		in := "<html><head></head><body>\n" +
			"<div><form></form> \n \t \r post</div>\n" +
			"</body></html>"
		out := `<html><head></head><body>
<div> post</div>
</body></html>`

		minificationCheck(in, out)
	})

	Convey("Left and right text", t, func() {
		in := `<html><head></head><body>
<div>pre<form></form>post</div>
</body></html>`
		out := `<html><head></head><body>
<div>pre post</div>
</body></html>`

		minificationCheck(in, out)
	})

	Convey("Left and right text with space", t, func() {
		in := "<html><head></head><body>\n" +
			"<div>pre \n \t \r <form></form> \n \t \r post</div>\n" +
			"</body></html>"
		out := `<html><head></head><body>
<div>pre post</div>
</body></html>`

		minificationCheck(in, out)
	})

	Convey("Left tag", t, func() {
		in := `<html><head></head><body>
<div><div>pre</div><form></form></div>
</body></html>`
		out := `<html><head></head><body>
<div><div>pre</div> </div>
</body></html>`

		minificationCheck(in, out)
	})

	Convey("Right tag", t, func() {
		in := `<html><head></head><body>
<div><form></form><div>post</div></div>
</body></html>`
		out := `<html><head></head><body>
<div> <div>post</div></div>
</body></html>`

		minificationCheck(in, out)
	})

	Convey("Left and right tag", t, func() {
		in := `<html><head></head><body>
<div><div>pre</div><form></form><div>post</div></div>
</body></html>`
		out := `<html><head></head><body>
<div><div>pre</div> <div>post</div></div>
</body></html>`

		minificationCheck(in, out)
	})
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
		"<br/>",
		"<hr/>",
	}
	out := `<html><head></head><body> </body></html>`
	for _, tagName := range tags {
		Convey("Removing tag "+html.EscapeString(tagName), t, func() {
			in := fmt.Sprintf(`<html><head></head><body>
%s
</body></html>`, tagName)
			minificationCheck(in, out)
		})
	}

	Convey("Removing tag hidden input", t, func() {
		in := `<html><head></head><body>
<input type="hidden" />
</body></html>`
		minificationCheck(in, out)
	})

	Convey("Not Removing no hidden tag input", t, func() {
		in := `<html><head></head><body>
<input type="hidden1"/>
</body></html>`
		minificationCheck(in, in)
	})

	Convey("Not Removing tag input without type", t, func() {
		in := `<html><head></head><body>
<input v="val"/>
</body></html>`
		minificationCheck(in, in)
	})

	Convey("Removing tag wdr", t, func() {
		in := `<html><head></head><body>
pre<wbr>post
</body></html>`
		out := `<html><head></head><body>
prepost
</body></html>`
		minificationCheck(in, out)
	})

	Convey("Removing tag wdr between tags", t, func() {
		in := `<html><head></head><body>
<div>pre</div><wbr><div>post</div>
</body></html>`
		out := `<html><head></head><body>
<div>pre</div><div>post</div>
</body></html>`
		minificationCheck(in, out)
	})
}

func HelperFuncOpenNode(name string, t *testing.T, in, out string) {
	Convey(name, t, func() {
		fin := fmt.Sprintf(`<html><head></head><body>
<div>%s</div>
</body></html>`, in)
		fout := fmt.Sprintf(`<html><head></head><body>
<div>%s</div>
</body></html>`, out)

		minificationCheck(fin, fout)
	})
}

func TestFuncOpenNode(t *testing.T) {
	HelperFuncOpenNode("Empty", t,
		"<b></b>", "")
	HelperFuncOpenNode("Text inside", t,
		"<b>itext</b>", "itext")
	HelperFuncOpenNode("One tag inside", t,
		"<b><a>itext</a></b>", "<a>itext</a>")
	HelperFuncOpenNode("Left text, right tag inside", t,
		"<b>ipre<a>itext</a></b>", "ipre<a>itext</a>")
	HelperFuncOpenNode("Left tag, right text inside", t,
		"<b><a>itext</a>ipost</b>", "<a>itext</a>ipost")
	HelperFuncOpenNode("Left text, right text inside", t,
		"<b>ipre<a>itext</a>ipost</b>", "ipre<a>itext</a>ipost")
	HelperFuncOpenNode("Text {text text} text", t,
		"pre<b>ipre<a>itext</a>ipost</b>post", "preipre<a>itext</a>ipostpost")
	HelperFuncOpenNode("Text {tag tag} text", t,
		"pre<b><a>itext1</a><a>itext2</a></b>post", "pre<a>itext1</a><a>itext2</a>post")
	HelperFuncOpenNode("Tag {text text} tag", t,
		"<div>pre</div><b>ipre<a>itext</a>ipost</b><div>post</div>",
		"<div>pre</div>ipre<a>itext</a>ipost<div>post</div>")
	HelperFuncOpenNode("Tag {tag tag} tag", t,
		"<div>pre</div><b><a>itext1</a><a>itext2</a></b><div>post</div>",
		"<div>pre</div><a>itext1</a><a>itext2</a><div>post</div>")
	HelperFuncOpenNode("Tag {text} tag", t,
		"<div>pre</div><b>itext</b><div>post</div>", "<div>pre</div>itext<div>post</div>")
	HelperFuncOpenNode("Tag {tag} tag", t,
		"<div>pre</div><b><a>itext</a></b><div>post</div>", "<div>pre</div><a>itext</a><div>post</div>")

	HelperFuncOpenNode("Text space {text} space text", t,
		"<div>pre <b>text</b> post</div>", "<div>pre text post</div>")
	HelperFuncOpenNode("space {text} space", t,
		"<div> <b>text</b> </div>", "<div> text </div>")
	HelperFuncOpenNode("{tag} space text", t,
		"<div><b><a>text</a></b> post</div>", "<div><a>text</a> post</div>")
	HelperFuncOpenNode("text space {tag}", t,
		"<div>pre <b><a>text</a></b></div>", "<div>pre <a>text</a></div>")

	HelperFuncOpenNode("Empty with delimeter", t,
		"<abbr></abbr>", " ")
	HelperFuncOpenNode("Text inside with delimeter", t,
		"<abbr>itext</abbr>", " itext ")
	HelperFuncOpenNode("One tag inside with delimeter", t,
		"<abbr><a>itext</a></abbr>", " <a>itext</a> ")
	HelperFuncOpenNode("Left text, right tag inside with delimeter", t,
		"<abbr>ipre<a>itext</a></abbr>", " ipre<a>itext</a> ")
	HelperFuncOpenNode("Left tag, right text inside with delimeter", t,
		"<abbr><a>itext</a>ipost</abbr>", " <a>itext</a>ipost ")
	HelperFuncOpenNode("Left text, right text inside with delimeter", t,
		"<abbr>ipre<a>itext</a>ipost</abbr>", " ipre<a>itext</a>ipost ")
	HelperFuncOpenNode("Text {text text} text with delimeter", t,
		"pre<abbr>ipre<a>itext</a>ipost</abbr>post", "pre ipre<a>itext</a>ipost post")
	HelperFuncOpenNode("Text {tag tag} text with delimeter", t,
		"pre<abbr><a>itext1</a><a>itext2</a></abbr>post", "pre <a>itext1</a><a>itext2</a> post")
	HelperFuncOpenNode("Tag {text text} tag with delimeter", t,
		"<div>pre</div><abbr>ipre<a>itext</a>ipost</abbr><div>post</div>",
		"<div>pre</div> ipre<a>itext</a>ipost <div>post</div>")
	HelperFuncOpenNode("Tag {tag tag} tag with delimeter", t,
		"<div>pre</div><abbr><a>itext1</a><a>itext2</a></abbr><div>post</div>",
		"<div>pre</div> <a>itext1</a><a>itext2</a> <div>post</div>")
	HelperFuncOpenNode("Tag {text} tag with delimeter", t,
		"<div>pre</div><abbr>itext</abbr><div>post</div>", "<div>pre</div> itext <div>post</div>")
	HelperFuncOpenNode("Tag {tag} tag with delimeter", t,
		"<div>pre</div><abbr><a>itext</a></abbr><div>post</div>", "<div>pre</div> <a>itext</a> <div>post</div>")

	HelperFuncOpenNode("One child", t,
		"<div> <b>text</b> <script>s</script></div>", "<div> text </div>")
}

func TestOpenTags(t *testing.T) {
	HelperFuncOpenNode("Open tag abbr with title", t,
		"<abbr title=\"title value\">text</abbr>", " title value text ")
	HelperFuncOpenNode("Open tag abbr without title", t, "<abbr>text</abbr>", " text ")
	HelperFuncOpenNode("Open tag abbr without text", t, "<abbr title=\"title value\"></abbr>", " title value ")
	HelperFuncOpenNode("Open tag blockquote", t, "<blockquote>pre<div>text</div>post</blockquote>", " pre<div>text</div>post ")

	tags := []string{
		"b",
		"i",
	}
	out := `<html><head></head><body>
<div>text</div>
</body></html>`
	for _, tagName := range tags {
		Convey("Open tag "+tagName, t, func() {
			in := fmt.Sprintf(`<html><head></head><body>
<div><%s>text</%s></div>
</body></html>`, tagName, tagName)

			minificationCheck(in, out)
		})
	}
}
