package crawler

import (
	"bufio"
	"bytes"
	"errors"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type miniHTML struct {
}

func (m *miniHTML) removeNode(node *html.Node) error {
	node.Parent.RemoveChild(node)

	return nil
}

func (m *miniHTML) parseChildren(node *html.Node) error {
	for it := node.FirstChild; it != nil; {
		currentNode := it
		it = it.NextSibling
		err := m.parseNode(currentNode)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *miniHTML) parseElements(node *html.Node) error {
	switch node.DataAtom {
	case atom.Script:
		return m.removeNode(node)
	case atom.Style:
		return m.removeNode(node)
	case atom.Form:
		return m.removeNode(node)
	case atom.Button:
		return m.removeNode(node)
	default:
	}

	return m.parseChildren(node)
}

func (m *miniHTML) parseNode(node *html.Node) error {
	switch node.Type {
	case html.ErrorNode:
		return errors.New("Found error node in html")
	case html.DocumentNode: // +children -attr (first node)
		return m.parseChildren(node)
	case html.ElementNode: // +children +attr
		return m.parseElements(node)
	case html.TextNode: // -children -attr
		return nil
	case html.DoctypeNode: // ignore
		return nil
	case html.CommentNode:
		node.Parent.RemoveChild(node)
		return nil
	default:
		return errors.New("Unknown node type on html")
	}
}

// Minification - start minification body
func Minification(body []byte) ([]byte, error) {
	startNode, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	parser := miniHTML{}
	err = parser.parseNode(startNode)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	err = html.Render(w, startNode)
	if err != nil {
		return nil, err
	}
	w.Flush()

	return buf.Bytes(), nil
}
