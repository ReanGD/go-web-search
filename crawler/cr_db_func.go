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

func (db *DB) readNotLoadedHosts(baseHosts map[string]int) ([]string, error) {
	result := make([]string, 0, len(baseHosts))
	err := db.View(func(tx *Tx) error {
		b := tx.Bucket(DbBucketHosts)

		var dbHost DbHost
		for host := range baseHosts {
			exists, err := b.Get([]byte(host), &dbHost)
			if err != nil {
				return err
			}
			if !exists {
				result = append(result, host)
			}

		}
		return nil
	})

	return result, err
}

func (db *DB) readHosts(baseHosts map[string]int) (map[string]*DbHost, error) {
	result := make(map[string]*DbHost, len(baseHosts))
	err := db.View(func(tx *Tx) error {
		b := tx.Bucket(DbBucketHosts)

		for host := range baseHosts {
			var dbHost DbHost
			exists, err := b.Get([]byte(host), &dbHost)
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("Not found robots.txt for host %s ", host)
			}
			result[host] = &dbHost
		}
		return nil
	})

	return result, err
}

func (db *DB) writeHosts(baseHosts map[string]*DbHost) error {
	return db.Update(func(tx *Tx) error {
		b := tx.Bucket(DbBucketHosts)

		for key, val := range baseHosts {
			err := b.Put([]byte(key), val)
			if err != nil {
				return err
			}
		}
		return nil
	})
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
		fmt.Printf("\nStatistics:\n")

		bHosts := tx.Bucket(DbBucketHosts)
		fmt.Printf("Hosts: %d\n", bHosts.Stats().KeyN)

		bContents := tx.Bucket(DbBucketContents)
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
