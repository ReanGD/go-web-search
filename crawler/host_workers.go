package crawler

import (
	"fmt"
	"sync"
	"time"

	"github.com/ReanGD/go-web-search/content"
)

type hostWorker struct {
	Request *request
	Tasks   []string
	ChDB    chan<- *content.PageData
}

// Run - start worker
func (w *hostWorker) Start(wgParent *sync.WaitGroup) {
	defer wgParent.Done()

	cnt := len(w.Tasks)
	for i := 0; i != cnt; i++ {
		result := w.Request.Process(w.Tasks[i])
		w.ChDB <- result
		fmt.Printf(".")
		state := result.MetaItem.State
		if state != content.StateErrorURLFormat && state != content.StateDisabledByRobotsTxt && i != cnt-1 {
			time.Sleep(1 * time.Second)
		}
	}
}

type hostWorkers struct {
	workers map[string]*hostWorker
}

func (w *hostWorkers) Init(db *content.DBrw, baseHosts []string, cnt int) error {
	var err error
	w.workers = make(map[string]*hostWorker)

	for _, host := range db.GetHosts() {
		robotTxt := new(robotTxt)
		err = robotTxt.FromHost(host)
		if err != nil {
			return fmt.Errorf("Init robot.txt for host %s from db data, message: %s", host.Name, err)
		}
		w.workers[host.Name] = &hostWorker{Request: &request{Robot: robotTxt}}
	}

	for _, hostNameRaw := range baseHosts {
		hostName := NormalizeHostName(hostNameRaw)
		_, exists := w.workers[hostName]
		if !exists {
			robotTxt := new(robotTxt)
			host, err := robotTxt.FromHostName(hostName)
			if err != nil {
				return fmt.Errorf("Load robot.txt for host %s, message: %s", hostName, err)
			}
			baseURL, err := GenerateURLByHostName(hostName)
			if err != nil {
				return err
			}
			err = db.AddHost(host, baseURL)
			if err != nil {
				return err
			}
			w.workers[host.Name] = &hostWorker{Request: &request{Robot: robotTxt}}
		}
	}

	cntPerHost := cnt / len(w.workers)
	if cntPerHost < 1 {
		cntPerHost = 1
	}

	for hostName, worker := range w.workers {
		worker.Request.Init()
		worker.Tasks, err = db.GetNewURLs(hostName, cntPerHost)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *hostWorkers) Start(chDB chan<- *content.PageData) {
	var wg sync.WaitGroup
	defer wg.Wait()

	for _, worker := range w.workers {
		worker.ChDB = chDB
		wg.Add(1)
		go worker.Start(&wg)
	}
}
