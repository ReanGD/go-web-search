package crawler

import (
	"bytes"
	"container/list"
	"errors"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type pageURLs struct {
	LinkList list.List
}

func (result *pageURLs) getAttrVal(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}

	return ""
}

func (result *pageURLs) parseChildren(node *html.Node) {
	for it := node.FirstChild; it != nil; it = it.NextSibling {
		result.parseNode(it)
	}
}

func (result *pageURLs) parseElements(node *html.Node) {
	switch node.DataAtom {
	case atom.A, atom.Area:
		if link := result.getAttrVal(node, "href"); link != "" {
			result.LinkList.PushBack(link)
		}
		break
	case atom.Frame:
		if link := result.getAttrVal(node, "src"); link != "" {
			result.LinkList.PushBack(link)
		}
	}

	result.parseChildren(node)
}

func (result *pageURLs) parseNode(node *html.Node) error {
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

func (result *pageURLs) Parse(body []byte) (*list.List, error) {
	node, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	err = result.parseNode(node)

	return &result.LinkList, err
}
