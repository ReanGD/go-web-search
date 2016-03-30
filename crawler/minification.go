package crawler

import (
	"errors"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Minification - struct with functions for minimize html
type Minification struct {
}

func (m *Minification) removeNode(node *html.Node) error {
	node.Parent.RemoveChild(node)

	return nil
}

func (m *Minification) removeTextNode(node *html.Node) error {
	prevCheck := node.PrevSibling == nil ||
		node.PrevSibling.Type != html.TextNode ||
		len(strings.TrimSpace(node.PrevSibling.Data)) == 0
	nextCheck := node.NextSibling == nil ||
		node.NextSibling.Type != html.TextNode ||
		len(strings.TrimSpace(node.NextSibling.Data)) == 0

	if prevCheck || nextCheck {
		node.Parent.RemoveChild(node)
	} else {
		node.Type = html.TextNode
		node.Data = " "
		node.FirstChild = nil
		node.LastChild = nil
	}

	return nil
}

func (m *Minification) parseChildren(node *html.Node) error {
	for it := node.FirstChild; it != nil; {
		currentNode := it
		it = it.NextSibling
		err := m.Run(currentNode)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Minification) parseElements(node *html.Node) error {
	switch node.DataAtom {
	case atom.Script:
		return m.removeNode(node)
	case atom.Style:
		return m.removeNode(node)
	case atom.Form:
		return m.removeTextNode(node)
	case atom.Button:
		return m.removeTextNode(node)
	case atom.Img:
		return m.removeTextNode(node)
	case atom.Time:
		return m.removeTextNode(node)
	case atom.Br:
		return m.removeTextNode(node)
	case atom.Hr:
		return m.removeTextNode(node)
	}

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

	return m.parseChildren(node)
}

// Run - start minification node
func (m *Minification) Run(node *html.Node) error {
	switch node.Type {
	case html.DocumentNode: // +children -attr (first node)
		return m.parseChildren(node)
	case html.ElementNode: // +children +attr
		return m.parseElements(node)
	case html.TextNode: // -children -attr
		return nil
	case html.DoctypeNode: // ignore
		return nil
	case html.CommentNode: // remove
		return m.removeNode(node)
	default:
		return errors.New("minification.Minification.Run: unexpected node type")
	}
}
