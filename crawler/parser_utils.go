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
	result := next
	prevText := prev != nil && prev.Type == html.TextNode
	nextText := next != nil && next.Type == html.TextNode
	delim := ""
	if addSeparator {
		delim = " "
	}

	if prevText && nextText {
		prev.Data = prev.Data + delim + next.Data
		parent.RemoveChild(next)
		result = prev.NextSibling
	} else if prevText {
		prev.Data = prev.Data + delim
	} else if nextText {
		next.Data = delim + next.Data
	} else if addSeparator {
		newNode := &html.Node{
			Parent:      parent,
			FirstChild:  nil,
			LastChild:   nil,
			PrevSibling: prev,
			NextSibling: next,
			Type:        html.TextNode,
			Data:        delim,
			Namespace:   parent.Namespace}
		if prev != nil {
			prev.NextSibling = newNode
		} else {
			parent.FirstChild = newNode
		}
		if next != nil {
			next.PrevSibling = newNode
		} else {
			parent.LastChild = newNode
		}
	}

	return result
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
			Parent:      nil,
			FirstChild:  nil,
			LastChild:   nil,
			PrevSibling: nil,
			NextSibling: nil,
			Type:        html.TextNode,
			Data:        text,
			Namespace:   node.Namespace}
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

	for it := first; it != nil; it = it.NextSibling {
		it.Parent = parent
	}

	if parent.FirstChild == node {
		parent.FirstChild = first
	}
	if parent.LastChild == node {
		parent.LastChild = last
	}

	if prev != nil {
		prev.NextSibling = first
		first.PrevSibling = prev
	}
	if next != nil {
		next.PrevSibling = last
		last.NextSibling = next
	}

	result := u.mergeNodes(parent, prev, first, addSeparator)
	if result != next {
		_ = u.mergeNodes(parent, last, next, addSeparator)
	} else {
		result = u.mergeNodes(parent, prev, next, addSeparator)
	}

	node.Parent = nil
	node.PrevSibling = nil
	node.NextSibling = nil

	return result, nil
}
