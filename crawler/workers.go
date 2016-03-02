package crawler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

var chPageWorkerTasks chan string
var wgPageWorkers sync.WaitGroup

func processURL(url string, hostFilter string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}

	if response.StatusCode == 404 {
		savePage404(url)
		return nil
	}

	if response.StatusCode != 200 {
		log.Printf("WARNING: URL: %s returns StatusCode = %d", url, response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return err
	}

	rawURLs, err := parseURLsInPage(bytes.NewReader(body))
	if err != nil {
		return err
	}

	urls, err := processURLs(url, rawURLs, hostFilter)
	if err != nil {
		return err
	}

	return savePage(url, body, urls)
}

func pageWorker(wgParent *sync.WaitGroup, hostFilter string) {
	defer wgParent.Done()

	for i := 0; ; i++ {
		url, more := <-chPageWorkerTasks
		if !more {
			break
		}

		if i%20 == 19 {
			fmt.Print(".")
		}
		err := processURL(url, hostFilter)
		if err != nil {
			log.Printf("ERROR: Process URL: %s message: %s", url, err)
		}
	}
	fmt.Print("!")
}

func startPageWorkersImpl(workersCnt int, hostFilter string) {
	defer wgPageWorkers.Done()

	var wg sync.WaitGroup
	wg.Add(workersCnt)
	for i := 0; i != workersCnt; i++ {
		go pageWorker(&wg, hostFilter)
	}

	wg.Wait()
}

func startPageWorkers(channelSize int, workersCnt int, hostFilter string) error {
	wgPageWorkers.Add(1)
	chPageWorkerTasks = make(chan string, channelSize)
	go startPageWorkersImpl(workersCnt, hostFilter)

	return nil
}

func addTaskToPageWorkers(URL string) {
	chPageWorkerTasks <- URL
}

func finishPageWorkers() {
	close(chPageWorkerTasks)
	wgPageWorkers.Wait()
}
