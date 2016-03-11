package crawler

import (
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
		return err
	}

	processedURL := content.URL{ID: meta.URL, HostID: w.Request.Robot.Host.ID, Loaded: true}
	err = db.Save(&processedURL).Error
	if err != nil {
		return err
	}

	for _, newURL := range urls {
		var urlRow content.URL
		err = db.Where("id = ?", newURL.ID).First(&urlRow).Error
		if err == gorm.RecordNotFound {
			err = db.Create(&newURL).Error
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			// nothing to update
		}
	}

	return nil
}

func (w *hostWorker) processURL(urlStr string) {
	meta, newURLs := w.Request.Process(urlStr)
	tr := w.DB.Begin()
	err := tr.Error
	if err == nil {
		err = w.saveResult(tr, meta, newURLs)
		if err != nil {
			tr.Rollback()
		} else {
			err = tr.Commit().Error
			time.Sleep(1000 * time.Millisecond)
		}
	}

	log.Printf("ERROR: DB error, message: %s", err)
}

// Run - start worker
// hostName - normalized host name
func (w *hostWorker) Run(hostName string, cnt int) {
	defer w.WgParent.Done()
	robotTxt := new(robotTxt)

	var host content.Host
	err := w.DB.Where("name = ?", hostName).First(&host).Error
	if err == gorm.RecordNotFound {
		err = robotTxt.FromHostName(hostName)
		if err != nil {
			log.Printf("ERROR: Get load robot.txt for host %s, message: %s", hostName, err)
			return
		}
		err = w.DB.Create(&host).Error
		if err != nil {
			log.Printf("ERROR: Save information for host %s to DB , message: %s", hostName, err)
			return
		}
	} else if err != nil {
		log.Printf("ERROR: Save information for host %s to DB , message: %s", hostName, err)
		return
	} else {
		err = robotTxt.FromHost(&host)
		if err != nil {
			log.Printf("ERROR: Parse robot.txt for host %s from DB, message: %s", hostName, err)
			return
		}
	}

	urlKey := URLFromHost(hostName)
	var urlRow content.URL
	err = w.DB.Where("id = ?", urlKey).First(&urlRow).Error
	if err == gorm.RecordNotFound {
		urlRow = content.URL{ID: urlKey, HostID: host.ID, Loaded: false}
		err = w.DB.Create(&urlRow).Error
		if err != nil {
			log.Printf("ERROR: DB error, message: %s", err)
			return
		}
	} else if err != nil {
		log.Printf("ERROR: DB error, message: %s", err)
		return
	}

	var urls []content.URL
	err = w.DB.Where("host_id = ? and loaded = ?", host.ID, false).Limit(cnt).Find(&urls).Error
	if err != nil {
		log.Printf("ERROR: DB error, message: %s", err)
		return
	}

	w.Request = &request{Robot: robotTxt}
	for _, rowURL := range urls {
		w.processURL(rowURL.ID)
	}
}
