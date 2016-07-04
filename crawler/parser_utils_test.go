package crawler

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/html"
)

func helperGetAttrVal(name string, t *testing.T, node *html.Node, result string) {
	Convey(name, t, func() {
		var u parserUtils
		So(u.getAttrVal(node, "key"), ShouldEqual, result)
	})
}

// TestGetAttrVal ...
func TestGetAttrVal(t *testing.T) {
	rightAttr := html.Attribute{Key: "key", Val: "val"}
	rightAttrUpper := html.Attribute{Key: "key", Val: strings.ToUpper("val")}
	wrongAttr := html.Attribute{Key: "err_key", Val: "err_val"}

	helperGetAttrVal("attr is nil", t,
		&html.Node{}, "")

	helperGetAttrVal("attr not found", t,
		&html.Node{Attr: []html.Attribute{wrongAttr}}, "")

	helperGetAttrVal("attr found", t,
		&html.Node{Attr: []html.Attribute{wrongAttr, rightAttr}}, "val")

	helperGetAttrVal("attr found in upper case", t,
		&html.Node{Attr: []html.Attribute{wrongAttr, rightAttrUpper}}, "VAL")

	Convey("getAttrValLower: attr found in upper case", t, func() {
		var u parserUtils
		node := &html.Node{Attr: []html.Attribute{wrongAttr, rightAttrUpper}}
		So(u.getAttrValLower(node, "key"), ShouldEqual, "val")
	})
}

func runMergeNodes(parent, prev, next *html.Node, addSeparator bool) *html.Node {
	var u parserUtils
	if prev != nil {
		parent.AppendChild(prev)
	}
	if next != nil {
		parent.AppendChild(next)
	}
	return u.mergeNodes(parent, prev, next, addSeparator)
}

