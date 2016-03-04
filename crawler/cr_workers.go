package crawler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type workerParse struct {
	BaseHosts map[string]bool
	WgParent  *sync.WaitGroup
	ChTasks   <-chan string
}

func (worker *workerParse) processURL(url string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == 404 {
		return savePage404(url)
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("StatusCode = %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	rawURLs, err := parseURLsInPage(bytes.NewReader(body))
	if err != nil {
		return err
	}

	urls, err := processURLs(url, rawURLs, worker.BaseHosts)
	if err != nil {
		return err
	}

	return savePage(url, body, urls)
}

func (worker *workerParse) run() {
	defer worker.WgParent.Done()

	for {
		url, more := <-worker.ChTasks
		if !more {
			break
		}
		err := worker.processURL(url)
		if err != nil {
			log.Printf("ERROR: Parse URL \"%s\". Message: %s", url, err)
		} else {
			fmt.Print(".")
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func runWorkers(baseHosts []string, cnt int, chDBTasks <-chan taskFromDB) {
	var wgWorkers sync.WaitGroup
	defer wgWorkers.Wait()

	baseHostsMap := make(map[string]bool)
	for _, host := range baseHosts {
		baseHostsMap[host] = true
	}

	cntPerHost := cnt / len(baseHosts)
	if cntPerHost < 1 {
		cntPerHost = 1
	}

	workers := make(map[string]chan string)
	for _, host := range baseHosts {
		chTasks := make(chan string, cntPerHost)

		worker := &workerParse{
			BaseHosts: baseHostsMap,
			WgParent:  &wgWorkers,
			ChTasks:   chTasks}

		workers[host] = chTasks
		wgWorkers.Add(1)
		go worker.run()
	}

	for {
		task, more := <-chDBTasks
		if !more {
			break
		}
		workers[task.Host] <- task.URL
	}

	for _, ch := range workers {
		close(ch)
	}
}
