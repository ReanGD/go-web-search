package crawler

// status: ok
import (
	"strings"

	"github.com/ReanGD/go-web-search/werrors"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type dataExtractor struct {
	parserUtils
	meta          *HTMLMetadata
	metaTagFollow bool
	noIndexLvl    int
}

func (extractor *dataExtractor) isEnableLinkParse() bool {
	return extractor.metaTagFollow && extractor.noIndexLvl == 0
}

func (extractor *dataExtractor) parseRef(node *html.Node) {
	if extractor.isEnableLinkParse() {
		rel := extractor.getAttrValLower(node, "rel")
		if rel != "nofollow" {
			extractor.meta.AddURL(extractor.getAttrValLower(node, "href"))
		}
	}
}

func (extractor *dataExtractor) parseLink(node *html.Node) {
	if extractor.isEnableLinkParse() {
		rel := extractor.getAttrValLower(node, "rel")
		if rel == "next" || rel == "prev" || rel == "previous" {
			extractor.meta.AddURL(extractor.getAttrValLower(node, "href"))
		}
	}
}

func (extractor *dataExtractor) parseFrame(node *html.Node) {
	if extractor.isEnableLinkParse() {
		extractor.meta.AddURL(extractor.getAttrValLower(node, "src"))
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
			extractor.meta.ClearURLs()
		}
		if strings.Contains(content, "none") {
			extractor.meta.MetaTagIndex = false
			extractor.metaTagFollow = false
			extractor.meta.ClearURLs()
		}
	} else if name == "title" {
		content := strings.TrimSpace(extractor.getAttrVal(node, "content"))
		extractor.meta.SetTitle(content, false)
	}
}

func (extractor *dataExtractor) parseTitle(node *html.Node) {
	child := node.FirstChild
	if child != nil && child.Type == html.TextNode {
		extractor.meta.SetTitle(strings.TrimSpace(child.Data), true)
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
func RunDataExtrator(hostMng *hostsManager, node *html.Node, urlStr string) (*HTMLMetadata, error) {
	meta, err := NewHTMLMetadata(hostMng, urlStr)
	if err != nil {
		return nil, err
	}

	extractor := dataExtractor{
		meta:          meta,
		metaTagFollow: true,
		noIndexLvl:    0}

	err = extractor.parseNode(node)
	if err != nil {
		return meta, err
	}

	return meta, nil
}
