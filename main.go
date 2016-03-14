package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ReanGD/go-web-search/crawler"
	_ "github.com/mattn/go-sqlite3"
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
	"computerra.ru"}

func test() error {
	return nil
}

func run() error {
	return crawler.Run(baseHosts, 1000)
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
	// err = test()
	if err != nil {
		fmt.Printf("%s", err)
	}
}
