package crawler

import (
	"bytes"
	"errors"
	"log"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// HTMLParser - find metatags and links on page
type HTMLParser struct {
	// [URL]hostname
	URLs          map[string]string
	baseURL       *url.URL
	MetaTagIndex  bool
	MetaTagFollow bool
}

func (result *HTMLParser) getAttrVal(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}

	return ""
}

func (result *HTMLParser) getAttrValLower(node *html.Node, attrName string) string {
	return strings.ToLower(result.getAttrVal(node, attrName))
}

func (result *HTMLParser) processLink(link string) {
	if link == "" {
		return
	}
	relative, err := url.Parse(strings.TrimSpace(link))
	if err != nil {
		log.Printf("ERROR: Parse URL on page %s, message: %s", result.baseURL, err)
		return
	}

	parsed := result.baseURL.ResolveReference(relative)
	urlStr := NormalizeURL(parsed)
	parsed, err = url.Parse(urlStr)

	if (parsed.Scheme == "http" || parsed.Scheme == "https") && urlStr != result.baseURL.String() {
		result.URLs[urlStr] = NormalizeHostName(parsed.Host)
	}
}

func (result *HTMLParser) parseChildren(node *html.Node) {
	for it := node.FirstChild; it != nil; it = it.NextSibling {
		result.parseNode(it)
	}
}

func (result *HTMLParser) parseElements(node *html.Node) {
	switch node.DataAtom {
	case atom.A, atom.Area:
		if result.MetaTagFollow {
			rel := result.getAttrValLower(node, "rel")
			if rel != "nofollow" {
				result.processLink(result.getAttrVal(node, "href"))
			}
		}
	case atom.Frame:
		if result.MetaTagFollow {
			result.processLink(result.getAttrVal(node, "src"))
		}
	case atom.Meta:
		name := result.getAttrValLower(node, "name")
		if name == "robots" || name == "googlebot" {
			content := result.getAttrValLower(node, "content")
			if strings.Contains(content, "noindex") {
				result.MetaTagIndex = false
			}
			if strings.Contains(content, "nofollow") {
				result.MetaTagFollow = false
				result.URLs = make(map[string]string)
			}
			if strings.Contains(content, "none") {
				result.MetaTagIndex = false
				result.MetaTagFollow = false
				result.URLs = make(map[string]string)
			}
		}
	}

	result.parseChildren(node)
}

func (result *HTMLParser) parseNode(node *html.Node) error {
	switch node.Type {
	case html.ErrorNode:
		return errors.New("ErrorNode in html")
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

// Parse - start parse
func (result *HTMLParser) Parse(body []byte, baseURL *url.URL) error {
	result.URLs = make(map[string]string)
	result.baseURL = baseURL
	result.MetaTagIndex = true
	result.MetaTagFollow = true
	node, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return err
	}

	return result.parseNode(node)
}

// IsHTML - check is content has html tag
func IsHTML(content []byte) bool {
	isHTML := false
	if len(content) == 0 {
		return isHTML
	}

	z := html.NewTokenizer(bytes.NewReader(content[:1024]))
	isFinish := false
	for !isFinish {
		switch z.Next() {
		case html.ErrorToken:
			isFinish = true
		case html.StartTagToken:
			tagName, _ := z.TagName()
			if bytes.Equal(tagName, []byte("html")) {
				isHTML = true
				isFinish = true
			}
		}
	}

	return isHTML
}
