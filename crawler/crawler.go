package crawler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ReanGD/go-web-search/content"
)

func showTotalTime(msg string, start time.Time) {
	fmt.Printf("\n%s%v\n", msg, time.Now().Sub(start))
}

// Run - start download cnt pages
func Run(baseHosts []string, cnt int) error {
	now := time.Now()
	defer showTotalTime("Total time=", now)
	if cnt <= 0 || len(baseHosts) == 0 {
		return nil
	}

	db, err := content.GetDBrw()
	if err != nil {
		log.Printf("ERROR: %s", err)
		return err
	}
	defer db.Close()

	workers := new(hostWorkers)
	workers.Init(db, baseHosts, cnt)
	if err != nil {
		log.Printf("ERROR: %s", err)
		return err
	}

	var wgDB sync.WaitGroup
	defer wgDB.Wait()
	chDB := make(chan *content.PageData, cnt)
	dbWorker := content.DBWorker{DB: db, ChDB: chDB}
	wgDB.Add(1)
	go dbWorker.Start(&wgDB)

	workers.Start(chDB)
	close(chDB)
	showTotalTime("Workes time=", now)

	return nil
}
