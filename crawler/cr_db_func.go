package crawler

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io/ioutil"
	"net/url"
)

type taskFromDB struct {
	Host string
	URL  string
}

func (db *DB) readNotLoadedURLs(baseHosts map[string]int, ch chan<- taskFromDB) error {
	leftCnts := make(map[string]int, len(baseHosts))
	for k, v := range baseHosts {
		leftCnts[k] = v
	}

	return db.View(func(tx *Tx) error {
		b := tx.Bucket(DbBucketURLs)
		c := b.Cursor()

		var urlVal DbURL
		for urlKey, urlData := c.First(); urlKey != nil; urlKey, urlData = c.Next() {
			parsedURL, err := url.Parse(string(urlKey[:]))
			if err != nil {
				return err
			}

			leftCnt, found := leftCnts[parsedURL.Host]
			if !found || leftCnt <= 0 {
				continue
			}

			_, err = urlVal.UnmarshalMsg(urlData)
			if err != nil {
				return err
			}

			if urlVal.ErrorType == PageTypeNone {
				ch <- taskFromDB{Host: parsedURL.Host, URL: string(urlKey[:])}
				leftCnts[parsedURL.Host]--
			}

			totalCnt := 0
			for _, i := range leftCnts {
				totalCnt += i
			}
			if totalCnt <= 0 {
				break
			}
		}

		for host, cntLeft := range leftCnts {
			if cntLeft > 0 {
				u := URLFromHost(host)
				exists, err := b.Get([]byte(u), &urlVal)
				if err != nil {
					return err
				}
				if !exists {
					ch <- taskFromDB{Host: host, URL: u}
				}
			}
		}

		return nil
	})
}

func (db *DB) showStatistics() error {
	return db.View(func(tx *Tx) error {
		bContents := tx.Bucket(DbBucketContents)
		fmt.Printf("\nStatistics:\n")
		fmt.Printf("Contents: %d\n", bContents.Stats().KeyN)

		bWrongURLs := tx.Bucket(DbBucketWrongURLs)
		fmt.Printf("WrongURLs: %d\n", bWrongURLs.Stats().KeyN)

		bURLs := tx.Bucket(DbBucketURLs)
		cURLs := bURLs.Cursor()
		fmt.Printf("URLs: %d:\n", bURLs.Stats().KeyN)

		hostsStat := make(map[string]int)
		for urlKey, _ := cURLs.First(); urlKey != nil; urlKey, _ = cURLs.Next() {
			parsedURL, err := url.Parse(string(urlKey[:]))
			if err != nil {
				return err
			}
			hostsStat[parsedURL.Host]++
		}
		for host, cnt := range hostsStat {
			fmt.Printf("  %s = %d\n", host, cnt)
		}

		return nil
	})
}

// GetContent - Get content by URL
func (db *DB) GetContent(u string) ([]byte, error) {
	var result []byte

	normURL, err := NormalizeRawURL(u)
	if err != nil {
		return result, err
	}

	err = db.View(func(tx *Tx) error {
		bContents := tx.Bucket(DbBucketContents)

		var content DbContent
		exists, err := bContents.Get([]byte(normURL), &content)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Not found url %s in bucket %s ", normURL, DbBucketContents)
		}

		r, err := zlib.NewReader(bytes.NewReader(content.Content))
		if err != nil {
			return err
		}
		result, err = ioutil.ReadAll(r)
		r.Close()
		return err
	})

	return result, err
}
