package crawler

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

const dbName = "pages.db"
const bcContents = "Contents"
const bcLinks = "Links"

var db *bolt.DB
var chURLForParse chan string
var chPageForSave chan *pageInfoForSave
var wgDbWorker sync.WaitGroup

// OpenDb - open or create database
func OpenDb() error {
	var err error
	db, err = bolt.Open(dbName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bcContents))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(bcLinks))
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

// CloseDb - close database
func CloseDb() {
	db.Close()
}

type linkData struct {
	ID    uint64
	Count uint32
	Hash  [16]byte
	State uint8
}

type linkDataUpdate func(data *linkData) error

func insertOrUpdateLinkData(bucket *bolt.Bucket, key []byte, fun linkDataUpdate) error {
	var data *linkData
	byteLink := bucket.Get(key)
	if byteLink == nil {
		id, _ := bucket.NextSequence()
		data = &linkData{ID: id, Count: 0, State: 0}
	} else {
		data = new(linkData)
		err := binary.Read(bytes.NewReader(byteLink), binary.BigEndian, data)
		if err != nil {
			return err
		}
	}

	err := fun(data)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, *data)
	if err != nil {
		return err
	}
	return bucket.Put(key, buf.Bytes())
}

type pageInfoForSave struct {
	BaseLink string
	Content  []byte
	Hash     [16]byte
	Links    map[string]uint32
}

func savePageImpl(data *pageInfoForSave, bLinks *bolt.Bucket, bContents *bolt.Bucket, cntURLs int) (int, error) {
	var err error

	for link, count := range data.Links {
		err = insertOrUpdateLinkData(
			bLinks,
			[]byte(link),
			func(obj *linkData) error {
				obj.Count += count

				if obj.State == 0 && cntURLs > 0 && data.BaseLink != link {
					chURLForParse <- link
					cntURLs--
				}

				return nil
			})
		if err != nil {
			return cntURLs, err
		}
	}

	err = insertOrUpdateLinkData(
		bLinks,
		[]byte(data.BaseLink),
		func(obj *linkData) error {
			obj.State = 1
			obj.Hash = data.Hash
			return nil
		})

	if err != nil {
		return cntURLs, err
	}

	err = bContents.Put([]byte(data.BaseLink), data.Content)

	return cntURLs, err
}

func savePage(baseLink string, content []byte, links map[string]uint32) error {
	var zContent bytes.Buffer
	w := zlib.NewWriter(&zContent)
	_, err := w.Write(content)
	w.Close()

	if err == nil {
		hash := md5.Sum(content)
		chPageForSave <- &pageInfoForSave{BaseLink: baseLink, Content: zContent.Bytes(), Hash: hash, Links: links}
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
			var err error
			bLinks := tx.Bucket([]byte(bcLinks))
			bContents := tx.Bucket([]byte(bcContents))

			for i := 0; i != 100; i++ {
				data, more := <-chPageForSave
				if !more {
					finish = true
					return nil
				}
				cntURLs, err = savePageImpl(data, bLinks, bContents, cntURLs)
				if err != nil {
					log.Printf("ERROR: Save parsed URL (%s) to db, message: %s", data.BaseLink, err)
					return err
				}
			}

			return nil
		})
		if err != nil {
			log.Printf("ERROR: Save transaction to db, message: %s", err)
		}
	}
}

func linkDataFromBytes(buf []byte) (*linkData, error) {
	result := new(linkData)
	err := binary.Read(bytes.NewReader(buf), binary.BigEndian, result)
	return result, err
}

func notLoadedDbURLsToChannel(defaultURL string, cntURLs int) (int, error) {
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bcLinks))
		c := b.Cursor()
		isFound := false
		for link, metaByte := c.First(); link != nil && cntURLs != 0; link, metaByte = c.Next() {
			meta, err := linkDataFromBytes(metaByte)
			if err != nil {
				return err
			}
			if meta.State == 0 {
				chURLForParse <- string(link[:])
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

// ShowDbStatistics -
func ShowDbStatistics() error {
	err := OpenDb()
	if err != nil {
		return err
	}

	defer CloseDb()
	return db.View(func(tx *bolt.Tx) error {
		bContents := tx.Bucket([]byte(bcContents))
		fmt.Printf("Contents len = %d\n", bContents.Stats().KeyN)

		bLinks := tx.Bucket([]byte(bcLinks))
		fmt.Printf("Links len = %d\n", bLinks.Stats().KeyN)

		return nil
	})

	// err := db.View(func(tx *bolt.Tx) error {
	// 	c := tx.Bucket([]byte(bcLinks)).Cursor()
	// 	for link, metaByte := c.First(); link != nil; link, metaByte = c.Next() {
	// 		meta, err := linkDataFromBytes(metaByte)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		fmt.Printf("%s: %d (%d)\n", link, meta.Count, meta.State)
	// 	}
	// 	return nil
	// })

	// return err
}
