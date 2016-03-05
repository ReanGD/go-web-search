package crawler

import (
	"fmt"
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
		fmt.Printf("Contents len = %d\n", bContents.Stats().KeyN)

		bURLs := tx.Bucket(DbBucketURLs)
		fmt.Printf("bURLs len = %d\n", bURLs.Stats().KeyN)

		bWrongURLs := tx.Bucket(DbBucketWrongURLs)
		fmt.Printf("WrongURLs len = %d\n", bWrongURLs.Stats().KeyN)
		return nil
	})
}
