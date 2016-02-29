package crawler

import (
	"fmt"
	"time"
)

func showTotalTime(start time.Time) {
	fmt.Printf("Time = %v\n", time.Now().Sub(start))
}

// Run - start download and process pageCnt pages
func Run(defaultURL string, hostFilter string, pageCnt int) error {
	if pageCnt <= 0 {
		return nil
	}

	chWorkersSize := 50
	if chWorkersSize > pageCnt {
		chWorkersSize = pageCnt
	}

	workersCnt := pageCnt / 2
	if workersCnt > 50 {
		workersCnt = 50
	} else if workersCnt < 1 {
		workersCnt = 1
	}

	err := startDbWorker(defaultURL, pageCnt)
	if err != nil {
		return err
	}
	defer showTotalTime(time.Now())
	defer finisDbWorker()

	err = startPageWorkers(chWorkersSize, workersCnt, hostFilter)
	if err != nil {
		return err
	}
	defer finishPageWorkers()

	for i := 0; i < pageCnt; i++ {
		addTaskToPageWorkers(getNextURLForParse())
	}

	return nil
}
