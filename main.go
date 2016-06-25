package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"

	"golang.org/x/net/html"

	"github.com/ReanGD/go-web-search/content"
	"github.com/ReanGD/go-web-search/crawler"
	"github.com/ReanGD/go-web-search/database"
	"github.com/ReanGD/go-web-search/proxy"
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
}

func (m *Minification) parseChildren(node *html.Node) error {
	for it := node.FirstChild; it != nil; {
		currentNode := it
		it = it.NextSibling
		err := m.Run(currentNode)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Minification) parseElements(node *html.Node) error {
	return m.parseChildren(node)
}

// Run - start minification node
func (m *Minification) Run(node *html.Node) error {
	switch node.Type {
	case html.DocumentNode: // +children -attr (first node)
		return m.parseChildren(node)
	case html.ElementNode: // +children +attr
		return m.parseElements(node)
	case html.TextNode: // -children -attr
		if node.NextSibling != nil && node.NextSibling.Type == html.TextNode {
			fmt.Println(gID, ": double text node! (", node.Data, ")")
		}
		return nil
	case html.DoctypeNode: // ignore
		return nil
	case html.CommentNode: // remove
		return nil
	default:
		return errors.New("minification.miniHTML.parseNode: unexpected node type")
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
	var contents []proxy.Content
	// err = db.Limit(10).Find(&contents).Error
	err = db.Find(&contents).Error
	// err = db.Where("url = 13049", 7).Find(&contents).Error
	if err != nil {
		return err
	}
	CountCompressStatistic := true
	newLength := true
	localMinification := Minification{}

	var oldLen uint64
	var oldLenCompress uint64
	var newLen uint64
	var newLenCompress uint64
	for _, rec := range contents {
		gID = rec.URL
		body := rec.Body.Data
		oldLen += uint64(len(body))
		if CountCompressStatistic {
			bodyCompress := rec.Body.Compress()
			oldLenCompress += uint64(len(bodyCompress))
		}

		node, err := html.Parse(bytes.NewReader(body))
		if err != nil {
			return err
		}

		err = crawler.RunMinificationHTML(node)
		if err != nil {
			return err
		}

		// buf, err := renderNode(node)
		// if err != nil {
		// 	return err
		// }
		// fmt.Println(string(buf))

		err = crawler.RunMinificationText(node)
		if err != nil {
			return err
		}

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
				newField := database.Compressed{Data: buf}
				bodyCompress := newField.Compress()
				newLenCompress += uint64(len(bodyCompress))
			}
		}
	}

	fmt.Println("\noldLen = ", oldLen)
	if CountCompressStatistic {
		fmt.Println("oldLenCompress = ", oldLenCompress)
	}
	if newLength {
		fmt.Println("newLen = ", newLen)
		if CountCompressStatistic {
			fmt.Println("newLenCompress = ", newLenCompress)
		}
	}

	// fmt.Println("\nMax len=", maxLen, " id=", maxLenID)

	return nil
}

func run() error {
	return crawler.Run(baseHosts, 300)
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
