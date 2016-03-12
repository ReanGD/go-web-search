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
	DB       *gorm.DB
	WgParent *sync.WaitGroup
	Request  *request
}

func (w *hostWorker) saveResult(db *gorm.DB, meta *content.Meta, urls map[string]*content.URL) error {
	err := db.Create(meta).Error
	if err != nil {
		return fmt.Errorf("add new meta record for URL %s, message: %s", meta.URL, err)
	}

	processedURL := content.URL{ID: meta.URL, HostID: w.Request.Robot.Host.ID, Loaded: true}
	err = db.Save(&processedURL).Error
	if err != nil {
		return fmt.Errorf("update Loaded state, message: %s", err)
	}

	for _, newURL := range urls {
		var urlRow content.URL
		err = db.Where("id = ?", newURL.ID).First(&urlRow).Error
		if err == gorm.RecordNotFound {
			err = db.Create(&newURL).Error
			if err != nil {
				return fmt.Errorf("add new URL record, message: %s", err)
			}
		} else if err != nil {
			return fmt.Errorf("find in URL table, message: %s", err)
		} else {
			// nothing to update
		}
	}

	return nil
}

func (w *hostWorker) processURL(urlStr string) {
	log.Printf("Start process URL %s", urlStr)
	meta, newURLs := w.Request.Process(urlStr)
	tr := w.DB.Begin()
	err := tr.Error
	if err == nil {
		err = w.saveResult(tr, meta, newURLs)
		if err != nil {
			log.Printf("ERROR: Save error, %s", err)
			err = tr.Rollback().Error
		} else {
			err = tr.Commit().Error
			time.Sleep(1000 * time.Millisecond)
		}
		if err != nil {
			log.Printf("ERROR: transaction, message: %s", err)
		}
	}
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

	for _, rowURL := range urls {
		w.processURL(rowURL.ID)
	}
}
