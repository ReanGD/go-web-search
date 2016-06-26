package main

import (
	"fmt"
	"log"
	"os"

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

// ClearClose ...
func ClearClose(db *content.DBrw) {
	err := db.Close()
	if err != nil {
		fmt.Printf("Error close db %s", err)
	}
}

func test() error {
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
