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
	ErrMinificationUnexpectedNodeType = errors.New("minification.Minification.parseNode: unexpected node type")
)

// Minification - struct with functions for minimize html
type Minification struct {
}

func (m *Minification) removeAttr(node *html.Node) {
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

func (m *Minification) removeNode(node *html.Node) (*html.Node, error) {
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

func (m *Minification) parseChildren(node *html.Node) (*html.Node, error) {
	var err error
	for it := node.FirstChild; it != nil; {
		it, err = m.parseNode(it)
		if err != nil {
			return node.NextSibling, err
		}
	}

	return node.NextSibling, nil
}

func (m *Minification) parseElements(node *html.Node) (*html.Node, error) {
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

func (m *Minification) parseNode(node *html.Node) (*html.Node, error) {
	switch node.Type {
	case html.DocumentNode: // +children -attr (first node)
		return m.parseChildren(node)
	case html.ElementNode: // +children +attr
		return m.parseElements(node)
	case html.TextNode: // -children -attr
		return nil, nil
	case html.DoctypeNode: // ignore
		return nil, nil
	case html.CommentNode: // remove
		return m.removeNode(node)
	default:
		return nil, ErrMinificationUnexpectedNodeType
	}
}

// Run - start minification node
func (m *Minification) Run(node *html.Node) error {
	_, err := parseNode(node)
	return err
}
