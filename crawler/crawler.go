package crawler

import (
	"fmt"
	"sync"
	"time"
)

func showTotalTime(start time.Time) {
	fmt.Printf("\nTime = %v\n", time.Now().Sub(start))
}

// Run - start download cnt pages
func Run(baseHostsArray []string, cnt int) error {
	if cnt <= 0 || len(baseHostsArray) == 0 {
		return nil
	}

	baseHosts := make(map[string]int)
	for _, host := range baseHostsArray {
		baseHosts[NormalizeHost(host)] = 0
	}
	cntPerHost := cnt / len(baseHosts)
	if cntPerHost < 1 {
		cntPerHost = 1
	}
	for key := range baseHosts {
		baseHosts[key] = cntPerHost
	}

	db := new(DB)
	err := db.Open()
	if err != nil {
		return err
	}

	defer showTotalTime(time.Now())
	defer db.Close()

	robots, err := getHostsRobotsTxt(db, baseHosts)
	if err != nil {
		return err
	}

	chFromDB := make(chan taskFromDB, cnt)
	chToDB := make(chan *taskToDB, 100)
	err = db.readNotLoadedURLs(baseHosts, chFromDB)
	close(chFromDB)
	if err != nil {
		return err
	}

	var wgDB sync.WaitGroup
	wgDB.Add(1)
	go runDbWorker(&wgDB, db, chToDB)

	runWorkers(baseHosts, robots, chFromDB, chToDB)
	close(chToDB)
	wgDB.Wait()
	db.showStatistics()

	return nil
}
