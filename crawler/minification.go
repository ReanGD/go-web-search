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
		return m.removeNode(node)
	case atom.Button:
		return m.removeNode(node)
	case atom.Time:
		return m.removeNode(node)
	}

	len := len(node.Attr)
	if len != 0 {
		attr := node.Attr
		i := 0
		j := 0
		for ; i != len; i++ {
			switch strings.ToLower(attr[i].Key) {
			case "id":
			case "style":
			case "onclick":
			case "target":
			case "title":
			case "class":
			case "width":
			case "height":
			case "alt":
			case "disabled":
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
