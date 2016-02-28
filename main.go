package main

import (
	"fmt"

	"github.com/ReanGD/go-web-search/crawler"
)

// func run() error {
// 	url := "http://habrahabr.ru/"
// 	p, err := parser.ParseURL(url)
// 	if err != nil {
// 		return err
// 	}

// 	err = storage.Open()
// 	if err != nil {
// 		return err
// 	}
// 	defer storage.Close()

// 	words, err := parser.ParseText(p.StringList)
// 	if err != nil {
// 		return err
// 	}

// 	err = storage.AddWords(words)
// 	if err != nil {
// 		return err
// 	}

// 	// err = storage.ShowWordStatistics()
// 	if err != nil {
// 		return err
// 	}

// 	links, err := parser.ParseLinks(url, p.LinkList)
// 	if err != nil {
// 		return err
// 	}

// 	err = storage.AddLinks(links)
// 	if err != nil {
// 		return err
// 	}

// 	err = storage.ShowLinksStatistics()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func run() error {
	err := crawler.Process("http://habrahabr.ru/", "habrahabr.ru", 10)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		fmt.Printf("%s", err)
	}
}
