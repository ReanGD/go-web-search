package crawler

import (
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/ReanGD/go-web-search/content"
	"github.com/ReanGD/go-web-search/proxy"
	"github.com/uber-go/zap"
)

type hostWorker struct {
	Request *request
	Tasks   []content.URL
	ChDB    chan<- *proxy.PageData
}

// Run - start worker
func (w *hostWorker) Start(wgParent *sync.WaitGroup) {
	defer wgParent.Done()

	cnt := len(w.Tasks)
	for i := 0; i != cnt; i++ {
		parsed, err := url.Parse(w.Tasks[i].URL)
		if err != nil {
			log.Printf("ERROR: Worker query. Parse URL %s, message: %s", w.Tasks[i].URL, err)
			continue
		}
		result, workDuration := w.Request.Process(parsed)
		result.SetParentURL(w.Tasks[i].ID)
		w.ChDB <- result
		fmt.Printf(".")
		if result.GetMeta().NeedWaitAfterRequest() && i != cnt-1 {
			time.Sleep(time.Duration(1000-workDuration) * time.Millisecond)
		}
	}
}

type hostWorkers struct {
	workers []*hostWorker
}

func (w *hostWorkers) Init(db *content.DBrw, logger zap.Logger, baseHosts []string, cnt int) error {
	hostMng := &hostsManager{}
	err := hostMng.Init(db, baseHosts)
	if err != nil {
		return err
	}

	hosts := hostMng.GetHosts()
	w.workers = make([]*hostWorker, len(hosts))
	cntPerHost := cnt / len(w.workers)
	if cntPerHost < 1 {
		cntPerHost = 1
	}
	for i, hostName := range hostMng.GetHosts() {
		worker := &hostWorker{Request: &request{hostMng: hostMng}}
		worker.Request.Init(logger.With(zap.String("host", hostName)))
		worker.Tasks, err = db.GetNewURLs(hostName, cntPerHost)
		if err != nil {
			return err
		}
		w.workers[i] = worker
	}

	return nil
}

func (w *hostWorkers) Start(chDB chan<- *proxy.PageData) {
	var wg sync.WaitGroup
	defer wg.Wait()

	for _, worker := range w.workers {
		worker.ChDB = chDB
		wg.Add(1)
		go worker.Start(&wg)
	}
}
