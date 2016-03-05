package crawler

import "net/url"

type taskFromDB struct {
	Host string
	URL  string
}

func (db *DB) readNotLoadedURLs(baseHosts []string, cnt int, ch chan<- taskFromDB) error {
	cntPerHost := cnt / len(baseHosts)
	if cntPerHost < 1 {
		cntPerHost = 1
	}

	leftCnts := make(map[string]int)
	for _, host := range baseHosts {
		leftCnts[host] = cntPerHost
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
				baseURL := url.URL{Host: host, Scheme: "http"}
				exists, err := b.Get([]byte(baseURL.String()), &urlVal)
				if err != nil {
					return err
				}
				if !exists {
					ch <- taskFromDB{Host: host, URL: baseURL.String()}
				}
			}
		}

		return nil
	})
}
