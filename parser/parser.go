package parser

import (
	"container/list"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var regexRus = regexp.MustCompile(`[а-яё][0-9а-яё]*`)

// ParseResult - parse result
type ParseResult struct {
	Words list.List
	Links list.List
}

// ParseURL - parse html page by URL
func ParseURL(url string) (*ParseResult, error) {
	result := new(ParseResult)
	err := result.parseURL(url)

	return result, err
}

// ParseStream - parse html page from io.Reader
func ParseStream(reader io.Reader) (*ParseResult, error) {
	result := new(ParseResult)
	err := result.parseStream(reader)

	return result, err
}

func normalizeHrefLink(link string) string {
	link = strings.TrimSpace(link)
	link = strings.TrimPrefix(link, "mailto:")
	return link
}

func getAttrVal(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}

	return ""
}

func (result *ParseResult) parseChildren(node *html.Node) {
	for it := node.FirstChild; it != nil; it = it.NextSibling {
		result.parseNode(it)
	}
}

func (result *ParseResult) parseElements(node *html.Node) {
	switch node.DataAtom {
	case atom.A:
		if link := getAttrVal(node, "href"); link != "" {
			link = normalizeHrefLink(link)
			if link != "" {
				result.Links.PushBack(link)
			}
		}
		break
	// skip
	case atom.Style, atom.Link, atom.Script, atom.Noscript, atom.Meta:
		return
	default:
	}

	result.parseChildren(node)
}

func (result *ParseResult) parseText(text string) {
	if len(text) > 2 {
		worlds := regexRus.FindAllString(strings.ToLower(text), -1)
		for i := 0; i != len(worlds); i++ {
			if len(worlds[i]) > 2 {
				result.Words.PushBack(worlds[i])
			}
		}
	}
}

func (result *ParseResult) parseNode(node *html.Node) error {
	switch node.Type {
	case html.ErrorNode:
		return errors.New("ErrorNode on html")
	case html.TextNode:
		result.parseText(node.Data)
		return nil
	case html.DocumentNode:
		result.parseChildren(node)
		return nil
	case html.ElementNode:
		result.parseElements(node)
		return nil
	case html.CommentNode, html.DoctypeNode: // skip
		return nil
	default:
		return errors.New("Unknown node type on html")
	}
}

func (result *ParseResult) parseStream(reader io.Reader) error {
	node, err := html.Parse(reader)
	if err != nil {
		return err
	}
	return result.parseNode(node)
}

func (result *ParseResult) parseURL(url string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return result.parseStream(response.Body)
}
