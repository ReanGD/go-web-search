package crawler

import (
	"github.com/ReanGD/go-web-search/werrors"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type minificationHTML struct {
	parserUtils
}

func (m *minificationHTML) toDiv(node *html.Node) (*html.Node, error) {
	node.DataAtom = atom.Div
	node.Data = "div"
	node.Attr = nil

	return m.parseChildren(node)
}

func (m *minificationHTML) parseChildren(node *html.Node) (*html.Node, error) {
	var err error
	for it := node.FirstChild; it != nil; {
		it, err = m.parseNode(it)
		if err != nil {
			return node.NextSibling, err
		}
	}

	return node.NextSibling, nil
}

func (m *minificationHTML) parseElements(node *html.Node) (*html.Node, error) {
	switch node.DataAtom {
	case atom.A:
		return m.openNode(node, false)
	case atom.Abbr:
		title := m.getAttrValLower(node, "title")
		if title != "" {
			m.addChildTextNodeToBegining(node, " "+title+" ")
		}
		return m.openNode(node, true)
	case atom.Address:
		return m.openNode(node, true)
	case atom.Applet:
		return m.removeNode(node, true)
	case atom.Area:
		return m.removeNode(node, true)
	case atom.Article:
		return m.toDiv(node)
	case atom.Aside:
		return m.toDiv(node)
	case atom.Audio:
		return m.removeNode(node, true)
	case atom.B:
		return m.openNode(node, false)
	case atom.Base:
		return m.removeNode(node, false)
	case atom.Basefont:
		return m.removeNode(node, false)
	case atom.Bdi:
		return m.openNode(node, false)
	case atom.Bdo:
		return m.openNode(node, false)
	case atom.Bgsound:
		return m.removeNode(node, false)
	case atom.Blockquote:
		return m.toDiv(node)
	case atom.Big:
		return m.openNode(node, false)
	case atom.Body:
		node.Attr = nil
		return m.parseChildren(node)
	case atom.Blink:
		return m.openNode(node, false)
	case atom.Br:
		return m.removeNode(node, true)
	case atom.Button:
		return m.removeNode(node, true)
	case atom.Canvas:
		return m.removeNode(node, true)
	case atom.Caption:
		return m.openNode(node, true)
	case atom.Center:
		return m.toDiv(node)
	case atom.Cite:
		return m.openNode(node, false)
	case atom.Code:
		return m.openNode(node, false)
	case atom.Colgroup, atom.Col:
		return m.removeNode(node, true)
	case atom.Command:
		return m.removeNode(node, true)
	case atom.Data:
		return m.removeNode(node, false)
	case atom.Datalist:
		return m.removeNode(node, false)
	case atom.Dd:
		return m.openNode(node, true)
	case atom.Del:
		return m.openNode(node, false)
	case atom.Details:
		return m.toDiv(node)
	case atom.Dfn:
		return m.openNode(node, false)
	case atom.Dialog:
		return m.toDiv(node)
	case atom.Div:
		return m.toDiv(node)
	case atom.Dl:
		return m.toDiv(node)
	case atom.Dt:
		return m.openNode(node, true)
	case atom.Em:
		return m.openNode(node, false)
	case atom.Embed:
		return m.removeNode(node, true)
	case atom.Figcaption:
		return m.openNode(node, true)
	case atom.Figure:
		return m.toDiv(node)
	case atom.Font:
		return m.openNode(node, false)
	case atom.Footer:
		return m.toDiv(node)
	case atom.Form:
		return m.removeNode(node, true)
	case atom.Frame, atom.Frameset, atom.Noframes:
		return m.removeNode(node, false)
	case atom.H1:
		return m.toDiv(node)
	case atom.H2:
		return m.toDiv(node)
	case atom.H3:
		return m.toDiv(node)
	case atom.H4:
		return m.toDiv(node)
	case atom.H5:
		return m.toDiv(node)
	case atom.H6:
		return m.toDiv(node)
	case atom.Head:
		node.Attr = nil
		return m.parseChildren(node)
	case atom.Header:
		return m.toDiv(node)
	case atom.Hr:
		return m.removeNode(node, true)
	case atom.Html:
		node.Attr = nil
		return m.parseChildren(node)
	case atom.I:
		return m.openNode(node, false)
	case atom.Iframe:
		return m.removeNode(node, true)
	case atom.Img:
		return m.removeNode(node, true)
	case atom.Input:
		return m.removeNode(node, true)
	case atom.Ins:
		return m.openNode(node, false)
	case atom.Label:
		return m.toDiv(node)
	case atom.Li:
		return m.openNode(node, true)
	case atom.Link:
		return m.removeNode(node, false)
	case atom.Listing:
		return m.toDiv(node)
	case atom.Marquee:
		return m.openNode(node, true)
	case atom.Meta:
		return m.removeNode(node, false)
	case atom.Name:
		return m.openNode(node, false)
	case atom.Nav:
		return m.toDiv(node)
	case atom.Nobr:
		return m.openNode(node, false)
	case atom.Noscript:
		return m.removeNode(node, true)
	case atom.Object:
		return m.removeNode(node, true)
	case atom.Ol:
		return m.toDiv(node)
	case atom.P:
		return m.toDiv(node)
	case atom.Param:
		return m.removeNode(node, true)
	case atom.Pre:
		return m.toDiv(node)
	case atom.Q:
		return m.openNode(node, false)
	case atom.S:
		return m.openNode(node, false)
	case atom.Script:
		return m.removeNode(node, true)
	case atom.Section:
		return m.toDiv(node)
	case atom.Select:
		return m.removeNode(node, true)
	case atom.Small:
		return m.openNode(node, false)
	case atom.Span:
		return m.openNode(node, false)
	case atom.Strike:
		return m.openNode(node, false)
	case atom.Strong:
		return m.openNode(node, false)
	case atom.Style:
		return m.removeNode(node, true)
	case atom.Sub:
		return m.openNode(node, false)
	case atom.Sup:
		return m.openNode(node, false)
	case atom.Svg:
		return m.removeNode(node, true)
	case atom.Table:
		return m.toDiv(node)
	case atom.Tbody:
		return m.openNode(node, true)
	case atom.Td:
		return m.toDiv(node)
	case atom.Textarea:
		return m.removeNode(node, true)
	case atom.Tfoot:
		return m.openNode(node, true)
	case atom.Th:
		return m.toDiv(node)
	case atom.Thead:
		return m.openNode(node, true)
	case atom.Time:
		return m.openNode(node, false)
	case atom.Title:
		return m.removeNode(node, false)
	case atom.Tr:
		return m.openNode(node, true)
	case atom.Tt:
		return m.openNode(node, false)
	case atom.U:
		return m.openNode(node, false)
	case atom.Ul:
		return m.toDiv(node)
	case atom.Var:
		return m.openNode(node, false)
	case atom.Video:
		return m.removeNode(node, true)
	case atom.Wbr:
		return m.removeNode(node, false)
	default:
		if node.Data == "noindex" {
			return m.removeNode(node, true)
		}
		return m.toDiv(node)
	}
}

func (m *minificationHTML) parseNode(node *html.Node) (*html.Node, error) {
	switch node.Type {
	case html.DocumentNode: // +children -attr (first node)
		return m.parseChildren(node)
	case html.ElementNode: // +children +attr
		return m.parseElements(node)
	case html.TextNode: // -children -attr
		return node.NextSibling, nil
	case html.DoctypeNode: // ignore
		return m.removeNode(node, false)
	case html.CommentNode: // remove
		return m.removeNode(node, false)
	default:
		return nil, werrors.New(ErrUnexpectedNodeType)
	}
}

func (m *minificationHTML) Run(node *html.Node) error {
	_, err := m.parseNode(node)
	return err
}
