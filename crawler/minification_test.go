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
	emptyHead = "<html><head>\n\n</head><body></body></html>"
	emptyBody = "<html><head></head><body>\n\n</body></html>"
)

func minificationCheck(in string, out string) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	node, err := html.Parse(bytes.NewReader([]byte(in)))
	if err == nil {
		m := Minification{}
		So(m.Run(node), ShouldEqual, nil)
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
		m := Minification{}
		So(m.Run(node).Error(), ShouldEqual, ErrMinificationUnexpectedNodeType.Error())
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

func TestRemoveTags(t *testing.T) {

	Convey("Removing the tag script", t, func() {
		in := `<html><head>
<script type="text/javascript">i=0;</script>
</head><body></body></html>`

		minificationCheck(in, emptyHead)
	})

	Convey("Removing the tag style", t, func() {
		in := `<html><head>
<style></style>
</head><body></body></html>`

		minificationCheck(in, emptyHead)
	})

	Convey("Removing the tag form", t, func() {
		in := `<html><head></head><body>
<form action="handler"><p>1</p></form>
</body></html>`

		minificationCheck(in, emptyBody)
	})

	Convey("Removing the tags button and img", t, func() {
		in := `<html><head></head><body>
<button>Text</button><img></img>
</body></html>`

		minificationCheck(in, emptyBody)
	})

	Convey("Removal of all other attribute", t, func() {
		tags := []string{
			"br",
			"hr",
		}
		for _, tagName := range tags {
			in := fmt.Sprintf(`<html><head></head><body>
<%s />
</body></html>`, tagName)
			minificationCheck(in, emptyBody)
		}
	})
}

func TestRemoveTextNode(t *testing.T) {
	Convey("Removing a tag inside a div", t, func() {
		in := `<html><head></head><body>
<div><time a="1">remove</time></div>
</body></html>`
		out := `<html><head></head><body>
<div></div>
</body></html>`

		minificationCheck(in, out)
	})
	Convey("Removing a tag with the text on the left", t, func() {
		in := `<html><head></head><body>
<div>pre<time a="1">remove</time></div>
</body></html>`
		out := `<html><head></head><body>
<div>pre</div>
</body></html>`

		minificationCheck(in, out)
	})
	Convey("Removing a tag with the text on the right", t, func() {
		in := `<html><head></head><body>
<div><time a="1">remove</time>post</div>
</body></html>`
		out := `<html><head></head><body>
<div>post</div>
</body></html>`

		minificationCheck(in, out)
	})
	Convey("Removing a tag with the text on the right and left", t, func() {
		in := `<html><head></head><body>
<div>pre<time a="1">remove</time>post</div>
</body></html>`
		out := `<html><head></head><body>
<div>pre post</div>
</body></html>`

		minificationCheck(in, out)
	})
	Convey("Removing a tag when left blank", t, func() {
		in := `<html><head></head><body>
<div> <time a="1">remove</time>post</div>
</body></html>`
		out := `<html><head></head><body>
<div> post</div>
</body></html>`

		minificationCheck(in, out)
	})
	Convey("Removing a tag when right blank", t, func() {
		in := `<html><head></head><body>
<div>pre<time a="1">remove</time> </div>
</body></html>`
		out := `<html><head></head><body>
<div>pre </div>
</body></html>`

		minificationCheck(in, out)
	})
	Convey("Removing the tag when the tag is left", t, func() {
		in := `<html><head></head><body>
<div><div>pre</div><time a="1">remove</time>post</div>
</body></html>`
		out := `<html><head></head><body>
<div><div>pre</div>post</div>
</body></html>`

		minificationCheck(in, out)
	})
	Convey("Removing the tag when the tag is right", t, func() {
		in := `<html><head></head><body>
<div>pre<time a="1">remove</time><div>post</div></div>
</body></html>`
		out := `<html><head></head><body>
<div>pre<div>post</div></div>
</body></html>`

		minificationCheck(in, out)
	})
	Convey("Removing a tag from the right line break", t, func() {
		in := `<html><head></head><body>
<div>pre<time a="1">remove</time>
<div>post</div></div>
</body></html>`
		out := `<html><head></head><body>
<div>pre
<div>post</div></div>
</body></html>`

		minificationCheck(in, out)
	})
}

func TestMinimizeHierarchyNode(t *testing.T) {

	Convey("Minimize. One tag inside", t, func() {
		in := `<html><head></head><body>
<div><span><a>ref</a></span></div>
</body></html>`
		out := `<html><head></head><body>
<div><a>ref</a></div>
</body></html>`

		minificationCheck(in, out)
	})

	Convey("Minimize. Text inside", t, func() {
		in := `<html><head></head><body>
<div><span>text</span></div>
</body></html>`
		out := `<html><head></head><body>
<div><span>text</span></div>
</body></html>`

		minificationCheck(in, out)
	})

	Convey("Minimize. Double tag inside", t, func() {
		in := `<html><head></head><body>
<div><span><br/><a>ref</a></span></div>
</body></html>`
		out := `<html><head></head><body>
<div><span><br/><a>ref</a></span></div>
</body></html>`

		minificationCheck(in, out)
	})

	Convey("Minimize. Tag with text inside", t, func() {
		in := `<html><head></head><body>
<div><span>text<a>ref</a></span></div>
</body></html>`
		out := `<html><head></head><body>
<div><span>text<a>ref</a></span></div>
</body></html>`

		minificationCheck(in, out)
	})

	Convey("Minimize. Tag with text inside and attr", t, func() {
		in := `<html><head></head><body>
<div><span class="a">text<a>ref</a></span></div>
</body></html>`
		out := `<html><head></head><body>
<div><span>text<a>ref</a></span></div>
</body></html>`

		minificationCheck(in, out)
	})
}

func TestRemoveAttrs(t *testing.T) {

	Convey("Attributes are not deleted", t, func() {
		in := `<html><head></head><body>
<div begin="begin" end="end">text</div>
</body></html>`
		out := `<html><head></head><body>
<div begin="begin" end="end">text</div>
</body></html>`
		minificationCheck(in, out)
	})

	Convey("Removal of the id attribute", t, func() {
		in := `<html><head></head><body>
<div begin="begin" id="remove" end="end">text</div>
</body></html>`
		out := `<html><head></head><body>
<div begin="begin" end="end">text</div>
</body></html>`
		minificationCheck(in, out)
	})

	Convey("Removal of the alt attribute", t, func() {
		in := `<html><head></head><body>
<div alt="remove" end="end">text</div>
</body></html>`
		out := `<html><head></head><body>
<div end="end">text</div>
</body></html>`
		minificationCheck(in, out)
	})

	Convey("Removal of the cols attribute", t, func() {
		in := `<html><head></head><body>
<div cols="remove">text</div>
</body></html>`
		out := `<html><head></head><body>
<div>text</div>
</body></html>`
		minificationCheck(in, out)
	})

	Convey("Removal of the class, title and width attribute", t, func() {
		in := `<html><head></head><body>
<div class="remove" begin="begin" title="remove" width="remove" end="end">text</div>
</body></html>`
		out := `<html><head></head><body>
<div begin="begin" end="end">text</div>
</body></html>`
		minificationCheck(in, out)
	})

	Convey("Removal of the align, style, color, disabled and valign attribute", t, func() {
		in := `<html><head></head><body>
<div align="remove" style="remove" begin="begin" color="remove" disabled valign="remove" end="end">text</div>
</body></html>`
		out := `<html><head></head><body>
<div begin="begin" end="end">text</div>
</body></html>`
		minificationCheck(in, out)
	})

	Convey("Removal of all other attribute", t, func() {
		attrs := []string{
			"data-ref",
			"data-url",
			"data-border",
			"target",
			"height",
			"border",
			"hspace",
			"vspace",
			"bgcolor",
			"onclick",
			"colspan",
			"itemprop",
			"itemtype",
			"itemscope",
			"cellspacing",
			"cellpadding",
			"bordercolor",
		}
		out := `<html><head></head><body>
<div>text</div>
</body></html>`
		for _, attr := range attrs {
			in := fmt.Sprintf(`<html><head></head><body>
<div %s="remove">text</div>
</body></html>`, attr)
			minificationCheck(in, out)
		}
	})

}
