package crawler

import (
	"errors"
	"log"
	"sync"
)

type taskToDB struct {
	URL string
	// see PageTypeSuccess, PageType404, etc.
	ErrorType  uint8
	Content    []byte
	Hash       [16]byte
	URLsOnPage map[string]bool
}

func saveWrongPage(bWrongURLs *Bucket, bURLs *Bucket, data *taskToDB) error {
	key := []byte(data.URL)
	dataWrongURL := DbWrongURL{ErrorType: data.ErrorType}
	err := bWrongURLs.Put(key, &dataWrongURL)
	if err != nil {
		return err
	}

	var obj DbURL
	_, err = bURLs.Get(key, &obj)
	if err != nil {
		return err
	}
	obj.ErrorType = data.ErrorType
	err = bURLs.Put(key, obj)
	return err
}

func savePage(bContents *Bucket, bURLs *Bucket, data *taskToDB, id uint64) error {
	key := []byte(data.URL)
	dataContent := DbContent{ID: id, Content: data.Content, Hash: data.Hash}
	err := bContents.Put(key, &dataContent)
	if err != nil {
		return err
	}

	var obj DbURL
	_, err = bURLs.Get(key, &obj)
	if err != nil {
		return err
	}
	obj.ID = id
	obj.ErrorType = PageTypeSuccess
	err = bURLs.Put(key, obj)
	if err != nil {
		return err
	}

	for url := range data.URLsOnPage {
		key = []byte(url)
		var obj DbURL
		exists, err := bURLs.Get(key, &obj)
		if err != nil {
			return err
		}
		if !exists {
			obj.ErrorType = PageTypeNone
		}
		obj.Count++
		err = bURLs.Put(key, obj)
		if err != nil {
			return err
		}
	}

	return err
}

func runDbWorker(wgParent *sync.WaitGroup, db *DB, ch <-chan *taskToDB) {
	defer wgParent.Done()

	finish := false
	for !finish {
		err := db.Update(func(tx *Tx) error {
			bContents := tx.Bucket(DbBucketContents)
			bWrongURLs := tx.Bucket(DbBucketWrongURLs)
			bURLs := tx.Bucket(DbBucketURLs)
			bMeta := tx.Bucket(DbBucketMeta)

			var metaVal DbMeta
			exists, err := bMeta.Get([]byte(DbKeyMeta), &metaVal)
			if err != nil {
				log.Printf("ERROR: Parse meta data value, message: %s", err)
				return err
			}
			if !exists {
				return errors.New("Can not load meta data for db page")
			}
			lastID := metaVal.LastID

			for i := 0; i != 100; i++ {
				data, more := <-ch
				if !more {
					finish = true
					return nil
				}

				if data.ErrorType == PageTypeSuccess {
					err = savePage(bContents, bURLs, data, lastID+1)
				} else {
					err = saveWrongPage(bWrongURLs, bURLs, data)
				}
				if err != nil {
					log.Printf("ERROR: Save parsed URL (%s) to db, message: %s", data.URL, err)
					return err
				}

				if data.ErrorType == PageTypeSuccess {
					lastID++
				}
			}

			metaVal.LastID = lastID
			err = bMeta.Put([]byte(DbKeyMeta), metaVal)
			if err != nil {
				log.Printf("ERROR: Save meta data value to db, message: %s", err)
				return err
			}

			return nil
		})

		if err != nil {
			log.Printf("ERROR: Save transaction to db, message: %s", err)
		}
	}
}
