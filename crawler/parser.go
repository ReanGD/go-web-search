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

var (
	// ErrParserUnexpectedNodeType - found unexpected node type
	ErrParserUnexpectedNodeType = errors.New("parser.HTMLParser.parseNode: unexpected node type")
)

// HTMLParser - find metatags and links on page
type HTMLParser struct {
	// [URL]hostname
	URLs          map[string]string
	baseURL       *url.URL
	MetaTagIndex  bool
	metaTagFollow bool
	noIndexLvl    int
}

func (result *HTMLParser) getAttrVal(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return strings.ToLower(attr.Val)
		}
	}

	return ""
}

func (result *HTMLParser) isEnableLinkParse() bool {
	return result.metaTagFollow && result.noIndexLvl == 0
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

func (result *HTMLParser) parseChildren(node *html.Node) error {
	for it := node.FirstChild; it != nil; it = it.NextSibling {
		err := result.parseNode(it)
		if err != nil {
			return err
		}
	}

	return nil
}

func (result *HTMLParser) parseElements(node *html.Node) error {
	switch node.DataAtom {
	case atom.A, atom.Area:
		if result.isEnableLinkParse() {
			rel := result.getAttrVal(node, "rel")
			if rel != "nofollow" {
				result.processLink(result.getAttrVal(node, "href"))
			}
		}
	case atom.Link:
		if result.isEnableLinkParse() {
			rel := result.getAttrVal(node, "rel")
			if rel == "next" || rel == "prev" || rel == "previous" {
				result.processLink(result.getAttrVal(node, "href"))
			}
		}
	case atom.Frame, atom.Iframe:
		if result.isEnableLinkParse() {
			result.processLink(result.getAttrVal(node, "src"))
		}
	case atom.Meta:
		name := result.getAttrVal(node, "name")
		if name == "robots" || name == "googlebot" {
			content := result.getAttrVal(node, "content")
			if strings.Contains(content, "noindex") {
				result.MetaTagIndex = false
			}
			if strings.Contains(content, "nofollow") {
				result.metaTagFollow = false
				result.URLs = make(map[string]string)
			}
			if strings.Contains(content, "none") {
				result.MetaTagIndex = false
				result.metaTagFollow = false
				result.URLs = make(map[string]string)
			}
		}
	default:
		if strings.ToLower(node.Data) == "noindex" {
			result.noIndexLvl++
			err := result.parseChildren(node)
			result.noIndexLvl--
			return err
		}
	}

	return result.parseChildren(node)
}

func (result *HTMLParser) parseNode(node *html.Node) error {
	switch node.Type {
	case html.ElementNode:
		return result.parseElements(node)
	case html.DocumentNode, html.CommentNode, html.TextNode, html.DoctypeNode: // skip
		return nil
	default:
		return ErrParserUnexpectedNodeType
	}
}

// Parse - start parse
func (result *HTMLParser) Parse(body []byte, baseURL *url.URL) error {
	result.URLs = make(map[string]string)
	result.baseURL = baseURL
	result.MetaTagIndex = true
	result.metaTagFollow = true
	result.noIndexLvl = 0
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
	if len(content) > 1024 {
		content = content[:1024]
	}

	z := html.NewTokenizer(bytes.NewReader(content))
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