// TestMergeNodes ...
func TestMergeNodes(t *testing.T) {
	Convey("text text with separator", t, func() {
		parent := &html.Node{}
		prev := &html.Node{Data: "text1", Type: html.TextNode}
		next := &html.Node{Data: "text2", Type: html.TextNode}

		So(runMergeNodes(parent, prev, next, true), ShouldEqual, nil)
		So(parent.FirstChild, ShouldEqual, prev)
		So(parent.FirstChild.NextSibling, ShouldEqual, nil)
		So(prev.Data, ShouldEqual, "text1 text2")
	})

	Convey("text text without separator", t, func() {
		parent := &html.Node{}
		prev := &html.Node{Data: "text1", Type: html.TextNode}
		next := &html.Node{Data: "text2", Type: html.TextNode}

		So(runMergeNodes(parent, prev, next, false), ShouldEqual, nil)
		So(parent.FirstChild, ShouldEqual, prev)
		So(parent.FirstChild.NextSibling, ShouldEqual, nil)
		So(prev.Data, ShouldEqual, "text1text2")
	})

	Convey("text tag with separator", t, func() {
		parent := &html.Node{}
		prev := &html.Node{Data: "text1", Type: html.TextNode}
		next := &html.Node{Type: html.ElementNode}

		So(runMergeNodes(parent, prev, next, true), ShouldEqual, next)
		So(parent.FirstChild, ShouldEqual, prev)
		So(parent.FirstChild.NextSibling, ShouldEqual, next)
		So(prev.Data, ShouldEqual, "text1 ")
	})

	Convey("text node without separator", t, func() {
		parent := &html.Node{}
		prev := &html.Node{Data: "text1", Type: html.TextNode}
		next := &html.Node{Type: html.ElementNode}

		So(runMergeNodes(parent, prev, next, false), ShouldEqual, next)
		So(parent.FirstChild, ShouldEqual, prev)
		So(parent.FirstChild.NextSibling, ShouldEqual, next)
		So(prev.Data, ShouldEqual, "text1")
	})

	Convey("tag text with separator", t, func() {
		parent := &html.Node{}
		prev := &html.Node{Type: html.ElementNode}
		next := &html.Node{Data: "text1", Type: html.TextNode}

		So(runMergeNodes(parent, prev, next, true), ShouldEqual, next)
		So(parent.FirstChild, ShouldEqual, prev)
		So(parent.FirstChild.NextSibling, ShouldEqual, next)
		So(next.Data, ShouldEqual, " text1")
	})

	Convey("tag text without separator", t, func() {
		parent := &html.Node{}
		prev := &html.Node{Type: html.ElementNode}
		next := &html.Node{Data: "text1", Type: html.TextNode}

		So(runMergeNodes(parent, prev, next, false), ShouldEqual, next)
		So(parent.FirstChild, ShouldEqual, prev)
		So(parent.FirstChild.NextSibling, ShouldEqual, next)
		So(next.Data, ShouldEqual, "text1")
	})

	Convey("tag tag with separator", t, func() {
		parent := &html.Node{}
		prev := &html.Node{Type: html.ElementNode}
		next := &html.Node{Type: html.ElementNode}

		So(runMergeNodes(parent, prev, next, true), ShouldEqual, next)
		node1 := parent.FirstChild
		node2 := node1.NextSibling
		node3 := node2.NextSibling
		So(node1, ShouldEqual, prev)
		So(node2.Data, ShouldEqual, " ")
		So(node2.Type, ShouldEqual, html.TextNode)
		So(node3, ShouldEqual, next)
	})

	Convey("tag tag without separator", t, func() {
		parent := &html.Node{}
		prev := &html.Node{Type: html.ElementNode}
		next := &html.Node{Type: html.ElementNode}

		So(runMergeNodes(parent, prev, next, false), ShouldEqual, next)
		node1 := parent.FirstChild
		node2 := node1.NextSibling
		node3 := node2.NextSibling
		So(node1, ShouldEqual, prev)
		So(node2, ShouldEqual, next)
		So(node3, ShouldEqual, nil)
	})

	Convey("nil tag with separator", t, func() {
		parent := &html.Node{}
		var prev *html.Node
		next := &html.Node{Type: html.ElementNode}

		So(runMergeNodes(parent, prev, next, true), ShouldEqual, next)
		node1 := parent.FirstChild
		node2 := node1.NextSibling
		So(node1.Data, ShouldEqual, " ")
		So(node1.Type, ShouldEqual, html.TextNode)
		So(node2, ShouldEqual, next)
	})

	Convey("nil tag with separator", t, func() {
		parent := &html.Node{}
		prev := &html.Node{Type: html.ElementNode}
		var next *html.Node

		So(runMergeNodes(parent, prev, next, true), ShouldEqual, next)
		node1 := parent.FirstChild
		node2 := node1.NextSibling
		node3 := node2.NextSibling
		So(node1, ShouldEqual, prev)
		So(node2.Data, ShouldEqual, " ")
		So(node2.Type, ShouldEqual, html.TextNode)
		So(node3, ShouldEqual, nil)
	})
}

func helperRemoveNode(parent, prev, next *html.Node, isSeparator bool) {
	removed := &html.Node{}
	parent.AppendChild(prev)
	parent.AppendChild(removed)
	parent.AppendChild(next)

	var u parserUtils
	result, err := u.removeNode(removed, isSeparator)
	So(err, ShouldBeNil)
	So(result, ShouldEqual, nil)
}

// TestRemoveNode ...
func TestRemoveNode(t *testing.T) {
	Convey("text text with separator", t, func() {
		parent := &html.Node{}
		prev := &html.Node{Data: "text1", Type: html.TextNode}
		next := &html.Node{Data: "text2", Type: html.TextNode}

		helperRemoveNode(parent, prev, next, true)
		So(parent.FirstChild, ShouldEqual, prev)
		So(parent.FirstChild.NextSibling, ShouldEqual, nil)
		So(prev.Data, ShouldEqual, "text1 text2")
	})

	Convey("text text without separator", t, func() {
		parent := &html.Node{}
		prev := &html.Node{Data: "text1", Type: html.TextNode}
		next := &html.Node{Data: "text2", Type: html.TextNode}

		helperRemoveNode(parent, prev, next, false)
		So(parent.FirstChild, ShouldEqual, prev)
		So(parent.FirstChild.NextSibling, ShouldEqual, nil)
		So(prev.Data, ShouldEqual, "text1text2")
	})
}

