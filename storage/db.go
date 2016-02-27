package storage

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/boltdb/bolt"
)

const dbName string = "data.db"

var db *bolt.DB

// Open - open or create database
func Open() error {
	var err error
	db, err = bolt.Open(dbName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Word"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("Link"))
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

// Close - close database
func Close() {
	db.Close()
}

type wordData struct {
	ID    uint64
	Count uint32
}

func wordDataFromBytes(buf []byte) (*wordData, error) {
	result := new(wordData)
	err := binary.Read(bytes.NewReader(buf), binary.BigEndian, result)
	return result, err
}

func (word *wordData) Bytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, *word)
	return buf.Bytes(), err
}

// AddWords - add the found words
func AddWords(words map[string]uint32) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Word"))
		var dbWord *wordData
		var err error
		for word, count := range words {
			byteWord := b.Get([]byte(word))
			if byteWord == nil {
				id, _ := b.NextSequence()
				dbWord = &wordData{ID: id, Count: 0}
			} else {
				dbWord, err = wordDataFromBytes(byteWord)
				if err != nil {
					return err
				}
			}
			dbWord.Count += count
			byteWord, err = dbWord.Bytes()
			if err != nil {
				return err
			}
			err = b.Put([]byte(word), byteWord)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

type linkData struct {
	ID    uint64
	Count uint32
}

func linkDataFromBytes(buf []byte) (*linkData, error) {
	result := new(linkData)
	err := binary.Read(bytes.NewReader(buf), binary.BigEndian, result)
	return result, err
}

func (link *linkData) Bytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, *link)
	return buf.Bytes(), err
}

// AddLinks - add the found links
func AddLinks(links map[string]uint32) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Link"))
		var dbLink *linkData
		var err error
		for link, count := range links {
			byteLink := b.Get([]byte(link))
			if byteLink == nil {
				id, _ := b.NextSequence()
				dbLink = &linkData{ID: id, Count: 0}
			} else {
				dbLink, err = linkDataFromBytes(byteLink)
				if err != nil {
					return err
				}
			}
			dbLink.Count += count
			byteLink, err = dbLink.Bytes()
			if err != nil {
				return err
			}
			err = b.Put([]byte(link), byteLink)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

// func readDb() error {
// 	err := db.View(func(tx *bolt.Tx) error {
// 		b := tx.Bucket([]byte("Words"))
// 		c := b.Cursor()
// 		for k, v := c.First(); k != nil; k, v = c.Next() {
// 			fmt.Printf("key=%s, value=%s\n", k, v)
// 		}
// 		return nil
// 	})

// 	return err
// }
