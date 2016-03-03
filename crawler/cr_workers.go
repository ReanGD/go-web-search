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

type workerInfo struct {
	ChTasks  chan<- string
	TotalURL int
}

type workersManager struct {
	WgWorkers sync.WaitGroup
	BaseHosts map[string]bool
	Workers   map[string]workerInfo
}

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

func (manager *workersManager) runWorker(host string, cnt int) {
	chTasks := make(chan string, cnt)

	worker := &workerParse{
		BaseHosts: manager.BaseHosts,
		WgParent:  &manager.WgWorkers,
		ChTasks:   chTasks}

	manager.Workers[host] = workerInfo{
		ChTasks:  chTasks,
		TotalURL: 0}

	manager.WgWorkers.Add(1)
	go worker.run()
}

func (manager *workersManager) run(baseHosts []string, cnt int) {
	defer manager.WgWorkers.Wait()

	manager.BaseHosts = make(map[string]bool)
	for _, host := range baseHosts {
		manager.BaseHosts[host] = true
	}

	manager.Workers = make(map[string]workerInfo)
	cntPerWorker := cnt / len(baseHosts)
	for _, host := range baseHosts {
		manager.runWorker(host, cntPerWorker)
	}
}

func startWorkersManager(baseHosts []string, cnt int) {
	new(workersManager).run(baseHosts, cnt)
}
