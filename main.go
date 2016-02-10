package main

import (
	"fmt"

	"github.com/ReanGD/go-web-search/parser"
)

func main() {
	p := parser.NewParser()
	err := p.ParseURL("http://habrahabr.ru/")
	// err := ParseUrl("https://www.linux.org.ru/")
	// err := ParseUrl("http://example.com/")
	// s := `<p>жаба</p>`
	// err := parser.ParseReader(strings.NewReader(s))
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	for e := p.Data.Front(); e != nil; e = e.Next() {
		fmt.Println(e.Value)
	}
	// r, err := zip.OpenReader("testdata/readme.zip")
}
