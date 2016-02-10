package parser

import (
	"container/list"
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

type ParserData struct {
	Data list.List
}

func NewParser() *ParserData {
	p := new(ParserData)

	return p
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

func (parser *ParserData) parseChildren(node *html.Node) {
	for it := node.FirstChild; it != nil; it = it.NextSibling {
		parser.parseNode(it)
	}
}

func (parser *ParserData) parseElements(node *html.Node) {
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

	parser.parseChildren(node)
}

func (parser *ParserData) parseText(text string) {
	if len(text) > 2 {
		worlds := regexRus.FindAllString(strings.ToLower(text), -1)
		for i := 0; i != len(worlds); i++ {
			if len(worlds[i]) > 2 {
				// self.data.PushBack(worlds[i])
				stemmed, err := snowball.Stem(worlds[i], "russian", true)
				if err == nil {
					parser.Data.PushBack(stemmed)
				} else {
					fmt.Println("error:", worlds[i])
				}
			}
		}
	}
}

func (parser *ParserData) parseNode(node *html.Node) {
	switch node.Type {
	case html.ErrorNode:
		fmt.Println("!!!ErrorNode")
		return
	case html.TextNode:
		parser.parseText(node.Data)
		return
	case html.DocumentNode:
		parser.parseChildren(node)
		return
	case html.ElementNode:
		parser.parseElements(node)
		return
	case html.CommentNode, html.DoctypeNode: // skip
		return
	}
}

func (parser *ParserData) parseReader(reader io.Reader) error {
	node, err := html.Parse(reader)
	if err != nil {
		return err
	}
	parser.parseNode(node)

	return nil
}

func (parser *ParserData) ParseURL(url string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return parser.parseReader(response.Body)
}
