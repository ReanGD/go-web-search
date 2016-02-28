package crawler

import (
	"bytes"
	"errors"
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

	return SavePage(url, body, links)
}

// Process - download and process cnt page
func Process(baseURL string, hostFilter string, cnt int) error {
	err := OpenDb()
	if err != nil {
		return err
	}
	defer CloseDb()

	isParsedBaseURL := false
	for i := 0; i != cnt; i++ {
		link, err := FindNotLoadedLink()
		if err != nil {
			return err
		}

		if link == "" {
			if isParsedBaseURL {
				return errors.New("Not found link for parse")
			}
			isParsedBaseURL = true
			link = baseURL
		}

		fmt.Println(link)
		err = processURL(link, hostFilter)
		if err != nil {
			return err
		}
	}

	return nil
}
