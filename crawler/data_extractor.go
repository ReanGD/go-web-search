package crawler

import (
	"errors"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	// ErrDataExtractorUnexpectedNodeType - found unexpected node type
	ErrDataExtractorUnexpectedNodeType = errors.New("data_extractor.dataExtractor.parseNode: unexpected node type")
)

// HTMLMetadata extracted meta data from HTML
type HTMLMetadata struct {
	// [URL]hostname
	URLs map[string]string
	// [URL]error
	WrongURLs    map[string]string
	Title        string
	MetaTagIndex bool
}

type dataExtractor struct {
	meta          *HTMLMetadata
	baseURL       *url.URL
	metaTagFollow bool
	noIndexLvl    int
}

func (extractor *dataExtractor) getAttrVal(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return strings.ToLower(attr.Val)
		}
	}

	return ""
}

func (extractor *dataExtractor) isEnableLinkParse() bool {
	return extractor.metaTagFollow && extractor.noIndexLvl == 0
}

func (extractor *dataExtractor) processLink(link string) {
	if link == "" {
		return
	}
	relative, err := url.Parse(strings.TrimSpace(link))
	if err != nil {
		extractor.meta.WrongURLs[link] = err.Error()
		return
	}

	parsed := extractor.baseURL.ResolveReference(relative)
	urlStr := NormalizeURL(parsed)
	parsed, err = url.Parse(urlStr)

	if (parsed.Scheme == "http" || parsed.Scheme == "https") && urlStr != extractor.baseURL.String() {
		extractor.meta.URLs[urlStr] = NormalizeHostName(parsed.Host)
	}
}

func (extractor *dataExtractor) parseChildren(node *html.Node) error {
	for it := node.FirstChild; it != nil; it = it.NextSibling {
		err := extractor.parseNode(it)
		if err != nil {
			return err
		}
	}

	return nil
}

func (extractor *dataExtractor) parseElements(node *html.Node) error {
	switch node.DataAtom {
	case atom.A, atom.Area:
		if extractor.isEnableLinkParse() {
			rel := extractor.getAttrVal(node, "rel")
			if rel != "nofollow" {
				extractor.processLink(extractor.getAttrVal(node, "href"))
			}
		}
	case atom.Link:
		if extractor.isEnableLinkParse() {
			rel := extractor.getAttrVal(node, "rel")
			if rel == "next" || rel == "prev" || rel == "previous" {
				extractor.processLink(extractor.getAttrVal(node, "href"))
			}
		}
	case atom.Frame, atom.Iframe:
		if extractor.isEnableLinkParse() {
			extractor.processLink(extractor.getAttrVal(node, "src"))
		}
	case atom.Meta:
		name := extractor.getAttrVal(node, "name")
		if name == "robots" || name == "googlebot" {
			content := extractor.getAttrVal(node, "content")
			if strings.Contains(content, "noindex") {
				extractor.meta.MetaTagIndex = false
			}
			if strings.Contains(content, "nofollow") {
				extractor.metaTagFollow = false
				extractor.meta.URLs = make(map[string]string)
			}
			if strings.Contains(content, "none") {
				extractor.meta.MetaTagIndex = false
				extractor.metaTagFollow = false
				extractor.meta.URLs = make(map[string]string)
			}
		} else if name == "title" {
			content := strings.TrimSpace(extractor.getAttrVal(node, "content"))
			if content != "" && extractor.meta.Title == "" {
				extractor.meta.Title = content
			}
		}
	case atom.Title:
		child := node.FirstChild
		if child != nil && child.Type == html.TextNode {
			title := strings.TrimSpace(child.Data)
			if title != "" {
				extractor.meta.Title = title
			}
		}
	default:
		if strings.ToLower(node.Data) == "noindex" {
			extractor.noIndexLvl++
			err := extractor.parseChildren(node)
			extractor.noIndexLvl--
			return err
		}
	}

	return extractor.parseChildren(node)
}

func (extractor *dataExtractor) parseNode(node *html.Node) error {
	if !extractor.meta.MetaTagIndex {
		return nil
	}
	switch node.Type {
	case html.DocumentNode:
		return extractor.parseChildren(node)
	case html.ElementNode:
		return extractor.parseElements(node)
	case html.CommentNode, html.TextNode, html.DoctypeNode: // skip
		return nil
	default:
		return ErrDataExtractorUnexpectedNodeType
	}
}

// RunDataExtrator - extart URLs and other meta data from page
func RunDataExtrator(node *html.Node, baseURL *url.URL) (*HTMLMetadata, error) {
	extractor := dataExtractor{
		meta: &HTMLMetadata{
			URLs:         make(map[string]string),
			WrongURLs:    make(map[string]string),
			Title:        "",
			MetaTagIndex: true,
		},
		baseURL:       baseURL,
		metaTagFollow: true,
		noIndexLvl:    0,
	}

	err := extractor.parseNode(node)
	if err != nil {
		return &HTMLMetadata{}, err
	}

	meta := extractor.meta
	if meta.Title == "" {
		meta.Title = baseURL.String()
	}

	runeTitle := []rune(meta.Title)
	if len(runeTitle) > 100 {
		meta.Title = string(runeTitle[0:97]) + "..."
	}

	return meta, nil
}
