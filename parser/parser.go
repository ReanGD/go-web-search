package parser

import (
	"container/list"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/kljensen/snowball"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var regexRus = regexp.MustCompile(`[а-яё][0-9а-яё]*`)

// ParseResult - parse result
type ParseResult struct {
	Data list.List
}

// ParseURL - parse html page by URL
func ParseURL(url string) (*ParseResult, error) {
	p := new(ParseResult)
	err := p.parseURL(url)

	return p, err
}

// ParseStream - parse html page from io.Reader
func ParseStream(reader io.Reader) (*ParseResult, error) {
	p := new(ParseResult)
	err := p.parseStream(reader)

	return p, err
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
		if attrVal := getAttrVal(node, "href"); attrVal != "" {
			attrVal = normalizeHrefLink(attrVal)
			if attrVal != "" {
				// fmt.Println("Link=", attrVal)
			}
		}
		break
	case atom.Style, atom.Link, atom.Script, atom.Noscript, atom.Meta:
		return
	default:
	}
	// fmt.Println("tag=", node.Data)

	result.parseChildren(node)
}

func (result *ParseResult) parseText(text string) {
	if len(text) > 2 {
		worlds := regexRus.FindAllString(strings.ToLower(text), -1)
		for i := 0; i != len(worlds); i++ {
			if len(worlds[i]) > 2 {
				// self.data.PushBack(worlds[i])
				stemmed, err := snowball.Stem(worlds[i], "russian", true)
				if err == nil {
					result.Data.PushBack(stemmed)
				} else {
					fmt.Println("error:", worlds[i])
				}
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
