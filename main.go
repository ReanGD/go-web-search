package main

import (
	"fmt"
	"log"
	"os"

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
	err := crawler.Run("http://habrahabr.ru/", "habrahabr.ru", 1000)
	// err := crawler.ShowDbStatistics()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	f, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	defer f.Close()

	log.SetOutput(f)

	// runtime.GOMAXPROCS(runtime.NumCPU())

	err = run()
	if err != nil {
		fmt.Printf("%s", err)
	}
}
