package crawler

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/temoto/robotstxt.go"
)

type workerParse struct {
	WgParent    *sync.WaitGroup
	ChTasks     <-chan string
	ChTasksToDB chan<- *taskToDB
	HandlerURLs *handlerURLs
}

func (worker *workerParse) processURL(url string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == 404 {
		worker.ChTasksToDB <- &taskToDB{URL: url, ErrorType: PageType404}
		return nil
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

	urls, err := worker.HandlerURLs.handle(url, rawURLs)
	if err != nil {
		return err
	}

	var zContent bytes.Buffer
	w := zlib.NewWriter(&zContent)
	_, err = w.Write(body)
	w.Close()

	if err == nil {
		worker.ChTasksToDB <- &taskToDB{
			URL:        url,
			ErrorType:  PageTypeSuccess,
			Content:    zContent.Bytes(),
			Hash:       md5.Sum(body),
			URLsOnPage: urls}
	}

	return err
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

func runWorkers(baseHosts map[string]int,
	robots map[string]*robotstxt.Group,
	chTasksFromDB <-chan taskFromDB,
	chTasksToDB chan<- *taskToDB) {

	var wgWorkers sync.WaitGroup
	defer wgWorkers.Wait()

	workers := make(map[string]chan string)
	for host, cnt := range baseHosts {
		chTasks := make(chan string, cnt)

		worker := &workerParse{
			WgParent:    &wgWorkers,
			ChTasks:     chTasks,
			ChTasksToDB: chTasksToDB,
			HandlerURLs: &handlerURLs{Robot: robots[host], BaseHosts: baseHosts}}

		workers[host] = chTasks
		wgWorkers.Add(1)
		go worker.run()
	}

	for {
		task, more := <-chTasksFromDB
		if !more {
			break
		}
		workers[task.Host] <- task.URL
	}

	for _, ch := range workers {
		close(ch)
	}
}
