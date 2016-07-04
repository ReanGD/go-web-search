package crawler

// status: ok
import (
	"net/url"
	"strings"

	"github.com/ReanGD/go-web-search/werrors"
	"github.com/uber-go/zap"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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
	parserUtils
	meta          *HTMLMetadata
	baseURL       *url.URL
	metaTagFollow bool
	noIndexLvl    int
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
	parsed, _ = url.Parse(urlStr)

	if (parsed.Scheme == "http" || parsed.Scheme == "https") && urlStr != extractor.baseURL.String() {
		extractor.meta.URLs[urlStr] = NormalizeHostName(parsed.Host)
	}
}

func (extractor *dataExtractor) parseRef(node *html.Node) {
	if extractor.isEnableLinkParse() {
		rel := extractor.getAttrValLower(node, "rel")
		if rel != "nofollow" {
			extractor.processLink(extractor.getAttrValLower(node, "href"))
		}
	}
}

func (extractor *dataExtractor) parseLink(node *html.Node) {
	if extractor.isEnableLinkParse() {
		rel := extractor.getAttrValLower(node, "rel")
		if rel == "next" || rel == "prev" || rel == "previous" {
			extractor.processLink(extractor.getAttrValLower(node, "href"))
		}
	}
}

func (extractor *dataExtractor) parseFrame(node *html.Node) {
	if extractor.isEnableLinkParse() {
		extractor.processLink(extractor.getAttrValLower(node, "src"))
	}
}

func (extractor *dataExtractor) parseMeta(node *html.Node) {
	name := extractor.getAttrValLower(node, "name")
	if name == "robots" || name == "googlebot" {
		content := extractor.getAttrValLower(node, "content")
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
}

func (extractor *dataExtractor) parseTitle(node *html.Node) {
	child := node.FirstChild
	if child != nil && child.Type == html.TextNode {
		title := strings.TrimSpace(child.Data)
		if title != "" {
			extractor.meta.Title = title
		}
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
		extractor.parseRef(node)
	case atom.Link:
		extractor.parseLink(node)
	case atom.Frame, atom.Iframe:
		extractor.parseFrame(node)
	case atom.Meta:
		extractor.parseMeta(node)
	case atom.Title:
		extractor.parseTitle(node)
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
		return werrors.New(ErrUnexpectedNodeType)
	}
}

// RunDataExtrator - extart URLs and other meta data from page
func RunDataExtrator(node *html.Node, urlStr string) (*HTMLMetadata, error) {
	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, werrors.NewFields(ErrParseBaseURL,
			zap.String("details", err.Error()),
			zap.String("parsed_url", urlStr))
	}

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

	err = extractor.parseNode(node)
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
