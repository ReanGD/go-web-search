package content

import (
	"fmt"
	"log"
	"sync"

	"github.com/jinzhu/gorm"
)

// TaskData - task data
type TaskData struct {
	URL  string
	Host string
}

// PageData - data for save parsed pages
type PageData struct {
	HostName string
	MetaItem *Meta
	// map[URL]HostName
	URLs map[string]string
}

// WriteWorker - worker data
type WriteWorker struct {
	hostPages map[string]int
	tasksLeft int
	DB        *DBrw
	ChTask    chan<- *TaskData
	WgParent  *sync.WaitGroup
	ChPage    <-chan *PageData
}

func (w *WriteWorker) sendTasks() error {
	for hostName, cnt := range w.hostPages {
		var urls []URL
		err := w.DB.Where("host_id = ? and loaded = ?", w.DB.GetHostID(hostName), false).Limit(cnt).Find(&urls).Error
		if err != nil {
			return fmt.Errorf("find not loaded pages in 'URL' table for host %s, message: %s", hostName, err)
		}
		for _, urlItem := range urls {
			w.ChTask <- &TaskData{URL: urlItem.ID, Host: hostName}
		}
		w.hostPages[hostName] = cnt - len(urls)
		w.tasksLeft -= len(urls)
	}
	if w.tasksLeft <= 0 {
		close(w.ChTask)
	}
	return nil
}

func (w *WriteWorker) savePageData(tr *DBrw, data *PageData) (bool, error) {
	newTasks := false
	err := tr.Create(data.MetaItem).Error
	if err != nil {
		return newTasks, fmt.Errorf("add new 'Meta' record for URL %s, message: %s", data.MetaItem.URL, err)
	}

	var dbItem URL
	urlStr := data.MetaItem.URL
	newItem := &URL{ID: urlStr, HostID: tr.GetHostID(data.HostName), Loaded: true}
	err = tr.Where("id = ?", urlStr).First(&dbItem).Error
	if err == gorm.RecordNotFound {
		err = tr.Create(newItem).Error
		if err != nil {
			return newTasks, fmt.Errorf("add new 'URL' record for URL %s, message: %s", urlStr, err)
		}
	} else if err != nil {
		return newTasks, fmt.Errorf("find in 'URL' table for URL %s, message: %s", urlStr, err)
	} else if dbItem.Loaded == false {
		err = tr.Save(newItem).Error
		if err != nil {
			return newTasks, fmt.Errorf("update 'URL' table with URL %s, message: %s", urlStr, err)
		}
	} else {
		// nothing to update
	}

	for urlStr, hostName := range data.URLs {
		var dbItem URL
		err = tr.Where("id = ?", urlStr).First(&dbItem).Error
		if err == gorm.RecordNotFound {
			newItem := &URL{ID: urlStr, HostID: tr.GetHostID(hostName), Loaded: false}
			err = tr.Create(&newItem).Error
			if err != nil {
				return newTasks, fmt.Errorf("add new 'URL' record for URL %s, message: %s", urlStr, err)
			}
			if w.tasksLeft > 0 {
				cnt, exists := w.hostPages[hostName]
				if exists && cnt > 0 {
					newTasks = true
					w.ChTask <- &TaskData{URL: urlStr, Host: hostName}
					w.hostPages[hostName]--
					w.tasksLeft--
					if w.tasksLeft <= 0 {
						close(w.ChTask)
					}
				}
			}
		} else if err != nil {
			return newTasks, fmt.Errorf("find in 'URL' table for URL %s, message: %s", urlStr, err)
		} else {
			// nothing to update
		}
	}

	return newTasks, nil
}

func (w *WriteWorker) processPages() {
	finish := false
	for !finish {
		newTasks := false
		err := w.DB.Transaction(func(tr *DBrw) error {
			for i := 0; i != 100; i++ {
				data, more := <-w.ChPage
				if !more {
					finish = true
					break
				}

				newTasksPage, err := w.savePageData(tr, data)
				if newTasksPage {
					newTasks = true
				}
				if err != nil {
					return err
				}
			}
			return nil
		})
		if !newTasks && w.tasksLeft > 0 {
			w.tasksLeft = 0
			close(w.ChTask)
		}

		if err != nil {
			log.Printf("ERROR: %s", err)
		}
	}
}

// Start - run db write worker
func (w *WriteWorker) Start(cntPerHost int) {
	defer w.WgParent.Done()

	w.hostPages = make(map[string]int)
	w.tasksLeft = 0
	for hostName := range w.DB.GetHosts() {
		w.hostPages[hostName] = cntPerHost
		w.tasksLeft += cntPerHost
	}

	err := w.sendTasks()
	if err != nil {
		log.Printf("ERROR: %s", err)
	}

	w.processPages()
}
