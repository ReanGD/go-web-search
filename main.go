package main

import (
	"fmt"

	"github.com/ReanGD/go-web-search/parser"
)

func main() {
	p, err := parser.ParseURL("http://habrahabr.ru/")
	// p, err := parser.ParseURL("https://www.linux.org.ru/")
	// p, err := parser.ParseURL("http://example.com/")
	// s := `<p>жаба</p>`
	// p, err := parser.ParseStream(strings.NewReader(s))
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	for e := p.Data.Front(); e != nil; e = e.Next() {
		fmt.Println(e.Value)
	}
	// r, err := zip.OpenReader("testdata/readme.zip")
}
