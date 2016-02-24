package main

import (
	"fmt"

	"github.com/ReanGD/go-web-search/parser"
)

func main() {
	// p, err := parser.ParseURL("https://www.linux.org.ru/")
	// p, err := parser.ParseURL("http://example.com/")
	// s := `<p>жаба</p>`
	// p, err := parser.ParseStream(strings.NewReader(s))

	// url := "http://habrahabr.ru/"
	url := "http://habrahabr.ru/"
	p, err := parser.ParseURL(url)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	words, err := parser.ParseText(p.StringList)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	links, err := parser.ParseLinks(url, p.LinkList)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	for word := range words {
		if word == "" {
			fmt.Println(word)
		}
	}

	for link := range links {
		if link != "" {
			fmt.Println(link)
		}
	}
}
