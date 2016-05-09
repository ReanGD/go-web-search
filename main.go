package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/ReanGD/go-web-search/content"
	"github.com/ReanGD/go-web-search/crawler"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var baseHosts = []string{
	"habrahabr.ru",
	"megamozg.ru",
	"geektimes.ru",
	"linux.org.ru",
	"gamedev.ru",
	"xakep.ru",
	"rsdn.ru",
	"3dnews.ru",
	"ferra.ru",
	"nixp.ru",
	"opennet.ru",
	"computerra.ru"}

var gID int64

func renderNode(node *html.Node) ([]byte, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	err := html.Render(w, node)
	if err != nil {
		return nil, err
	}
	err = w.Flush()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

// Minification -
type Minification struct {
	ShowEnable bool
	Tags       map[string]int
	Attrs      map[string]int
	lvl        int
	RenderTag  atom.Atom
}

func (m *Minification) removeNode(node *html.Node) error {
	node.Parent.RemoveChild(node)

	return nil
}

func (m *Minification) parseChildren(node *html.Node) error {
	m.lvl++
	for it := node.FirstChild; it != nil; {
		currentNode := it
		it = it.NextSibling
		err := m.Run(currentNode)
		if err != nil {
			return err
		}
	}

	m.lvl--
	return nil
}

func (m *Minification) show(node *html.Node) {
	if !m.ShowEnable {
		return
	}
	switch node.Type {
	case html.DocumentNode: // +children -attr (first node)
		fmt.Println(strings.Repeat(" ", m.lvl), "DocumentNode:", node.Data)
	case html.ElementNode: // +children +attr
		fmt.Println(strings.Repeat(" ", m.lvl), "ElementNode:", node.Data)
	case html.TextNode: // -children -attr
		fmt.Println(strings.Repeat(" ", m.lvl), "TextNode:\"", node.Data, "\"")
	case html.DoctypeNode: // ignore
		fmt.Println(strings.Repeat(" ", m.lvl), "DoctypeNode:", node.Data)
	case html.CommentNode: // remove
		fmt.Println(strings.Repeat(" ", m.lvl), "CommentNode:", node.Data)
	default:
		fmt.Println(strings.Repeat(" ", m.lvl), "UnexpectedNode:", node.Data)
	}
}

func (m *Minification) parseElements(node *html.Node) error {
	// if node.DataAtom == atom.Image {
	// ln := len(node.Attr)
	// if ln != 0 {
	// 	attr := node.Attr
	// 	i := 0
	// 	j := 0
	// 	for ; i != ln; i++ {
	// 		switch strings.ToLower(attr[i].Key) {
	// 		case "width":
	// 		case "widht":
	// 		case "height":
	// 		case "alt":
	// 		case "border":
	// 		case "style":
	// 		case "title":
	// 		case "class":
	// 		case "src":
	// 		case "align":
	// 		case "onclick":
	// 		case "sizes":
	// 		case "srcset":
	// 		case "itemprop":
	// 		case "hspace":
	// 		case "vspace":
	// 		case "id":
	// 		case "eight":
	// 		case "imgfield":
	// 		default:
	// 			if i != j {
	// 				attr[j] = attr[i]
	// 			}
	// 			j++
	// 		}
	// 	}
	// 	if i != j {
	// 		node.Attr = attr[:j]
	// 	}
	// }

	// if len(node.Attr) != 0 {
	// 	buf, err := renderNode(node)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fmt.Println(string(buf))
	// }
	// }
	// switch node.DataAtom {
	// case atom.Br:
	// 	node.Type = html.TextNode
	// 	node.Data = " "
	// 	node.FirstChild = nil
	// 	node.LastChild = nil
	// 	node.NextSibling
	// 	// return m.removeNode(node)
	// }

	for _, v := range node.Attr {
		cnt, ok := m.Attrs[v.Key]
		if ok {
			m.Attrs[v.Key] = cnt + 1
		} else {
			m.Attrs[v.Key] = 1
		}
	}

	// if node.Data == "noindex" {
	if node.DataAtom == m.RenderTag {
		buf, err := renderNode(node)
		// buf, err := renderNode(node.Parent)
		if err != nil {
			return err
		}
		fmt.Println(gID, string(buf))
		// fmt.Println(string(buf))
	}

	// for _, v := range node.Attr {
	// 	if v.Key == "data-href" {
	// 		buf, err := renderNode(node)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		fmt.Println(gID, string(buf))
	// 	}
	// }

	return m.parseChildren(node)
}

// Run - start minification node
func (m *Minification) Run(node *html.Node) error {
	m.show(node)
	// if node.Data == "body" {
	// 	m.ShowEnable = true
	// }
	switch node.Type {
	case html.DocumentNode: // +children -attr (first node)
		return m.parseChildren(node)
	case html.ElementNode: // +children +attr
		v, ok := m.Tags[node.Data]
		if ok {
			m.Tags[node.Data] = v + 1
		} else {
			m.Tags[node.Data] = 1
		}
		return m.parseElements(node)
	case html.TextNode: // -children -attr
		return nil
	case html.DoctypeNode: // ignore
		return nil
	case html.CommentNode: // remove
		return nil
	default:
		return errors.New("minification.miniHTML.parseNode: unexpected node type")
	}
}

func sortAndShow(m map[string]int) {
	n := map[int][]string{}
	var a []int
	for k, v := range m {
		n[v] = append(n[v], k)
	}
	for k := range n {
		a = append(a, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(a)))
	for _, k := range a {
		for _, s := range n[k] {
			fmt.Printf("%s, %d\n", s, k)
		}
	}
}

// ClearClose ...
func ClearClose(db *content.DBrw) {
	err := db.Close()
	if err != nil {
		fmt.Printf("Error close db %s", err)
	}
}

func test() error {
	db, err := content.GetDBrw()
	if err != nil {
		return err
	}
	defer ClearClose(db)
	var contents []content.Content
	// err = db.Limit(500).Find(&contents).Error
	err = db.Find(&contents).Error
	// err = db.Where("id = 169", 7).Find(&contents).Error
	if err != nil {
		return err
	}
	CountCompressStatistic := false
	newLength := false
	statisticAttr := false
	statisticTag := true
	localMinification := Minification{
		ShowEnable: false,
		Tags:       make(map[string]int),
		Attrs:      make(map[string]int),
		RenderTag:  atom.B}

	var oldLen uint64
	var oldLenCompress uint64
	var newLen uint64
	var newLenCompress uint64
	for _, rec := range contents {
		gID = rec.ID
		body := rec.Body.Data
		// body = []byte(`<html><head><a href="https://xakep.ru"/></head><body>
		// <b>М</b>
		// <script>S</script>
		// </body></html>`)
		oldLen += uint64(len(body))
		if CountCompressStatistic {
			bodyCompress, err := rec.Body.Compress()
			if err != nil {
				return err
			}
			oldLenCompress += uint64(len(bodyCompress))
		}

		node, err := html.Parse(bytes.NewReader(body))
		if err != nil {
			return err
		}
		err = crawler.RunMinification1(node)
		if err != nil {
			return err
		}

		// buf, err := renderNode(node)
		// if err != nil {
		// 	return err
		// }
		// fmt.Println(string(buf))

		err = localMinification.Run(node)
		if err != nil {
			return err
		}

		if newLength {
			buf, err := renderNode(node)
			if err != nil {
				return err
			}
			newLen += uint64(len(buf))
			if CountCompressStatistic {
				newField := content.Compressed{Data: buf}
				bodyCompress, err := newField.Compress()
				if err != nil {
					return err
				}
				newLenCompress += uint64(len(bodyCompress))
			}
		}

		// fmt.Println(string(buf))
	}

	fmt.Println("oldLen = ", oldLen)
	if CountCompressStatistic {
		fmt.Println("oldLenCompress = ", oldLenCompress)
	}
	if newLength {
		fmt.Println("newLen = ", newLen)
		if CountCompressStatistic {
			fmt.Println("newLenCompress = ", newLenCompress)
		}
	}
	if statisticAttr {
		sortAndShow(localMinification.Attrs)
	}
	if statisticTag {
		sortAndShow(localMinification.Tags)
	}

	return nil
}

func run() error {
	return crawler.Run(baseHosts, 1000)
}

func clearCloseFile(f *os.File) {
	err := f.Close()
	if err != nil {
		fmt.Printf("Error close file: %s", err)
	}
}

func main() {
	f, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	defer clearCloseFile(f)

	log.SetOutput(f)

	// runtime.GOMAXPROCS(runtime.NumCPU())

	err = run()
	// err = test()
	if err != nil {
		fmt.Printf("%s", err)
	}
}
