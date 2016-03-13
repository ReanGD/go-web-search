package crawler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ReanGD/go-web-search/content"
)

type hostWorker struct {
	WgParent *sync.WaitGroup
	Request  *request
	ChTask   chan string
	ChDB     chan<- *content.PageData
}

// Run - start worker
func (w *hostWorker) Run() {
	defer w.WgParent.Done()

	for {
		urlStr, more := <-w.ChTask
		if !more {
			break
		}
		data := w.Request.Process(urlStr)
		fmt.Printf(".")
		w.ChDB <- data
		state := data.MetaItem.State
		if state != content.StateErrorURLFormat && state != content.StateDisabledByRobotsTxt {
			time.Sleep(1 * time.Second)
		}
	}
}

func startWorkers(chTask <-chan *content.TaskData,
	chDB chan<- *content.PageData,
	requests map[string]*request,
	cntPerHost int) {

	workers := make(map[string]*hostWorker, len(requests))

	var wgWorkers sync.WaitGroup
	defer wgWorkers.Wait()
	for hostName, requestItem := range requests {
		workers[hostName] = &hostWorker{
			WgParent: &wgWorkers,
			Request:  requestItem,
			ChTask:   make(chan string, cntPerHost),
			ChDB:     chDB}

		wgWorkers.Add(1)
		go workers[hostName].Run()
	}

	for {
		task, more := <-chTask
		if !more {
			break
		}
		worker, exists := workers[task.Host]
		if !exists {
			log.Printf("WARNING: not found host worker %s", task.Host)
		} else {
			worker.ChTask <- task.URL
		}
	}

	for _, worker := range workers {
		close(worker.ChTask)
	}
}
