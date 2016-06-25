package content

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/ReanGD/go-web-search/proxy"
	"github.com/jinzhu/gorm"
)

// DBWorker - worker for save data to db
type DBWorker struct {
	DB   *DBrw
	ChDB <-chan *proxy.PageData
}

func (w *DBWorker) getURLIDByStr(tr *DBrw, urlStr string) (sql.NullInt64, error) {
	var urlRec URL
	err := tr.Where("url = ?", urlStr).First(&urlRec).Error
	if err == gorm.ErrRecordNotFound {
		return sql.NullInt64{Valid: false}, nil
	} else if err != nil {
		return sql.NullInt64{Valid: false}, fmt.Errorf("find in 'URL' table for URL %s, message: %s", urlStr, err)
	} else {
		return sql.NullInt64{Int64: urlRec.ID, Valid: true}, nil
	}
}

func (w *DBWorker) markURLLoaded(tr *DBrw, id sql.NullInt64, urlStr string, hostID sql.NullInt64) (int64, error) {
	var errID int64
	var urlRec URL
	if !id.Valid {
		urlRec = URL{
			URL:    urlStr,
			HostID: hostID,
			Loaded: true}
		err := tr.Create(&urlRec).Error
		if err != nil {
			return errID, fmt.Errorf("add new 'URL' record for URL %s, message: %s", urlStr, err)
		}
		return urlRec.ID, nil
	}
	err := tr.Model(&urlRec).Where("id = ?", id.Int64).Update("Loaded", true).Error
	if err != nil {
		return errID, fmt.Errorf("update 'URL' table with URL %s, message: %s", urlStr, err)
	}
	return id.Int64, nil
}

func (w *DBWorker) insertURLIfNotExists(tr *DBrw, urlStr string, hostName string) (int64, error) {
	var rec URL
	err := tr.Where("url = ?", urlStr).First(&rec).Error
	if err == gorm.ErrRecordNotFound {
		rec = URL{URL: urlStr, HostID: tr.GetHostID(hostName), Loaded: false}
		err = tr.Create(&rec).Error
		if err != nil {
			return rec.ID, fmt.Errorf("add new 'URL' record for URL %s, message: %s", urlStr, err)
		}
	} else if err != nil {
		return rec.ID, fmt.Errorf("find in 'URL' table for URL %s, message: %s", urlStr, err)
	}

	return rec.ID, nil
}

func (w *DBWorker) insertLinkIfNotExists(tr *DBrw, master int64, slave int64) error {
	var rec Link
	err := tr.Where("master = ? and slave = ?", master, slave).First(&rec).Error
	if err == gorm.ErrRecordNotFound {
		rec = Link{Master: master, Slave: slave}
		err = tr.Create(&rec).Error
		if err != nil {
			return fmt.Errorf("add new 'Link' record for master %d and slave %d, message: %s",
				uint64(master), uint64(slave), err)
		}
	} else if err != nil {
		return fmt.Errorf("find in 'Link' table for master %d and slave %d, message: %s",
			uint64(master), uint64(slave), err)
	}

	return nil
}

func (w *DBWorker) saveMeta(tr *DBrw, meta *proxy.InMeta, origin sql.NullInt64) error {
	hostID := tr.GetHostID(meta.GetHostName())
	urlStr := meta.GetURL()
	urlNullID, err := w.getURLIDByStr(tr, urlStr)
	if err != nil {
		return err
	}
	urlID, err := w.markURLLoaded(tr, urlNullID, urlStr, hostID)
	if err != nil {
		return err
	}
	if origin.Valid {
		err = w.insertLinkIfNotExists(tr, urlID, origin.Int64)
		if err != nil {
			return err
		}
	}

	var metaRec proxy.Meta
	err = tr.Where("url = ?", urlID).First(&metaRec).Error
	if err == gorm.ErrRecordNotFound {
		if !hostID.Valid {
			if origin.Valid {
				meta.SetState(proxy.StateDublicate)
			} else {
				meta.SetState(proxy.StateExternal)
			}
		}

		if !origin.Valid {
			hash := meta.GetHash()
			if len(hash) != 0 {
				origin = tr.FindOrigin(hash)
			}
		}
		meta.SetOrigin(origin)

		if meta.GetState() == proxy.StateSuccess {
			content := meta.GetContent()
			if content == nil {
				return fmt.Errorf("field 'content' is nil")
			}
			err = tr.Create(content.GetContent(urlID)).Error
			if err != nil {
				return fmt.Errorf("add new 'Content' record for URL %s, message: %s", urlStr, err)
			}
		}

		err = tr.Create(meta.GetMeta(urlID)).Error
		if err != nil {
			return fmt.Errorf("add new 'Meta' record for URL %s, message: %s", urlStr, err)
		}
		hash := meta.GetHash()
		if len(hash) != 0 {
			tr.AddHash(hash, urlID)
		}

		if meta.GetReferer() != nil {
			err = w.saveMeta(tr, meta.GetReferer(), sql.NullInt64{Int64: urlID, Valid: true})
			if err != nil {
				return err
			}
		}
	} else if err != nil {
		return fmt.Errorf("find in 'Meta' table for URL %s, message: %s", urlStr, err)
	} else {
		if meta.GetReferer() != nil {
			err = w.saveMeta(tr, meta.GetReferer(), sql.NullInt64{Int64: urlID, Valid: true})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *DBWorker) savePageData(tr *DBrw, data *proxy.PageData) error {
	err := w.saveMeta(tr, data.GetMeta(), sql.NullInt64{Valid: false})
	if err != nil {
		return err
	}

	var id int64
	for urlStr, hostName := range data.GetURLs() {
		id, err = w.insertURLIfNotExists(tr, urlStr, hostName)
		if err != nil {
			return err
		}
		err = w.insertLinkIfNotExists(tr, data.GetParentURL(), id)
		if err != nil {
			return err
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
