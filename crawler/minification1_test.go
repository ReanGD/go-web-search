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
		err := html.Render(w, node)
		if err == nil {
			err = w.Flush()
			if err == nil {
				So(string(buf.Bytes()), ShouldEqual, out)
			}
		}
	}
	So(err, ShouldEqual, nil)
}

func HelperHead(name string, t *testing.T, in, out string) {
	Convey(name, t, func() {
		fin := fmt.Sprintf("<html><head>%s</head><body><div>text</div></body></html>", in)
		fout := fmt.Sprintf("<html><head>%s</head><body><div>text</div></body></html>", out)
		minificationCheck(fin, fout)
	})
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
		So(RunMinification1(node).Error(), ShouldEqual, ErrMinification1UnexpectedNodeType.Error())
	})
}

func TestDoctype(t *testing.T) {
	Convey("Test doctype", t, func() {
		in := "<!DOCTYPE html><html><head></head><body></body></html>"
		out := "<html><head></head><body></body></html>"

		minificationCheck(in, out)
	})
}

func TestRemoveComments(t *testing.T) {
	HelperHead("Remove comment in head", t, "<!-- Comment1 -->", "")
	HelperBody("Remove comment in body", t, "pre<!-- Comment1 -->post", "prepost")
	HelperHead("Remove double comment in head", t, "<!-- Comment1 --><!-- Comment2 -->", "")
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
		"<img src=\"URL\"></img>",
		"<svg>text</svg>",
		"<canvas>text</canvas>",
		"<video>text</video>",
		"<textarea>text</textarea>",
		"<noscript>text</noscript>",
		"<applet>text</applet>",
		"<audio>text</audio>",
		"<noindex>text<div>text</div>text</noindex>",
		`<iframe src="page/banner.html" width="468" height="60" align="left">
Ваш браузер не поддерживает встроенные фреймы!</iframe>`,
		`<input type="hidden"/>`,
		`<command onclick="alert">`,
		"<area/>",
		"<br/>",
		"<hr/>",
	}
	for _, tagName := range tags {
		HelperBody("Removing tag "+html.EscapeString(tagName), t,
			fmt.Sprintf("pre%spost", tagName),
			"pre post")
	}

	tags = []string{
		"<title>text</title>",
		`<meta name="title" content="title meta">`,
		`<link rel="stylesheet" href="ie.css">`,
		`<base href="ie.css">`,
		`<basefont face="Arial">`,
		`<bgsound src="1.ogg">`,
	}
	for _, tagName := range tags {
		HelperHead("Removing tag "+html.EscapeString(tagName), t, tagName, "")
	}

	tags = []string{
		`<data value="1"></data>`,
		`<datalist id="<идентификатор>"><option value="text1"></datalist>`,
	}
	for _, tagName := range tags {
		HelperDiv("Removing tag "+html.EscapeString(tagName), t,
			fmt.Sprintf("pre%spost", tagName),
			"prepost")
	}

	HelperBody("Removing tag wdr", t,
		"pre<wbr>post",
		"prepost")

	HelperDiv("Removing tag wdr between tags", t,
		"pre</div><wbr><div>post",
		"pre</div><div>post")

	Convey("Removing tag frameset", t, func() {
		in := `<html><head></head><frameset></frameset></html>`
		out := "<html><head></head></html>"
		minificationCheck(in, out)
	})
}

