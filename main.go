package main

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

type Parser struct {
	data     list.List
	regexRus *regexp.Regexp
}

func newParser() *Parser {
	p := new(Parser)
	p.regexRus = regexp.MustCompile(`[а-яё][0-9а-яё]*`)

	return p
}

func NormalizeHrefLink(link string) string {
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

func (self *Parser) ParseChildren(node *html.Node) {
	for it := node.FirstChild; it != nil; it = it.NextSibling {
		self.ParseNode(it)
	}
}

func (self *Parser) ParseElements(node *html.Node) {
	switch node.DataAtom {
	case atom.A:
		if attrVal := getAttrVal(node, "href"); attrVal != "" {
			attrVal = NormalizeHrefLink(attrVal)
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

	self.ParseChildren(node)
}

func (self *Parser) ParseText(text string) {
	if len(text) > 2 {
		worlds := self.regexRus.FindAllString(strings.ToLower(text), -1)
		for i := 0; i != len(worlds); i++ {
			if len(worlds[i]) > 2 {
				// self.data.PushBack(worlds[i])
				stemmed, err := snowball.Stem(worlds[i], "russian", true)
				if err == nil {
					self.data.PushBack(stemmed)
				} else {
					fmt.Println("error:", worlds[i])
				}
			}
		}
	}
}

func (self *Parser) ParseNode(node *html.Node) {
	switch node.Type {
	case html.ErrorNode:
		fmt.Println("!!!ErrorNode")
		return
	case html.TextNode:
		self.ParseText(node.Data)
		return
	case html.DocumentNode:
		self.ParseChildren(node)
		return
	case html.ElementNode:
		self.ParseElements(node)
		return
	case html.CommentNode, html.DoctypeNode: // skip
		return
	}
}

func (self *Parser) ParseReader(reader io.Reader) error {
	node, err := html.Parse(reader)
	if err != nil {
		return err
	}
	self.ParseNode(node)

	return nil
}

func (self *Parser) ParseUrl(url string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return self.ParseReader(response.Body)
}

func main() {
	parser := newParser()
	err := parser.ParseUrl("http://habrahabr.ru/")
	// err := ParseUrl("https://www.linux.org.ru/")
	// err := ParseUrl("http://example.com/")
	// s := `<p>жаба</p>`
	// err := parser.ParseReader(strings.NewReader(s))
	if err != nil {
		fmt.Printf("%s", err)
		return
	}
	for e := parser.data.Front(); e != nil; e = e.Next() {
		fmt.Println(e.Value)
	}
	// r, err := zip.OpenReader("testdata/readme.zip")
}
