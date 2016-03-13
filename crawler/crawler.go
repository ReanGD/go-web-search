package crawler

import (
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/ReanGD/go-web-search/content"
)

func showTotalTime(msg string, start time.Time) {
	fmt.Printf("\n%s%v\n", msg, time.Now().Sub(start))
}

func loadHosts(db *content.DBrw, baseHosts []string) (map[string]*request, error) {
	var err error
	requests := make(map[string]*request)

	for _, host := range db.GetHosts() {
		robotTxt := new(robotTxt)
		err = robotTxt.FromHost(host)
		if err != nil {
			return requests, fmt.Errorf("Init robot.txt for host %s from db data, message: %s", host.Name, err)
		}
		requests[host.Name] = &request{Robot: robotTxt}
	}

	for _, hostNameRaw := range baseHosts {
		hostName := NormalizeHost(hostNameRaw)
		_, exists := requests[hostName]
		if !exists {
			robotTxt := new(robotTxt)
			host, err := robotTxt.FromHostName(hostName)
			if err != nil {
				return requests, fmt.Errorf("Load robot.txt for host %s, message: %s", hostName, err)
			}
			err = db.AddHost(host, NormalizeURL(&url.URL{Scheme: "http", Host: hostName}))
			if err != nil {
				return requests, err
			}
			requests[hostName] = &request{Robot: robotTxt}
		}
	}

	return requests, nil
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

	requests, err := loadHosts(db, baseHosts)
	if err != nil {
		log.Printf("ERROR: %s", err)
		return err
	}

	cntPerHost := cnt / len(requests)
	if cntPerHost < 1 {
		cntPerHost = 1
	}

	var wgDB sync.WaitGroup
	chTask := make(chan *content.TaskData, cnt)
	chPage := make(chan *content.PageData, cnt)
	dbWorker := &content.WriteWorker{
		DB:       db,
		ChTask:   chTask,
		WgParent: &wgDB,
		ChPage:   chPage}

	wgDB.Add(1)
	go dbWorker.Start(cntPerHost)

	startWorkers(chTask, chPage, requests, cntPerHost)
	close(chPage)
	showTotalTime("Workes time=", now)

	wgDB.Wait()

	return nil
}
