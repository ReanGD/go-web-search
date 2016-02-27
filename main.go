package main

import (
	"fmt"

	"github.com/ReanGD/go-web-search/parser"
	"github.com/ReanGD/go-web-search/storage"
)

func main() {
	url := "http://habrahabr.ru/"
	p, err := parser.ParseURL(url)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	err = storage.Open()
	if err != nil {
		fmt.Printf("%s", err)
		return
	}
	defer storage.Close()

	words, err := parser.ParseText(p.StringList)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	err = storage.AddWords(words)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	links, err := parser.ParseLinks(url, p.LinkList)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	err = storage.AddLinks(links)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}
}