func TestFuncOpenNodeWithoutSeparator(t *testing.T) {
	HelperDiv("Empty", t,
		"<b></b>",
		"")

	HelperDiv("Text inside", t,
		"<b>itext</b>",
		"itext")

	HelperDiv("One tag inside", t,
		`<b><test_tag>itext</test_tag></b>`,
		`<test_tag>itext</test_tag>`)

	HelperDiv("Left text, right tag inside", t,
		`<b>ipre<test_tag>itext</test_tag></b>`,
		`ipre<test_tag>itext</test_tag>`)

	HelperDiv("Left tag, right text inside", t,
		`<b><test_tag>itext</test_tag>ipost</b>`,
		`<test_tag>itext</test_tag>ipost`)

	HelperDiv("Left text, right text inside", t,
		`<b>ipre<test_tag>itext</test_tag>ipost</b>`,
		`ipre<test_tag>itext</test_tag>ipost`)

	HelperDiv("Text {text text} text", t,
		`pre<b>ipre<test_tag>itext</test_tag>ipost</b>post`,
		`preipre<test_tag>itext</test_tag>ipostpost`)

	HelperDiv("Text {tag tag} text", t,
		`pre<b><test_tag>itext1</test_tag><test_tag>itext2</test_tag></b>post`,
		`pre<test_tag>itext1</test_tag><test_tag>itext2</test_tag>post`)

	HelperDiv("Tag {text text} tag", t,
		`<div>pre</div><b>ipre<test_tag>itext</test_tag>ipost</b><div>post</div>`,
		`<div>pre</div>ipre<test_tag>itext</test_tag>ipost<div>post</div>`)

	HelperDiv("Tag {tag tag} tag", t,
		`<div>pre</div><b><test_tag>itext1</test_tag><test_tag>itext2</test_tag></b><div>post</div>`,
		`<div>pre</div><test_tag>itext1</test_tag><test_tag>itext2</test_tag><div>post</div>`)

	HelperDiv("Tag {text} tag", t,
		"<div>pre</div><b>itext</b><div>post</div>",
		"<div>pre</div>itext<div>post</div>")

	HelperDiv("Tag {tag} tag", t,
		`<div>pre</div><b><test_tag>itext</test_tag></b><div>post</div>`,
		`<div>pre</div><test_tag>itext</test_tag><div>post</div>`)

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
		`<b><test_tag>text</test_tag></b> post`,
		`<test_tag>text</test_tag> post`)

	HelperDiv("text space {tag}", t,
		`pre <b><test_tag>text</test_tag></b>`,
		`pre <test_tag>text</test_tag>`)
}

func TestFuncOpenNodeWithSeparator(t *testing.T) {
	HelperDiv("Empty", t,
		"<abbr></abbr>",
		" ")

	HelperDiv("Text inside", t,
		"<abbr>itext</abbr>",
		" itext ")

	HelperDiv("One tag inside", t,
		`<abbr><test_tag>itext</test_tag></abbr>`,
		` <test_tag>itext</test_tag> `)

	HelperDiv("Left text, right tag inside", t,
		`<abbr>ipre<test_tag>itext</test_tag></abbr>`,
		` ipre<test_tag>itext</test_tag> `)

	HelperDiv("Left tag, right text inside", t,
		`<abbr><test_tag>itext</test_tag>ipost</abbr>`,
		` <test_tag>itext</test_tag>ipost `)

	HelperDiv("Left text, right text inside", t,
		`<abbr>ipre<test_tag>itext</test_tag>ipost</abbr>`,
		` ipre<test_tag>itext</test_tag>ipost `)

	HelperDiv("Text {text text} text", t,
		`pre<abbr>ipre<test_tag>itext</test_tag>ipost</abbr>post`,
		`pre ipre<test_tag>itext</test_tag>ipost post`)

	HelperDiv("Text {tag tag} text", t,
		`pre<abbr><test_tag>itext1</test_tag><test_tag>itext2</test_tag></abbr>post`,
		`pre <test_tag>itext1</test_tag><test_tag>itext2</test_tag> post`)

	HelperDiv("Tag {text text} tag", t,
		`<div>pre</div><abbr>ipre<test_tag>itext</test_tag>ipost</abbr><div>post</div>`,
		`<div>pre</div> ipre<test_tag>itext</test_tag>ipost <div>post</div>`)

	HelperDiv("Tag {tag tag} tag", t,
		`<div>pre</div><abbr><test_tag>itext1</test_tag><test_tag>itext2</test_tag></abbr><div>post</div>`,
		`<div>pre</div> <test_tag>itext1</test_tag><test_tag>itext2</test_tag> <div>post</div>`)

	HelperDiv("Tag {text} tag", t,
		"<div>pre</div><abbr>itext</abbr><div>post</div>",
		"<div>pre</div> itext <div>post</div>")

	HelperDiv("Tag {tag} tag", t,
		`<div>pre</div><abbr><test_tag>itext</test_tag></abbr><div>post</div>`,
		`<div>pre</div> <test_tag>itext</test_tag> <div>post</div>`)
}