// TestAddChildTextNodeToBegining ...
func TestAddChildTextNodeToBegining(t *testing.T) {
	Convey("first child is text", t, func() {
		parent := &html.Node{}
		child1 := &html.Node{Data: "text1", Type: html.TextNode}
		parent.AppendChild(child1)

		var u parserUtils
		u.addChildTextNodeToBegining(parent, "text2")
		So(parent.FirstChild, ShouldEqual, child1)
		So(parent.LastChild, ShouldEqual, child1)
		So(child1.Data, ShouldEqual, "text2text1")
	})

	Convey("without children", t, func() {
		parent := &html.Node{}

		var u parserUtils
		u.addChildTextNodeToBegining(parent, "text1")
		child := parent.FirstChild
		So(parent.FirstChild, ShouldEqual, child)
		So(parent.LastChild, ShouldEqual, child)
		So(child.Data, ShouldEqual, "text1")
	})

	Convey("first child is tag", t, func() {
		parent := &html.Node{}
		child1 := &html.Node{Type: html.ElementNode}
		parent.AppendChild(child1)

		var u parserUtils
		u.addChildTextNodeToBegining(parent, "text1")
		child0 := parent.FirstChild
		So(parent.FirstChild, ShouldEqual, child0)
		So(parent.LastChild, ShouldEqual, child1)
		So(child0.Data, ShouldEqual, "text1")
	})
}

// TestOpenNode ...
func TestOpenNode(t *testing.T) {
	Convey("without children with separator", t, func() {
		parent := &html.Node{}
		node := &html.Node{Type: html.ElementNode}
		parent.AppendChild(node)

		var u parserUtils
		result, err := u.openNode(node, true)
		So(err, ShouldBeNil)
		So(result, ShouldBeNil)
		child := parent.FirstChild
		So(parent.FirstChild, ShouldEqual, child)
		So(parent.LastChild, ShouldEqual, child)
		So(child.Data, ShouldEqual, " ")
	})

	Convey("without children without separator", t, func() {
		parent := &html.Node{}
		node := &html.Node{Type: html.ElementNode}
		parent.AppendChild(node)

		var u parserUtils
		result, err := u.openNode(node, false)
		So(err, ShouldBeNil)
		So(result, ShouldBeNil)
		So(parent.FirstChild, ShouldEqual, nil)
	})

	Convey("parent has one children, node has children", t, func() {
		parent := &html.Node{}
		node := &html.Node{Type: html.ElementNode}
		child1 := &html.Node{Data: "text1", Type: html.TextNode}
		child2 := &html.Node{Data: "text2", Type: html.TextNode}
		child3 := &html.Node{Data: "text3", Type: html.TextNode}
		node.AppendChild(child1)
		node.AppendChild(child2)
		node.AppendChild(child3)
		parent.AppendChild(node)

		var u parserUtils
		result, err := u.openNode(node, true)
		So(err, ShouldBeNil)
		So(result, ShouldEqual, child1)
		ch1 := parent.FirstChild
		ch2 := ch1.NextSibling
		ch3 := ch2.NextSibling
		ch4 := ch3.NextSibling
		So(ch1, ShouldEqual, child1)
		So(ch1.Data, ShouldEqual, " text1")
		So(ch2, ShouldEqual, child2)
		So(ch2.Data, ShouldEqual, "text2")
		So(ch3, ShouldEqual, child3)
		So(ch3.Data, ShouldEqual, "text3 ")
		So(ch4, ShouldEqual, nil)
	})

	Convey("parent has children, node has children", t, func() {
		parent := &html.Node{}
		prev := &html.Node{Data: "text1", Type: html.TextNode}
		node := &html.Node{Type: html.ElementNode}
		next := &html.Node{Data: "text3", Type: html.TextNode}
		child := &html.Node{Data: "text2", Type: html.TextNode}
		node.AppendChild(child)
		parent.AppendChild(prev)
		parent.AppendChild(node)
		parent.AppendChild(next)

		var u parserUtils
		result, err := u.openNode(node, true)
		So(err, ShouldBeNil)
		So(result, ShouldEqual, nil)
		ch0 := parent.FirstChild
		ch1 := ch0.NextSibling
		So(ch0, ShouldEqual, prev)
		So(ch0.Data, ShouldEqual, "text1 text2 text3")
		So(ch1, ShouldEqual, nil)
	})
}
