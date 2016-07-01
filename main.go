package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ReanGD/go-web-search/content"
	"github.com/ReanGD/go-web-search/crawler"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/uber-go/zap"
)

var baseHosts = []string{
	"habrahabr.ru",
	"geektimes.ru",
	"linux.org.ru",
	"gamedev.ru",
	"xakep.ru",
	"rsdn.ru",
	"ixbt.com",
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

func run(logger zap.Logger) error {
	return crawler.Run(logger, baseHosts, 4300)
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

	file, err := os.OpenFile("app.json", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	logger := zap.NewJSON(zap.DebugLevel, zap.Output(zap.AddSync(file)))

	log.SetOutput(f)

	// runtime.GOMAXPROCS(runtime.NumCPU())
	err = run(logger)
	// err = test()
	if err != nil {
		fmt.Printf("%s", err)
	}
}
