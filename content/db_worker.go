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
	MetaItem *Meta
	// map[URL]HostName
	URLs map[string]string
}

// DBWorker - worker for save data to db
type DBWorker struct {
	DB   *DBrw
	ChDB <-chan *PageData
}

func (w *DBWorker) markURLLoaded(tr *DBrw, urlStr string, hostID sql.NullInt64) (sql.NullInt64, error) {
	var urlRec URL
	parent := sql.NullInt64{Valid: false}
	err := tr.Where("id = ?", urlStr).First(&urlRec).Error
	if err == gorm.RecordNotFound {
		newItem := URL{
			ID:     urlStr,
			Parent: parent,
			HostID: hostID,
			Loaded: true}
		err = tr.Create(&newItem).Error
		if err != nil {
			return parent, fmt.Errorf("add new 'URL' record for URL %s, message: %s", urlStr, err)
		}
	} else if err != nil {
		return parent, fmt.Errorf("find in 'URL' table for URL %s, message: %s", urlStr, err)
	} else if urlRec.Loaded == false {
		err = tr.Model(&urlRec).Update("Loaded", true).Error
		if err != nil {
			return parent, fmt.Errorf("update 'URL' table with URL %s, message: %s", urlStr, err)
		}
	} else {
		// nothing to update
	}

	return urlRec.Parent, nil
}

func (w *DBWorker) saveMeta(tr *DBrw, meta *Meta, origin sql.NullInt64) error {
	hostID := tr.GetHostID(meta.HostName)
	urlStr := meta.URL
	var metaRec Meta
	err := tr.Where("url = ?", urlStr).First(&metaRec).Error
	if err == gorm.RecordNotFound {
		meta.Parent, err = w.markURLLoaded(tr, urlStr, hostID)
		if err != nil {
			return err
		}
		if !hostID.Valid {
			if origin.Valid {
				meta.State = StateDublicate
			} else {
				meta.State = StateExternal
			}
			meta.Content = Content{}
		}
		if origin.Valid {
			meta.Origin = origin
		} else {
			meta.Origin, err = tr.GetOrigin(meta)
			if err != nil {
				return err
			}
		}
		if meta.Origin.Valid {
			meta.State = StateDublicate
			meta.Content = Content{}
		}
		err = tr.Create(meta).Error
		if err != nil {
			return fmt.Errorf("add new 'Meta' record for URL %s, message: %s", urlStr, err)
		}
		err = tr.AddHash(meta)
		if err != nil {
			return err
		}
		if meta.RedirectReferer != nil {
			err = w.saveMeta(tr, meta.RedirectReferer, sql.NullInt64{Int64: meta.ID, Valid: true})
			if err != nil {
				return err
			}
		}
	} else if err != nil {
		return fmt.Errorf("find in 'Meta' table for URL %s, message: %s", urlStr, err)
	} else {
		_, err = w.markURLLoaded(tr, urlStr, hostID)
		if meta.RedirectReferer != nil {
			err = w.saveMeta(tr, meta.RedirectReferer, sql.NullInt64{Int64: metaRec.ID, Valid: true})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *DBWorker) savePageData(tr *DBrw, data *PageData) error {
	err := w.saveMeta(tr, data.MetaItem, sql.NullInt64{Valid: false})
	if err != nil {
		return nil
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
