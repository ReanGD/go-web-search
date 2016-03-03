package crawler

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

var db *bolt.DB
var chURLForParse chan string
var chPageForSave chan *pageInfoForSave
var wgDbWorker sync.WaitGroup

// OpenDb - open or create database
func OpenDb() error {
	var err error
	db, err = bolt.Open(DbName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		var err error
		_, err = tx.CreateBucketIfNotExists([]byte(DbBucketContents))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(DbBucketWrongURLs))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(DbBucketURLs))
		if err != nil {
			return err
		}
		bMeta, err := tx.CreateBucketIfNotExists([]byte(DbBucketMeta))
		if err != nil {
			return err
		}

		metaVal := DbMeta{LastID: 0}
		bytes, err := metaVal.MarshalMsg(nil)
		if err != nil {
			return err
		}

		err = bMeta.Put([]byte(DbKeyMeta), bytes)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		db.Close()
	}

	return err
}

// CloseDb - close database
func CloseDb() {
	db.Close()
}

type updateDbURL func(data *DbURL) error

func insertOrUpdateDbURL(bucket *bolt.Bucket, key []byte, fun updateDbURL) error {
	var data DbURL
	var err error
	bytes := bucket.Get(key)
	if bytes != nil {
		data.ErrorType = PageTypeSuccess
		_, err = data.UnmarshalMsg(bytes)
		if err != nil {
			return err
		}
	}

	err = fun(&data)
	if err != nil {
		return err
	}

	bytes, err = data.MarshalMsg(nil)
	if err != nil {
		return err
	}
	return bucket.Put(key, bytes)
}

type pageInfoForSave struct {
	URL string
	// see PageTypeSuccess, PageType404, etc.
	ErrorType  uint8
	Content    []byte
	Hash       [16]byte
	URLsOnPage map[string]bool
}

func savePageImpl(
	bContents *bolt.Bucket,
	bWrongURLs *bolt.Bucket,
	bURLs *bolt.Bucket,
	data *pageInfoForSave,
	id uint64,
	cntURLs int) (int, error) {

	if data.ErrorType != PageTypeSuccess {
		dataWrongURL := DbWrongURL{ErrorType: data.ErrorType}
		bytes, err := dataWrongURL.MarshalMsg(nil)
		if err != nil {
			return cntURLs, err
		}
		err = bWrongURLs.Put([]byte(data.URL), bytes)
		if err != nil {
			return cntURLs, err
		}

		err = insertOrUpdateDbURL(bURLs, []byte(data.URL),
			func(obj *DbURL) error {
				obj.ErrorType = data.ErrorType
				return nil
			})

		return cntURLs, err
	}

	dataContent := DbContent{ID: id, Content: data.Content, Hash: data.Hash}
	bytes, err := dataContent.MarshalMsg(nil)
	if err != nil {
		return cntURLs, err
	}
	err = bContents.Put([]byte(data.URL), bytes)
	if err != nil {
		return cntURLs, err
	}

	err = insertOrUpdateDbURL(bURLs, []byte(data.URL),
		func(obj *DbURL) error {
			obj.ID = id
			obj.Count = 0
			return nil
		})

	if err != nil {
		return cntURLs, err
	}

	for url := range data.URLsOnPage {
		err = insertOrUpdateDbURL(bURLs, []byte(url),
			func(obj *DbURL) error {
				if obj.ID == 0 && obj.ErrorType == PageTypeSuccess && cntURLs > 0 {
					chURLForParse <- url
					cntURLs--
				}
				obj.Count++
				return nil
			})

		if err != nil {
			return cntURLs, err
		}
	}

	return cntURLs, err
}

func savePage404(url string) error {
	chPageForSave <- &pageInfoForSave{URL: url, ErrorType: PageType404}

	return nil
}

func savePage(url string, content []byte, urlsOnPage map[string]bool) error {
	var zContent bytes.Buffer
	w := zlib.NewWriter(&zContent)
	_, err := w.Write(content)
	w.Close()

	if err == nil {
		hash := md5.Sum(content)
		chPageForSave <- &pageInfoForSave{URL: url, ErrorType: PageTypeSuccess, Content: zContent.Bytes(), Hash: hash, URLsOnPage: urlsOnPage}
	}

	return err
}

