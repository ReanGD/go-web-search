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

func (m *minification1) removeNodeEx(node *html.Node, addSeparator bool) (*html.Node, error) {
	parent := node.Parent
	prev := node.PrevSibling
	prevText := prev != nil && prev.Type == html.TextNode
	next := node.NextSibling
	nextText := next != nil && next.Type == html.TextNode
	var result *html.Node
	delim := " "
	if !addSeparator {
		delim = ""
	}

	if prevText && nextText {
		prev.Data = strings.TrimRightFunc(prev.Data, unicode.IsSpace) +
			delim + strings.TrimLeftFunc(next.Data, unicode.IsSpace)
		parent.RemoveChild(node)
		parent.RemoveChild(next)
		result = prev.NextSibling
	} else if prevText {
		prev.Data = strings.TrimRightFunc(prev.Data, unicode.IsSpace) + delim
		parent.RemoveChild(node)
		result = prev.NextSibling
	} else if nextText {
		next.Data = delim + strings.TrimLeftFunc(next.Data, unicode.IsSpace)
		parent.RemoveChild(node)
		result = next
	} else if addSeparator {
		node.Type = html.TextNode
		node.Data = delim
		node.FirstChild = nil
		node.LastChild = nil
		result = next
	} else {
		parent.RemoveChild(node)
		result = next
	}

	return result, nil
}

func (m *minification1) removeNode(node *html.Node) (*html.Node, error) {
	return m.removeNodeEx(node, true)
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
	case atom.Button:
		return m.removeNode(node)
	case atom.Object:
		return m.removeNode(node)
	case atom.Select:
		return m.removeNode(node)
	case atom.Style:
		return m.removeNode(node)
	case atom.Param:
		return m.removeNode(node)
	case atom.Embed:
		return m.removeNode(node)
	case atom.Form:
		return m.removeNode(node)
	case atom.Time:
		return m.removeNode(node)
	case atom.Img:
		return m.removeNode(node)
	case atom.Svg:
		return m.removeNode(node)
	case atom.Br:
		return m.removeNode(node)
	case atom.Hr:
		return m.removeNode(node)
	case atom.Wbr:
		return m.removeNodeEx(node, false)
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
			return m.removeNode(node)
		}
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
