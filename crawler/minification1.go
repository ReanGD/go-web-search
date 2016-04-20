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

func (m *minification1) removeNode(node *html.Node) (*html.Node, error) {
	parent := node.Parent
	prev := node.PrevSibling
	prevText := prev != nil && prev.Type == html.TextNode
	next := node.NextSibling
	nextText := next != nil && next.Type == html.TextNode
	var result *html.Node

	if prevText && nextText {
		prev.Data = strings.TrimRightFunc(prev.Data, unicode.IsSpace) +
			" " + strings.TrimLeftFunc(next.Data, unicode.IsSpace)
		parent.RemoveChild(node)
		parent.RemoveChild(next)
		result = prev.NextSibling
	} else if prevText {
		prev.Data = strings.TrimRightFunc(prev.Data, unicode.IsSpace) + " "
		parent.RemoveChild(node)
		result = prev.NextSibling
	} else if nextText {
		next.Data = " " + strings.TrimLeftFunc(next.Data, unicode.IsSpace)
		parent.RemoveChild(node)
		result = next
	} else {
		node.Type = html.TextNode
		node.Data = " "
		node.FirstChild = nil
		node.LastChild = nil
		result = next
	}

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
		return m.removeNode(node)
	case atom.Style:
		return m.removeNode(node)
	case atom.Form:
		return m.removeNode(node)
	case atom.Button:
		return m.removeNode(node)
	case atom.Img:
		return m.removeNode(node)
	case atom.Time:
		return m.removeNode(node)
	case atom.Br:
		return m.removeNode(node)
	case atom.Hr:
		return m.removeNode(node)
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
		return m.removeNode(node)
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
