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

func helperHead(name string, t *testing.T, in, out string) {
	Convey(name, t, func() {
		fin := fmt.Sprintf("<html><head>%s</head><body><div>text</div></body></html>", in)
		fout := fmt.Sprintf("<html><head>%s</head><body><div>text</div></body></html>", out)
		minificationCheck(fin, fout)
	})
}

func helperBody(name string, t *testing.T, in, out string) {
	Convey(name, t, func() {
		fin := fmt.Sprintf("<html><head></head><body>\n%s\n</body></html>", in)
		fout := fmt.Sprintf("<html><head></head><body>\n%s\n</body></html>", out)
		minificationCheck(fin, fout)
	})
}

func helperDiv(name string, t *testing.T, in, out string) {
	Convey(name, t, func() {
		fin := fmt.Sprintf("<html><head></head><body>\n<div>%s</div>\n</body></html>", in)
		fout := fmt.Sprintf("<html><head></head><body>\n<div>%s</div>\n</body></html>", out)
		minificationCheck(fin, fout)
	})
}

func TestErrorNodeType(t *testing.T) {
	Convey("Test error node type", t, func() {
		in := "<html><head></head><body></body></html>"
		node, err := html.Parse(bytes.NewReader([]byte(in)))
		So(err, ShouldEqual, nil)

		node.FirstChild.Type = html.ErrorNode
		So(RunMinification1(node).Error(), ShouldEqual, ErrMinificationHTMLUnexpectedNodeType.Error())
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
	helperHead("Remove comment in head", t, "<!-- Comment1 -->", "")
	helperBody("Remove comment in body", t, "pre<!-- Comment1 -->post", "prepost")
	helperHead("Remove double comment in head", t, "<!-- Comment1 --><!-- Comment2 -->", "")
}

func TestFuncRemoveNode(t *testing.T) {
	helperDiv("One tag inside", t,
		`<form attr="a"><div>aaa</div></form>`,
		" ")

	helperDiv("Left text", t,
		"pre<form></form>",
		"pre ")

	helperDiv("Left text with space", t,
		"pre \n \t \r <form></form>",
		"pre ")

	helperDiv("Right text", t,
		"<form></form>post",
		" post")

	helperDiv("Right text with space", t,
		"<form></form> \n \t \r post",
		" post")

	helperDiv("Left and right text", t,
		"pre<form></form>post",
		"pre post")

	helperDiv("Left and right text with space", t,
		"pre \n \t \r <form></form> \n \t \r post",
		"pre post")

	helperDiv("Left tag", t,
		"<div>pre</div><form></form>",
		"<div>pre</div> ")

	helperDiv("Right tag", t,
		"<form></form><div>post</div>",
		" <div>post</div>")

	helperDiv("Left and right tag", t,
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
		helperBody("Removing tag "+html.EscapeString(tagName), t,
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
		helperHead("Removing tag "+html.EscapeString(tagName), t, tagName, "")
	}

	tags = []string{
		`<data value="1"></data>`,
		`<datalist id="<идентификатор>"><option value="text1"></datalist>`,
	}
	for _, tagName := range tags {
		helperDiv("Removing tag "+html.EscapeString(tagName), t,
			fmt.Sprintf("pre%spost", tagName),
			"prepost")
	}

	helperBody("Removing tag wdr", t,
		"pre<wbr>post",
		"prepost")

	helperDiv("Removing tag wdr between tags", t,
		"pre</div><wbr><div>post",
		"pre</div><div>post")

	Convey("Removing tag frameset", t, func() {
		in := `<html><head></head><frameset></frameset></html>`
		out := "<html><head></head></html>"
		minificationCheck(in, out)
	})
}

func TestFuncOpenNodeWithoutSeparator(t *testing.T) {
	helperDiv("Empty", t,
		"<b></b>",
		"")

	helperDiv("Text inside", t,
		"<b>itext</b>",
		"itext")

	helperDiv("One tag inside", t,
		`<b><test_tag>itext</test_tag></b>`,
		`<test_tag>itext</test_tag>`)

	helperDiv("Left text, right tag inside", t,
		`<b>ipre<test_tag>itext</test_tag></b>`,
		`ipre<test_tag>itext</test_tag>`)

	helperDiv("Left tag, right text inside", t,
		`<b><test_tag>itext</test_tag>ipost</b>`,
		`<test_tag>itext</test_tag>ipost`)

	helperDiv("Left text, right text inside", t,
		`<b>ipre<test_tag>itext</test_tag>ipost</b>`,
		`ipre<test_tag>itext</test_tag>ipost`)

	helperDiv("Text {text text} text", t,
		`pre<b>ipre<test_tag>itext</test_tag>ipost</b>post`,
		`preipre<test_tag>itext</test_tag>ipostpost`)

	helperDiv("Text {tag tag} text", t,
		`pre<b><test_tag>itext1</test_tag><test_tag>itext2</test_tag></b>post`,
		`pre<test_tag>itext1</test_tag><test_tag>itext2</test_tag>post`)

	helperDiv("Tag {text text} tag", t,
		`<div>pre</div><b>ipre<test_tag>itext</test_tag>ipost</b><div>post</div>`,
		`<div>pre</div>ipre<test_tag>itext</test_tag>ipost<div>post</div>`)

	helperDiv("Tag {tag tag} tag", t,
		`<div>pre</div><b><test_tag>itext1</test_tag><test_tag>itext2</test_tag></b><div>post</div>`,
		`<div>pre</div><test_tag>itext1</test_tag><test_tag>itext2</test_tag><div>post</div>`)

	helperDiv("Tag {text} tag", t,
		"<div>pre</div><b>itext</b><div>post</div>",
		"<div>pre</div>itext<div>post</div>")

	helperDiv("Tag {tag} tag", t,
		`<div>pre</div><b><test_tag>itext</test_tag></b><div>post</div>`,
		`<div>pre</div><test_tag>itext</test_tag><div>post</div>`)

	helperDiv("One child", t,
		" <b>text</b> <script>s</script>",
		" text ")
}

func TestFuncOpenNodeWithSpaces(t *testing.T) {
	helperDiv("Text space {text} space text", t,
		"pre <b>text</b> post",
		"pre text post")

	helperDiv("space {text} space", t,
		" <b>text</b> ",
		" text ")

	helperDiv("{tag} space text", t,
		`<b><test_tag>text</test_tag></b> post`,
		`<test_tag>text</test_tag> post`)

	helperDiv("text space {tag}", t,
		`pre <b><test_tag>text</test_tag></b>`,
		`pre <test_tag>text</test_tag>`)
}

func TestFuncOpenNodeWithSeparator(t *testing.T) {
	helperDiv("Empty", t,
		"<abbr></abbr>",
		" ")

	helperDiv("Text inside", t,
		"<abbr>itext</abbr>",
		" itext ")

	helperDiv("One tag inside", t,
		`<abbr><test_tag>itext</test_tag></abbr>`,
		` <test_tag>itext</test_tag> `)

	helperDiv("Left text, right tag inside", t,
		`<abbr>ipre<test_tag>itext</test_tag></abbr>`,
		` ipre<test_tag>itext</test_tag> `)

	helperDiv("Left tag, right text inside", t,
		`<abbr><test_tag>itext</test_tag>ipost</abbr>`,
		` <test_tag>itext</test_tag>ipost `)

	helperDiv("Left text, right text inside", t,
		`<abbr>ipre<test_tag>itext</test_tag>ipost</abbr>`,
		` ipre<test_tag>itext</test_tag>ipost `)

	helperDiv("Text {text text} text", t,
		`pre<abbr>ipre<test_tag>itext</test_tag>ipost</abbr>post`,
		`pre ipre<test_tag>itext</test_tag>ipost post`)

	helperDiv("Text {tag tag} text", t,
		`pre<abbr><test_tag>itext1</test_tag><test_tag>itext2</test_tag></abbr>post`,
		`pre <test_tag>itext1</test_tag><test_tag>itext2</test_tag> post`)

	helperDiv("Tag {text text} tag", t,
		`<div>pre</div><abbr>ipre<test_tag>itext</test_tag>ipost</abbr><div>post</div>`,
		`<div>pre</div> ipre<test_tag>itext</test_tag>ipost <div>post</div>`)

	helperDiv("Tag {tag tag} tag", t,
		`<div>pre</div><abbr><test_tag>itext1</test_tag><test_tag>itext2</test_tag></abbr><div>post</div>`,
		`<div>pre</div> <test_tag>itext1</test_tag><test_tag>itext2</test_tag> <div>post</div>`)

	helperDiv("Tag {text} tag", t,
		"<div>pre</div><abbr>itext</abbr><div>post</div>",
		"<div>pre</div> itext <div>post</div>")

	helperDiv("Tag {tag} tag", t,
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
		helperDiv("Convert tag "+tagName, t,
			fmt.Sprintf("<%s>text</%s>", tagName, tagName),
			"<div>text</div>")
	}
}

func TestOpenTags(t *testing.T) {
	helperDiv("Open tag abbr with title", t,
		"<abbr title=\"title value\">text</abbr>",
		" title value text ")

	helperDiv("Open tag abbr without title", t,
		"<abbr>text</abbr>",
		" text ")

	helperDiv("Open tag abbr without text", t,
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
		helperDiv("Open tag "+tagName, t,
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
		helperDiv("Open tag "+tagName, t,
			fmt.Sprintf("pre<%s>text</%s>post", tagName, tagName),
			"pre text post")
	}
}

func TestLists(t *testing.T) {
	helperBody("Ul", t,
		`<ul a="a"><li a="a">text1</li><li a="a">text2</li></ul>`,
		"<div> text1 text2 </div>")

	helperBody("Ol", t,
		`<ol a="a"><li a="a">text1</li><li a="a">text2</li></ol>`,
		"<div> text1 text2 </div>")

	helperBody("Dt", t,
		`<dl a="a">
<dt a="a">dt text1</dt><dd a="a">dd text1</dd>
<dt a="a">dt text2</dt><dd a="a">dd text2</dd>
</dl>`,
		"<div> dt text1 dd text1 dt text2 dd text2 </div>")
}

func TestTable(t *testing.T) {
	helperBody("Open tag table.caption", t,
		`<table a="a"><caption>caption text</caption></table>`,
		`<div> caption text </div>`)

	helperBody("Remove tag table.col table.colgroup", t,
		`<table>
<col width="150" valign="top1">
<col width="150" valign="top2">
<colgroup width="150">
</table>`,
		`<div> </div>`)

	helperBody("Open tag table.thead table.tbody table.tfoot", t,
		`<table><thead></thead><tbody></tbody><tfoot></tfoot></table>`,
		`<div> </div>`)

	helperBody("Open table data tags", t,
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

	helperBody("Table full version", t,
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
