package content

import (
	"fmt"
	"log"
	"sync"

	"github.com/jinzhu/gorm"
)

// SaveData - data for save
type SaveData struct {
	HostID   uint64
	MetaItem *Meta
	URLs     map[string]*URL
}

// WriteWorker - worker data
type WriteWorker struct {
	DB          *gorm.DB
	WgParent    *sync.WaitGroup
	TaskChannel <-chan *SaveData
}

func (w *WriteWorker) save(db *gorm.DB, item *SaveData) error {
	err := db.Create(item.MetaItem).Error
	if err != nil {
		return fmt.Errorf("add new 'Meta' record for URL %s, message: %s", item.MetaItem.URL, err)
	}

	var dbItem URL
	urlStr := item.MetaItem.URL
	newItem := &URL{ID: urlStr, HostID: item.HostID, Loaded: true}
	err = db.Where("id = ?", urlStr).First(&dbItem).Error
	if err == gorm.RecordNotFound {
		err = db.Create(newItem).Error
		if err != nil {
			return fmt.Errorf("add new 'URL' record for URL %s, message: %s", urlStr, err)
		}
	} else if err != nil {
		return fmt.Errorf("find in 'URL' table for URL %s, message: %s", urlStr, err)
	} else if dbItem.Loaded == false {
		err = db.Save(newItem).Error
		if err != nil {
			return fmt.Errorf("update 'URL' table with URL %s, message: %s", urlStr, err)
		}
	} else {
		// nothing to update
	}

	for _, newItem := range item.URLs {
		var dbItem URL
		urlStr = newItem.ID
		err = db.Where("id = ?", urlStr).First(&dbItem).Error
		if err == gorm.RecordNotFound {
			err = db.Create(&newItem).Error
			if err != nil {
				return fmt.Errorf("add new 'URL' record for URL %s, message: %s", urlStr, err)
			}
		} else if err != nil {
			return fmt.Errorf("find in 'URL' table for URL %s, message: %s", urlStr, err)
		} else {
			// nothing to update
		}
	}

	return nil
}
func (w *WriteWorker) processChunk() bool {
	finish := false

	tr := w.DB.Begin()
	if tr.Error != nil {
		log.Printf("ERROR: create transaction, message: %s", tr.Error)
		return finish
	}

	var err error
	for i := 0; i != 100; i++ {
		data, more := <-w.TaskChannel
		if !more {
			finish = true
			break
		}
		err = w.save(tr, data)
		if err != nil {
			log.Printf("ERROR: %s", err)
			break
		}
	}

	if err != nil {
		err = tr.Rollback().Error
		if err != nil {
			log.Printf("ERROR: rollback transaction, message: %s", err)
		}
	} else {
		err = tr.Commit().Error
		if err != nil {
			log.Printf("ERROR: commit transaction, message: %s", err)
		}
	}

	return finish
}

// Start - run db write worker
func (w *WriteWorker) Start() {
	defer w.WgParent.Done()

	for !w.processChunk() {
	}
}
