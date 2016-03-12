package crawler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ReanGD/go-web-search/content"
	"github.com/jinzhu/gorm"
)

type hostWorker struct {
	DB          *gorm.DB
	WgParent    *sync.WaitGroup
	Request     *request
	TaskChannel chan<- *content.SaveData
}

// Run - start worker
func (w *hostWorker) Run(cnt int) {
	defer w.WgParent.Done()

	var urls []content.URL
	tr := w.DB.Begin()
	err := tr.Where("host_id = ? and loaded = ?", w.Request.Robot.Host.ID, false).Limit(cnt).Find(&urls).Error
	tr.Commit()
	if err != nil {
		log.Printf("ERROR: DB error, message: %s", err)
		return
	}

	hostID := w.Request.Robot.Host.ID
	for _, rowURL := range urls {
		meta, newURLs := w.Request.Process(rowURL.ID)
		data := &content.SaveData{HostID: hostID, MetaItem: meta, URLs: newURLs}
		w.TaskChannel <- data
		fmt.Printf(".")
		if meta.State != content.StateErrorURLFormat && meta.State != content.StateDisabledByRobotsTxt {
			time.Sleep(1 * time.Second)
		}
	}
}
