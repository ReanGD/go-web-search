package crawler

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	// ErrMinificationUnexpectedNodeType - found unexpected node type
	ErrMinificationUnexpectedNodeType = errors.New("minification1.minification1.parseNode: unexpected node type")
)

type minification1 struct {
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

func (m *minification1) mergeNodes(parent, prev, next *html.Node, addSeparator bool) *html.Node {
	var result *html.Node
	prevText := prev != nil && prev.Type == html.TextNode
	nextText := next != nil && next.Type == html.TextNode
	delim := ""
	if addSeparator {
		delim = " "
	}

	if prevText && nextText {
		if !addSeparator {
			r, _ := utf8.DecodeLastRuneInString(prev.Data)
			if unicode.IsSpace(r) {
				delim = " "
			} else {
				r, _ = utf8.DecodeRuneInString(next.Data)
				if unicode.IsSpace(r) {
					delim = " "
				}
			}
		}

		prev.Data = strings.TrimRightFunc(prev.Data, unicode.IsSpace) +
			delim + strings.TrimLeftFunc(next.Data, unicode.IsSpace)
		parent.RemoveChild(next)
		result = prev.NextSibling
	} else if prevText {
		if !addSeparator {
			r, _ := utf8.DecodeLastRuneInString(prev.Data)
			if unicode.IsSpace(r) {
				delim = " "
			}
		}

		prev.Data = strings.TrimRightFunc(prev.Data, unicode.IsSpace) + delim
		result = next
	} else if nextText {
		if !addSeparator {
			r, _ := utf8.DecodeRuneInString(next.Data)
			if unicode.IsSpace(r) {
				delim = " "
			}
		}

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

func (m *minification1) AddChildTextNodeToBegining(node *html.Node, text string) {
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
		prev.NextSibling = first
		first.PrevSibling = prev
	}
	if next != nil {
		next.PrevSibling = last
		last.NextSibling = next
	}

	result := m.mergeNodes(parent, prev, first, addSeparator)
	if result != next {
		_ = m.mergeNodes(parent, last, next, addSeparator)
	} else {
		result = m.mergeNodes(parent, prev, next, addSeparator)
	}

	node.Parent = nil
	node.PrevSibling = nil
	node.NextSibling = nil

	return result, nil
}

func (m *minification1) toDiv(node *html.Node) {
	node.DataAtom = atom.Div
	node.Data = "div"
	node.Attr = make([]html.Attribute, 0)
}

func (m *minification1) toRef(node *html.Node, ref string) (*html.Node, error) {
	node.DataAtom = atom.A
	node.Data = "a"
	node.FirstChild = nil
	node.LastChild = nil
	node.Attr = make([]html.Attribute, 1)
	node.Attr[0] = html.Attribute{Key: "href", Val: ref}

	prev, next := node.PrevSibling, node.NextSibling
	_ = m.mergeNodes(node.Parent, prev, node, true)
	return m.mergeNodes(node.Parent, node, next, true), nil
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
	// case atom.A:
	case atom.Abbr:
		title := m.getAttrValLower(node, "title")
		if title != "" {
			m.AddChildTextNodeToBegining(node, " "+title+" ")
		}
		return m.openNode(node, true)
	// case atom.Acronym:
	// case atom.Address:
	// case atom.Applet:
	// case atom.Area:
	// case atom.Article:
	// case atom.Aside:
	// case atom.Audio:
	case atom.B:
		return m.openNode(node, false)
	// case atom.Base:
	// case atom.Basefont:
	// case atom.Bdi:
	// case atom.Bdo:
	// case atom.Bgsound:
	case atom.Blockquote:
		m.toDiv(node)
	// case atom.Big:
	// case atom.Body:
	// case atom.Blink:
	case atom.Br:
		return m.removeNode(node, true)
	case atom.Button:
		return m.removeNode(node, true)
	case atom.Canvas:
		return m.removeNode(node, true)
	case atom.Caption:
		return m.openNode(node, true)
	case atom.Center:
		m.toDiv(node)
	case atom.Cite:
		return m.openNode(node, false)
	case atom.Code:
		return m.openNode(node, false)
	case atom.Colgroup, atom.Col:
		return m.removeNode(node, true)
	// case atom.Command:
	// case atom.Comment:
	// case atom.Data:
	// case atom.Datalist:
	case atom.Dd:
		return m.openNode(node, true)
	case atom.Del:
		return m.openNode(node, false)
	// case atom.Details:
	// case atom.Dfn:
	// case atom.Dialog:
	// case atom.Dir:
	// case atom.Div:
	case atom.Dl:
		m.toDiv(node)
	case atom.Dt:
		return m.openNode(node, true)
	case atom.Em:
		return m.openNode(node, false)
	case atom.Embed:
		return m.removeNode(node, true)
	// case atom.Fieldset:
	case atom.Figcaption:
		return m.openNode(node, true)
	case atom.Figure:
		m.toDiv(node)
	case atom.Font:
		return m.openNode(node, false)
	// case atom.Footer:
	case atom.Form:
		return m.removeNode(node, true)
	// case atom.Frame:
	// case atom.Frameset:
	// case atom.H1:
	// case atom.H2:
	// case atom.H3:
	// case atom.H4:
	// case atom.H5:
	// case atom.H6:
	// case atom.Head:
	// case atom.Header:
	// case atom.Hgroup:
	case atom.Hr:
		return m.removeNode(node, true)
	// case atom.Html:
	case atom.I:
		return m.openNode(node, false)
	case atom.Iframe:
		ref := m.getAttrValLower(node, "src")
		if ref == "" {
			return m.removeNode(node, true)
		}
		return m.toRef(node, ref)
	case atom.Img:
		return m.removeNode(node, true)
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
	case atom.Ins:
		return m.openNode(node, false)
	// case atom.Isindex:
	// case atom.Kbd:
	// case atom.Keygen:
	// case atom.Label:
	// case atom.Legend:
	case atom.Li:
		return m.openNode(node, true)
	// case atom.Link:
	// case atom.Listing:
	// case atom.Main:
	// case atom.Map:
	// case atom.Marquee:
	// case atom.Mark:
	// case atom.Menu:
	// case atom.Menuitem:
	// case atom.Meta:
	// case atom.Meter:
	// case atom.Multicol:
	case atom.Nav:
		m.toDiv(node)
	case atom.Nobr:
		return m.openNode(node, false)
		// case atom.Noembed:
		// case atom.Noindex:
		// case atom.Noframes:
		// case atom.Noscript:
	case atom.Object:
		return m.removeNode(node, true)
	case atom.Ol:
		m.toDiv(node)
		// case atom.Optgroup:
		// case atom.Option:
		// case atom.Output:
		// case atom.P:
	case atom.Param:
		return m.removeNode(node, true)
	// case atom.Picture:
	// case atom.Plaintext:
	case atom.Pre:
		m.toDiv(node)
		// case atom.Progress:
	case atom.Q:
		return m.openNode(node, false)
		// case atom.Rp:
		// case atom.Rt:
		// case atom.Rtc:
		// case atom.Ruby:
	case atom.S:
		return m.openNode(node, false)
	// case atom.Samp:
	case atom.Script:
		return m.removeNode(node, true)
	case atom.Section:
		m.toDiv(node)
	case atom.Select:
		return m.removeNode(node, true)
	case atom.Small:
		return m.openNode(node, false)
	// case atom.Source:
	// case atom.Spacer:
	case atom.Span:
		return m.openNode(node, false)
	case atom.Strike:
		return m.openNode(node, false)
	case atom.Strong:
		return m.openNode(node, false)
	case atom.Style:
		return m.removeNode(node, true)
	case atom.Sub:
		return m.openNode(node, false)
		// case atom.Summary:
	case atom.Sup:
		return m.openNode(node, false)
	case atom.Svg:
		return m.removeNode(node, true)
	case atom.Table:
		m.toDiv(node)
	case atom.Tbody:
		return m.openNode(node, true)
	case atom.Td:
		m.toDiv(node)
	case atom.Textarea:
		return m.removeNode(node, true)
	case atom.Tfoot:
		return m.openNode(node, true)
	case atom.Th:
		m.toDiv(node)
	case atom.Thead:
		return m.openNode(node, true)
	case atom.Time:
		return m.openNode(node, false)
	// case atom.Title:
	case atom.Tr:
		return m.openNode(node, true)
	// case atom.Track:
	case atom.Tt:
		return m.openNode(node, false)
	case atom.U:
		return m.openNode(node, false)
	case atom.Ul:
		m.toDiv(node)
	case atom.Var:
		return m.openNode(node, false)
	case atom.Video:
		return m.removeNode(node, true)
	case atom.Wbr:
		return m.removeNode(node, false)
		// case atom.Xmp:
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
