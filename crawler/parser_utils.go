package crawler

// status: ok
import (
	"strings"

	"golang.org/x/net/html"
)

type parserUtils struct {
}

func (u *parserUtils) getAttrValLower(node *html.Node, attrName string) string {
	if node.Attr == nil {
		return ""
	}

	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return strings.ToLower(attr.Val)
		}
	}

	return ""
}

func (u *parserUtils) mergeNodes(parent, prev, next *html.Node, addSeparator bool) *html.Node {
	prevText := prev != nil && prev.Type == html.TextNode
	nextText := next != nil && next.Type == html.TextNode
	delim := ""
	if addSeparator {
		delim = " "
	}

	if prevText && nextText {
		prev.Data = prev.Data + delim + next.Data
		parent.RemoveChild(next)
		return prev.NextSibling
	}

	if prevText {
		prev.Data = prev.Data + delim
	} else if nextText {
		next.Data = delim + next.Data
	} else if addSeparator {
		newNode := &html.Node{
			Type: html.TextNode,
			Data: delim}
		parent.InsertBefore(newNode, next)
	}

	return next
}

func (u *parserUtils) removeNode(node *html.Node, addSeparator bool) (*html.Node, error) {
	prev, next, parent := node.PrevSibling, node.NextSibling, node.Parent
	parent.RemoveChild(node)

	return u.mergeNodes(parent, prev, next, addSeparator), nil
}

func (u *parserUtils) addChildTextNodeToBegining(node *html.Node, text string) {
	if node.FirstChild != nil && node.FirstChild.Type == html.TextNode {
		node.FirstChild.Data = text + node.FirstChild.Data
	} else {
		newNode := &html.Node{
			Type: html.TextNode,
			Data: text}
		if node.FirstChild == nil {
			node.AppendChild(newNode)
		} else {
			node.InsertBefore(newNode, node.FirstChild)
		}
	}
}

func (u *parserUtils) openNode(node *html.Node, addSeparator bool) (*html.Node, error) {
	parent := node.Parent
	first, last := node.FirstChild, node.LastChild
	prev, next := node.PrevSibling, node.NextSibling

	if first == nil {
		return u.removeNode(node, addSeparator)
	}

	for it := first; it != nil; {
		cur := it
		it = it.NextSibling

		cur.Parent = nil
		cur.PrevSibling = nil
		cur.NextSibling = nil
		parent.InsertBefore(cur, next)
	}

	parent.RemoveChild(node)

	result := u.mergeNodes(parent, prev, first, addSeparator)
	if result != next {
		_ = u.mergeNodes(parent, last, next, addSeparator)
	} else {
		result = u.mergeNodes(parent, prev, next, addSeparator)
	}

	return result, nil
}
