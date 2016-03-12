package crawler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ReanGD/go-web-search/content"
	"github.com/jinzhu/gorm"
)

func showTotalTime(start time.Time) {
	fmt.Printf("\nTime = %v\n", time.Now().Sub(start))
}

func initRequests(db *gorm.DB, baseHosts []string) (map[string]*request, error) {
	requests := make(map[string]*request)

	var hosts []content.Host
	err := db.Find(&hosts).Error
	if err != nil {
		return requests, fmt.Errorf("Get hosts list from db, message: %s", err)
	}
	for i, host := range hosts {
		robotTxt := new(robotTxt)
		err = robotTxt.FromHost(&hosts[i])
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
			err = robotTxt.FromHostName(hostName)
			if err != nil {
				return requests, fmt.Errorf("Load robot.txt for host %s, message: %s", hostName, err)
			}
			err = db.Create(&robotTxt.Host).Error
			if err != nil {
				return requests, fmt.Errorf("Save information for host %s to DB , message: %s", hostName, err)
			}
			requests[hostName] = &request{Robot: robotTxt}
		}
	}

	for hostName, requestItem := range requests {
		urlKey := URLFromHost(hostName)
		var urlItem content.URL
		err = db.Where("id = ?", urlKey).First(&urlItem).Error
		if err == gorm.RecordNotFound {
			urlItem = content.URL{ID: urlKey, HostID: requestItem.Robot.Host.ID, Loaded: false}
			err = db.Create(&urlItem).Error
			if err != nil {
				return requests, fmt.Errorf("Save information for new URL %s to DB , message: %s", urlKey, err)
			}
		} else if err != nil {
			return requests, fmt.Errorf("Get information about URL %s to DB , message: %s", urlKey, err)
		} else {
			// nothing to do
		}
	}

	return requests, nil
}

// Run - start download cnt pages
func Run(baseHosts []string, cnt int) error {
	defer showTotalTime(time.Now())
	if cnt <= 0 || len(baseHosts) == 0 {
		return nil
	}

	db, err := content.GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	requests, err := initRequests(db, baseHosts)
	if err != nil {
		log.Printf("ERROR: %s", err)
		return err
	}

	cntPerHost := cnt / len(requests)
	if cntPerHost < 1 {
		cntPerHost = 1
	}

	var wgWorkers sync.WaitGroup
	for _, requestItem := range requests {
		worker := hostWorker{DB: db, WgParent: &wgWorkers, Request: requestItem}
		wgWorkers.Add(1)
		go worker.Run(cntPerHost)
	}
	wgWorkers.Wait()

	return nil
}
