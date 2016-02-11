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

	// for e := p.Words.Front(); e != nil; e = e.Next() {
	// 	world := e.Value.(string)
	// 	stemmed, err := snowball.Stem(world, "russian", true)
	// 	if err == nil {
	// 		fmt.Println(stemmed)
	// 	} else {
	// 		fmt.Println("error:", world)
	// 	}
	// }

	for e := p.Links.Front(); e != nil; e = e.Next() {
		link := e.Value.(string)
		fmt.Println(link)
	}
}
