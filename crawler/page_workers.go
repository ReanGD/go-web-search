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

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return err
	}

	rawLinks, err := parseURLsInPage(bytes.NewReader(body))
	if err != nil {
		return err
	}

	links, err := processLinks(url, rawLinks, hostFilter)
	if err != nil {
		return err
	}

	return savePage(url, body, links)
}

func pageWorker(wgParent *sync.WaitGroup, hostFilter string) {
	defer wgParent.Done()

	for {
		url, more := <-chPageWorkerTasks
		if !more {
			break
		}

		fmt.Println(url)
		err := processURL(url, hostFilter)
		if err != nil {
			log.Printf("ERROR: Process URL: %s message: %s", url, err)
		}
	}
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
