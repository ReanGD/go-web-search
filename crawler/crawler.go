package crawler

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
)

func processURL(url string, hostFilter string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return err
	}

	rawLinks, err := ParseURLsInPage(bytes.NewReader(body))
	if err != nil {
		return err
	}

	links, err := ProcessLinks(url, rawLinks, hostFilter)
	if err != nil {
		return err
	}

	return SavePage(url, body, md5.Sum(body), links)
}

// Process - download and process cnt page
func Process(url string, hostFilter string, cnt int) error {
	err := OpenDb()
	if err != nil {
		return err
	}
	defer CloseDb()

	fmt.Println(url)
	err = processURL(url, hostFilter)
	if err != nil {
		return err
	}

	for i := 0; i != cnt; i++ {
		link, err := FindNotLoadedLink()
		if err != nil {
			return err
		}

		fmt.Println(link)
		err = processURL(link, hostFilter)
		if err != nil {
			return err
		}
	}

	return nil
}
