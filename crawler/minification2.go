package crawler

import (
	"bytes"
	"errors"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/text/unicode/rangetable"
)

var (
	// ErrMinification2UnexpectedNodeType - found unexpected node type
	ErrMinification2UnexpectedNodeType = errors.New("minification2.minification2.parseNode: unexpected node type")
)

var notSeparatorRT = rangetable.New(
	'&', '-', '@', '_', '+', '\'',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k',
	'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v',
	'w', 'x', 'y', 'z',
	'а', 'б', 'в', 'г', 'д', 'е', 'ё', 'ж', 'з', 'и', 'й',
	'к', 'л', 'м', 'н', 'о', 'п', 'р', 'с', 'т', 'у', 'ф',
	'х', 'ц', 'ч', 'ш', 'щ', 'ъ', 'ы', 'ь', 'э', 'ю', 'я')

type minification2 struct {
}

func (m *minification2) processText(in string) string {
	var buffer bytes.Buffer
	var rRaw, r rune
	var size int
	prevIsSeparator := false
	prevRune := ' '
	isFirst := true
	for len(in) > 0 {
		rRaw, size = utf8.DecodeRuneInString(in)
		r = unicode.ToLower(rRaw)

		if unicode.Is(notSeparatorRT, r) {
			if prevIsSeparator {
				prevIsSeparator = false
				if !isFirst {
					_ = buffer.WriteByte(' ')
				}
			}
			_, _ = buffer.WriteRune(r)
			isFirst = false
		} else {
			if !prevIsSeparator {
				prevIsSeparator = true
				// digits
				if '0' <= prevRune && prevRune <= '9' && (r == '.' || r == ',' || r == ':') {
					rRaw, _ = utf8.DecodeRuneInString(in[size:])
					if '0' <= rRaw && rRaw <= '9' {
						_, _ = buffer.WriteRune(r)
						prevIsSeparator = false
					}
				}
			} else {
				prevIsSeparator = true
			}
		}

		prevRune = r
		in = in[size:]
	}

	return buffer.String()
}

func (m *minification2) parseText(node *html.Node) (*html.Node, error) {
	next := node.NextSibling
	text := m.processText(node.Data)
	if len(text) != 0 {
		node.Data = text
	} else {
		node.Parent.RemoveChild(node)
	}
	return next, nil
}

func (m *minification2) parseChildren(node *html.Node) (*html.Node, error) {
	var err error
	for it := node.FirstChild; it != nil; {
		it, err = m.parseNode(it)
		if err != nil {
			return node.NextSibling, err
		}
	}

	return node.NextSibling, nil
}

func (m *minification2) parseElements(node *html.Node) (*html.Node, error) {
	switch node.DataAtom {
	case atom.Body, atom.Head, atom.Html:
		return m.parseChildren(node)
	case atom.Div:
		next, err := m.parseChildren(node)
		if err == nil && node.FirstChild == nil {
			node.Parent.RemoveChild(node)
		}
		return next, err
	default:
		return nil, ErrMinification2UnexpectedNodeType
	}
}

func (m *minification2) parseNode(node *html.Node) (*html.Node, error) {
	switch node.Type {
	case html.DocumentNode: // +children -attr (first node)
		return m.parseChildren(node)
	case html.ElementNode: // +children +attr
		return m.parseElements(node)
	case html.TextNode: // -children -attr
		return m.parseText(node)
	default: // ErrorNode, CommentNode, DoctypeNode
		return nil, ErrMinification2UnexpectedNodeType
	}
}

// RunMinification2 - start minification node
func RunMinification2(node *html.Node) error {
	m := minification2{}
	_, err := m.parseNode(node)
	return err
}
