package crawler

import (
	"errors"
	"strings"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	// ErrMinificationUnexpectedNodeType - found unexpected node type
	ErrMinificationUnexpectedNodeType = errors.New("minification1.minification1.parseNode: unexpected node type")
)

type minification1 struct {
}

func (m *minification1) removeAttr(node *html.Node) {
	lenAttr := len(node.Attr)
	if lenAttr != 0 {
		attr := node.Attr
		i := 0
		j := 0
		for ; i != lenAttr; i++ {
			key := strings.ToLower(attr[i].Key)
			if strings.HasPrefix(key, "data-") {
				continue
			}
			switch key {
			case "id":
			case "alt":
			case "cols":
			case "class":
			case "title":
			case "width":
			case "align":
			case "style":
			case "color":
			case "valign":
			case "target":
			case "height":
			case "border":
			case "hspace":
			case "vspace":
			case "bgcolor":
			case "onclick":
			case "colspan":
			case "itemprop":
			case "disabled":
			case "itemtype":
			case "itemscope":
			case "cellspacing":
			case "cellpadding":
			case "bordercolor":
			default:
				if i != j {
					attr[j] = attr[i]
				}
				j++
			}
		}
		if i != j {
			node.Attr = attr[:j]
		}
	}
}

func (m *minification1) getAttrVal(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}

	return ""
}

func (m *minification1) getAttrValLower(node *html.Node, attrName string) string {
	return strings.ToLower(m.getAttrVal(node, attrName))
}

func (m *minification1) mergeNodes(parent, prev, next *html.Node, addSeparator bool) *html.Node {
	var result *html.Node
	prevText := prev != nil && prev.Type == html.TextNode
	nextText := next != nil && next.Type == html.TextNode
	delim := ""
	if addSeparator {
		delim = " "
	}

	if prevText && nextText {
		prev.Data = strings.TrimRightFunc(prev.Data, unicode.IsSpace) +
			delim + strings.TrimLeftFunc(next.Data, unicode.IsSpace)
		parent.RemoveChild(next)
		result = prev.NextSibling
	} else if prevText {
		prev.Data = strings.TrimRightFunc(prev.Data, unicode.IsSpace) + delim
		result = next
	} else if nextText {
		next.Data = delim + strings.TrimLeftFunc(next.Data, unicode.IsSpace)
		result = next
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
		result = next
	} else {
		result = next
	}

	return result
}

func (m *minification1) removeNode(node *html.Node, addSeparator bool) (*html.Node, error) {
	prev, next, parent := node.PrevSibling, node.NextSibling, node.Parent
	parent.RemoveChild(node)
	result := m.mergeNodes(parent, prev, next, addSeparator)

	return result, nil
}

func (m *minification1) openNode(node *html.Node, addSeparator bool) (*html.Node, error) {
	parent := node.Parent
	first, last := node.FirstChild, node.LastChild
	prev, next := node.PrevSibling, node.NextSibling

	if first == nil {
		return m.removeNode(node, addSeparator)
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
		first.PrevSibling = prev
		prev.NextSibling = first
	}
	if next != nil {
		last.NextSibling = next
		next.PrevSibling = last
	}

	node.Parent = nil
	node.PrevSibling = nil
	node.NextSibling = nil

	result := m.mergeNodes(parent, prev, first, addSeparator)
	result = m.mergeNodes(parent, last, next, addSeparator)

	return result, nil
}

func (m *minification1) parseChildren(node *html.Node) (*html.Node, error) {
	var err error
	for it := node.FirstChild; it != nil; {
		it, err = m.parseNode(it)
		if err != nil {
			return node.NextSibling, err
		}
	}

	return node.NextSibling, nil
}

func (m *minification1) parseElements(node *html.Node) (*html.Node, error) {
	switch node.DataAtom {
	case atom.Script:
		return m.removeNode(node, true)
	case atom.Button:
		return m.removeNode(node, true)
	case atom.Object:
		return m.removeNode(node, true)
	case atom.Select:
		return m.removeNode(node, true)
	case atom.Style:
		return m.removeNode(node, true)
	case atom.Param:
		return m.removeNode(node, true)
	case atom.Embed:
		return m.removeNode(node, true)
	case atom.Form:
		return m.removeNode(node, true)
	case atom.Time:
		return m.removeNode(node, true)
	case atom.Img:
		return m.removeNode(node, true)
	case atom.Svg:
		return m.removeNode(node, true)
	case atom.Br:
		return m.removeNode(node, true)
	case atom.Hr:
		return m.removeNode(node, true)
	case atom.Wbr:
		return m.removeNode(node, false)
	case atom.Input:
		typeInput := m.getAttrValLower(node, "type")
		if typeInput == "radio" ||
			typeInput == "checkbox" ||
			typeInput == "hidden" ||
			typeInput == "button" ||
			typeInput == "submit" ||
			typeInput == "reset" ||
			typeInput == "file" ||
			typeInput == "image" {
			return m.removeNode(node, true)
		}
	}

	switch node.DataAtom {
	case atom.B:
		return m.openNode(node, false)
	case atom.I:
		return m.openNode(node, false)
	}

	m.removeAttr(node)

	return m.parseChildren(node)
}

func (m *minification1) parseNode(node *html.Node) (*html.Node, error) {
	switch node.Type {
	case html.DocumentNode: // +children -attr (first node)
		return m.parseChildren(node)
	case html.ElementNode: // +children +attr
		return m.parseElements(node)
	case html.TextNode: // -children -attr
		return node.NextSibling, nil
	case html.DoctypeNode: // ignore
		return node.NextSibling, nil
	case html.CommentNode: // remove
		return m.removeNode(node, true)
	default:
		return nil, ErrMinificationUnexpectedNodeType
	}
}

// RunMinification1 - start minification node
func RunMinification1(node *html.Node) error {
	m := minification1{}
	_, err := m.parseNode(node)
	return err
}
