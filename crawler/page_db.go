package crawler

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

const dbName = "pages.db"
const bcContents = "Contents"
const bcLinks = "Links"

var db *bolt.DB

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

// SavePage - save raw content of the page and it links
func SavePage(baseLink string, content []byte, hash [16]byte, links map[string]uint32) error {
	err := db.Update(func(tx *bolt.Tx) error {
		var err error
		bLinks := tx.Bucket([]byte(bcLinks))
		bContents := tx.Bucket([]byte(bcContents))

		for link, count := range links {
			err = insertOrUpdateLinkData(
				bLinks,
				[]byte(link),
				func(data *linkData) error {
					data.Count += count

					return nil
				})
			if err != nil {
				return err
			}
		}

		err = insertOrUpdateLinkData(
			bLinks,
			[]byte(baseLink),
			func(data *linkData) error {
				data.State = 1
				data.Hash = hash
				return nil
			})

		if err != nil {
			return err
		}

		return bContents.Put([]byte(baseLink), content)
	})

	return err
}

func linkDataFromBytes(buf []byte) (*linkData, error) {
	result := new(linkData)
	err := binary.Read(bytes.NewReader(buf), binary.BigEndian, result)
	return result, err
}

// FindNotLoadedLink - find not loaded link
func FindNotLoadedLink() (string, error) {
	var result string
	err := db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bcLinks)).Cursor()
		for link, metaByte := c.First(); link != nil; link, metaByte = c.Next() {
			meta, err := linkDataFromBytes(metaByte)
			if err != nil {
				return err
			}
			if meta.State == 0 {
				result = string(link[:])
				return nil
			}
		}
		return errors.New("Not found free link")
	})

	return result, err
}

// ShowLinksStatistics - show words statistics
func ShowLinksStatistics() error {
	err := db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bcLinks)).Cursor()
		for link, metaByte := c.First(); link != nil; link, metaByte = c.Next() {
			meta, err := linkDataFromBytes(metaByte)
			if err != nil {
				return err
			}
			fmt.Printf("%s: %d (%d)\n", link, meta.Count, meta.State)
		}
		return nil
	})

	return err
}
