package crawler

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"math"
)

// CheckDb - check db
func CheckDb() error {
	db := new(DB)
	err := db.Open()
	if err != nil {
		return err
	}

	defer db.Close()

	isSuccess := true
	maxLen := 0
	var maxLenURL string
	minLen := math.MaxUint32
	var minLenURL string
	hashMap := make(map[[16]byte]string)
	err = db.View(func(tx *Tx) error {
		c := tx.Bucket(DbBucketContents).Cursor()

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
				minLenURL = string(url)
			}
			if lenContent > maxLen {
				maxLen = lenContent
				maxLenURL = string(url)
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
		fmt.Printf("Min len = %d (%s)\n", minLen, minLenURL)
		fmt.Printf("Max len = %d (%s)\n", maxLen, maxLenURL)
		fmt.Println("Checking ended successfully")
	}
	return err
}
