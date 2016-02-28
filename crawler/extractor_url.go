package crawler

import (
	"container/list"
	"errors"
	"io"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// PageLinks - parse result with list of links
type PageLinks struct {
	LinkList list.List
}

func getAttrVal(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}

	return ""
}

func (result *PageLinks) parseChildren(node *html.Node) {
	for it := node.FirstChild; it != nil; it = it.NextSibling {
		result.parseNode(it)
	}
}

func (result *PageLinks) parseElements(node *html.Node) {
	switch node.DataAtom {
	case atom.A:
		if link := getAttrVal(node, "href"); link != "" {
			result.LinkList.PushBack(link)
		}
		break
		// skip
	case atom.Style, atom.Link, atom.Script, atom.Noscript, atom.Meta:
		return
	default:
	}

	result.parseChildren(node)
}

func (result *PageLinks) parseNode(node *html.Node) error {
	switch node.Type {
	case html.ErrorNode:
		return errors.New("ErrorNode on html")
	case html.DocumentNode:
		result.parseChildren(node)
		return nil
	case html.ElementNode:
		result.parseElements(node)
		return nil
	case html.TextNode, html.CommentNode, html.DoctypeNode: // skip
		return nil
	default:
		return errors.New("Unknown node type on html")
	}
}

// ParseURLsInPage - parse html page from io.Reader
func ParseURLsInPage(reader io.Reader) (*PageLinks, error) {
	node, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}

	result := new(PageLinks)
	err = result.parseNode(node)

	return result, err
}
