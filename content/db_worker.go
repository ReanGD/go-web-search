package content

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/jinzhu/gorm"
)

// PageData - data for save parsed pages
type PageData struct {
	HostName string
	MetaItem *Meta
	// map[URL]HostName
	URLs map[string]string
}

// DBWorker - worker for save data to db
type DBWorker struct {
	DB   *DBrw
	ChDB <-chan *PageData
}

func (w *DBWorker) savePageData(tr *DBrw, data *PageData) error {
	var dbItem URL
	urlStr := data.MetaItem.URL
	newItem := &URL{ID: urlStr, HostID: tr.GetHostID(data.HostName), Loaded: true}
	err := tr.Where("id = ?", urlStr).First(&dbItem).Error
	if err == gorm.RecordNotFound {
		newItem.Parent = sql.NullInt64{Valid: false}
		err = tr.Create(newItem).Error
		if err != nil {
			return fmt.Errorf("add new 'URL' record for URL %s, message: %s", urlStr, err)
		}
	} else if err != nil {
		return fmt.Errorf("find in 'URL' table for URL %s, message: %s", urlStr, err)
	} else if dbItem.Loaded == false {
		newItem.Parent = dbItem.Parent
		err = tr.Save(newItem).Error
		if err != nil {
			return fmt.Errorf("update 'URL' table with URL %s, message: %s", urlStr, err)
		}
	} else {
		// nothing to update
	}

	data.MetaItem.Parent = newItem.Parent
	err = tr.Create(data.MetaItem).Error
	if err != nil {
		return fmt.Errorf("add new 'Meta' record for URL %s, message: %s", data.MetaItem.URL, err)
	}

	parent := sql.NullInt64{Int64: data.MetaItem.ID, Valid: true}
	for urlStr, hostName := range data.URLs {
		var dbItem URL
		err = tr.Where("id = ?", urlStr).First(&dbItem).Error
		if err == gorm.RecordNotFound {
			newItem := &URL{ID: urlStr, Parent: parent, HostID: tr.GetHostID(hostName), Loaded: false}
			err = tr.Create(&newItem).Error
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

// Start - run db write worker
func (w *DBWorker) Start(wgParent *sync.WaitGroup) {
	defer wgParent.Done()

	finish := false
	for !finish {
		err := w.DB.Transaction(func(tr *DBrw) error {
			for i := 0; i != 100; i++ {
				data, more := <-w.ChDB
				if !more {
					finish = true
					break
				}

				err := w.savePageData(tr, data)
				if err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("ERROR: %s", err)
		}
	}
}