func startDbWorkerImpl(cntURLs int) {
	defer wgDbWorker.Done()
	defer CloseDb()
	defer close(chURLForParse)

	finish := false
	for !finish {
		err := db.Update(func(tx *bolt.Tx) error {
			bContents := tx.Bucket([]byte(DbBucketContents))
			bWrongURLs := tx.Bucket([]byte(DbBucketWrongURLs))
			bURLs := tx.Bucket([]byte(DbBucketURLs))
			bMeta := tx.Bucket([]byte(DbBucketMeta))

			var metaVal DbMeta
			metaBytes := bMeta.Get([]byte(DbKeyMeta))
			if metaBytes == nil {
				return errors.New("Can not load meta data for db page")
			}

			_, err := metaVal.UnmarshalMsg(metaBytes)
			if err != nil {
				log.Printf("ERROR: Parse meta data value, message: %s", err)
				return err
			}

			lastID := metaVal.LastID

			for i := 0; i != 100; i++ {
				data, more := <-chPageForSave
				if !more {
					finish = true
					return nil
				}

				cntURLs, err = savePageImpl(bContents, bWrongURLs, bURLs, data, lastID+1, cntURLs)
				if err != nil {
					log.Printf("ERROR: Save parsed URL (%s) to db, message: %s", data.URL, err)
					return err
				}

				lastID++
			}

			metaVal.LastID = lastID
			metaBytes, err = metaVal.MarshalMsg(nil)
			if err != nil {
				log.Printf("ERROR: Serrialize meta data value, message: %s", err)
				return err
			}

			err = bMeta.Put([]byte(DbKeyMeta), metaBytes)
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

func notLoadedDbURLsToChannel(defaultURL string, cntURLs int) (int, error) {
	err := db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(DbBucketURLs)).Cursor()
		isFound := false

		var urlVal DbURL
		for url, urlBytes := c.First(); url != nil && cntURLs != 0; url, urlBytes = c.Next() {
			_, err := urlVal.UnmarshalMsg(urlBytes)
			if err != nil {
				return err
			}

			if urlVal.ID == 0 {
				chURLForParse <- string(url[:])
				cntURLs--
				isFound = true
			}
		}

		if !isFound {
			chURLForParse <- string(defaultURL)
			cntURLs--
		}
		return nil
	})

	return cntURLs, err
}

func startDbWorker(defaultURL string, cntURLs int) error {
	err := OpenDb()
	if err != nil {
		return err
	}

	chURLForParse = make(chan string, cntURLs)
	cntURLs, err = notLoadedDbURLsToChannel(defaultURL, cntURLs)
	if err != nil {
		close(chURLForParse)
		CloseDb()
		return err
	}

	chPageForSave = make(chan *pageInfoForSave, 50)
	wgDbWorker.Add(1)
	go startDbWorkerImpl(cntURLs)

	return nil
}

func getNextURLForParse() string {
	return <-chURLForParse
}

func finisDbWorker() {
	close(chPageForSave)
	wgDbWorker.Wait()
}

// ShowDbStatistics - show db statistics
func ShowDbStatistics() error {
	err := OpenDb()
	if err != nil {
		return err
	}

	defer CloseDb()
	return db.View(func(tx *bolt.Tx) error {
		bContents := tx.Bucket([]byte(DbBucketContents))
		fmt.Printf("Contents len = %d\n", bContents.Stats().KeyN)

		bURLs := tx.Bucket([]byte(DbBucketURLs))
		fmt.Printf("bURLs len = %d\n", bURLs.Stats().KeyN)

		bWrongURLs := tx.Bucket([]byte(DbBucketWrongURLs))
		fmt.Printf("WrongURLs len = %d\n", bWrongURLs.Stats().KeyN)
		return nil
	})
}

// CheckDb - check db
func CheckDb() error {
	err := OpenDb()
	if err != nil {
		return err
	}

	defer CloseDb()

	isSuccess := true
	minLen := math.MaxUint32
	maxLen := 0
	hashMap := make(map[[16]byte]string)
	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(DbBucketContents)).Cursor()

		var content DbContent
		for url, contentBytes := c.First(); url != nil; url, contentBytes = c.Next() {
			_, err := content.UnmarshalMsg(contentBytes)
			if err != nil {
				fmt.Printf("Error unmarshal value in db for url %s, message: %s\n", url, err)
				isSuccess = false
				continue
			}

			r, err := zlib.NewReader(bytes.NewReader(content.Content))
			if err != nil {
				fmt.Printf("Error unzip content for url %s, message: %s\n", url, err)
				isSuccess = false
				continue
			}
			contentOrig, err := ioutil.ReadAll(r)
			r.Close()
			if err != nil {
				fmt.Printf("Error read unzip content for url %s, message: %s\n", url, err)
				isSuccess = false
				continue
			}

			lenContent := len(contentOrig)
			if lenContent < minLen {
				minLen = lenContent
			}
			if lenContent > maxLen {
				maxLen = lenContent
			}

			hash := md5.Sum(contentOrig)
			if hash != content.Hash {
				fmt.Printf("Error content hash does not match for url %s\n", url)
				isSuccess = false
				continue
			}

			if val, ok := hashMap[hash]; ok {
				fmt.Printf("Duplicated pages content:\n%s\n%s\n\n", url, val)
			} else {
				hashMap[hash] = string(url[:])
			}
		}

		return nil
	})

	if isSuccess {
		fmt.Printf("Min len = %d\n", minLen)
		fmt.Printf("Max len = %d\n", maxLen)
		fmt.Println("Checking ended successfully")
	}
	return err
}