func TestConvertTagToDiv(t *testing.T) {
	tags := []string{
		"article",
		"header",
		"details",
		"dialog",
		"listing",
		"aside",
		"footer",
		"blockquote",
		"center",
		"figure",
		"section",
		"label",
		"nav",
		"pre",
		"h1",
		"h2",
		"h3",
		"h4",
		"h5",
		"h6",
		"p",
		"undef_tag",
	}
	for _, tagName := range tags {
		HelperDiv("Convert tag "+tagName, t,
			fmt.Sprintf("<%s>text</%s>", tagName, tagName),
			"<div>text</div>")
	}
}

func TestOpenTags(t *testing.T) {
	HelperDiv("Open tag abbr with title", t,
		"<abbr title=\"title value\">text</abbr>",
		" title value text ")

	HelperDiv("Open tag abbr without title", t,
		"<abbr>text</abbr>",
		" text ")

	HelperDiv("Open tag abbr without text", t,
		"<abbr title=\"title value\"></abbr>",
		" title value ")

	// without add space
	tags := []string{
		"b",
		"i",
		"del",
		"dfn",
		"em",
		"font",
		"ins",
		"q",
		"nobr",
		"s",
		"small",
		"span",
		"strike",
		"strong",
		"sub",
		"sup",
		"tt",
		"u",
		"var",
		"time",
		"cite",
		"code",
		"a",
		"blink",
		"big",
		"bdi",
		"bdo",
		"name",
	}
	for _, tagName := range tags {
		HelperDiv("Open tag "+tagName, t,
			fmt.Sprintf("pre<%s>text</%s>post", tagName, tagName),
			"pretextpost")
	}

	// with add space
	tags = []string{
		"figcaption",
		"address",
		"marquee",
	}
	for _, tagName := range tags {
		HelperDiv("Open tag "+tagName, t,
			fmt.Sprintf("pre<%s>text</%s>post", tagName, tagName),
			"pre text post")
	}
}

func TestLists(t *testing.T) {
	HelperBody("Ul", t,
		`<ul a="a"><li a="a">text1</li><li a="a">text2</li></ul>`,
		"<div> text1 text2 </div>")

	HelperBody("Ol", t,
		`<ol a="a"><li a="a">text1</li><li a="a">text2</li></ol>`,
		"<div> text1 text2 </div>")

	HelperBody("Dt", t,
		`<dl a="a">
<dt a="a">dt text1</dt><dd a="a">dd text1</dd>
<dt a="a">dt text2</dt><dd a="a">dd text2</dd>
</dl>`,
		"<div> dt text1 dd text1 dt text2 dd text2 </div>")
}

func TestTable(t *testing.T) {
	HelperBody("Open tag table.caption", t,
		`<table a="a"><caption>caption text</caption></table>`,
		`<div> caption text </div>`)

	HelperBody("Remove tag table.col table.colgroup", t,
		`<table>
<col width="150" valign="top1">
<col width="150" valign="top2">
<colgroup width="150">
</table>`,
		`<div> </div>`)

	HelperBody("Open tag table.thead table.tbody table.tfoot", t,
		`<table><thead></thead><tbody></tbody><tfoot></tfoot></table>`,
		`<div> </div>`)

	HelperBody("Open table data tags", t,
		`<table>
<tr>
 <th>Text th1</th>
 <th>Text th2</th>
</tr>
<tr>
 <td>Text td1</td>
 <td>Text td2</td>
</tr>
</table>`,
		`<div> <div>Text th1</div>
 <div>Text th2</div> <div>Text td1</div>
 <div>Text td2</div> </div>`)

	HelperBody("Table full version", t,
		`<table>
	<col width="150" valign="top1">
	<col width="150" valign="top2">
	<colgroup width="150">
<thead>
 <tr>
  <th>Text th1</th>
  <th>Text th2</th>
 </tr>
</thead>
<tbody>
 <tr>
  <td>Text td1</td>
  <td>Text td2</td>
 </tr>
</tbody>
<tfoot>
 <tr>
  <td>Text td3</td>
  <td>Text td4</td>
 </tr>
</tfoot>
	</table>`,
		`<div> <div>Text th1</div>
  <div>Text th2</div> <div>Text td1</div>
  <div>Text td2</div> <div>Text td3</div>
  <div>Text td4</div> </div>`)
}

func TestRemoveAttr(t *testing.T) {
	Convey("Remove attributes in special tags", t, func() {
		in := `<html a="1"><head a="1"></head><body a="1"><div a="1">text</div></body></html>`
		out := `<html><head></head><body><div>text</div></body></html>`
		minificationCheck(in, out)
	})
}
