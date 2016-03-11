package crawler

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ReanGD/go-web-search/content"
)

type request struct {
	Robot *robotTxt
}

func (r *request) parsePage(body []byte, baseURL string) (map[string]*content.URL, error) {
	result := make(map[string]*content.URL)

	parser := new(pageURLs)
	rawURLs, err := parser.Parse(body)
	if err != nil {
		return result, err
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return result, err
	}

	for itRawURL := rawURLs.Front(); itRawURL != nil; itRawURL = itRawURL.Next() {
		relative, err := url.Parse(strings.TrimSpace(itRawURL.Value.(string)))
		if err != nil {
			log.Printf("ERROR: Parse URL on page %s, message: %s", baseURL, err)
			continue
		}
		parsed := base.ResolveReference(relative)
		urlStr := NormalizeURL(parsed)
		parsed, err = url.Parse(urlStr)

		if parsed.Scheme == "http" && r.Host.Name == parsed.Host && urlStr != baseURL {
			result[urlStr] = &content.URL{ID: urlStr, HostID: r.Host.ID, Loaded: false}
		}
	}

	return result, nil
}

func (r *request) get(urlStr string) (*content.Meta, error) {
	result := content.Meta{URL: urlStr, Timestamp: time.Now()}

	allow, err := r.Robot.Test(urlStr)
	if err != nil {
		result.State = content.StateErrorURLFormat
		return &result, err
	}
	if !allow {
		result.State = content.StateDisabledByRobotsTxt
		return &result, fmt.Errorf("Blocked by robot.txt")
	}

	response, err := http.Get(urlStr)
	if err != nil {
		result.State = content.StateConnectError
		return &result, err
	}
	defer response.Body.Close()

	result.StatusCode = response.StatusCode
	if response.StatusCode != 200 {
		result.State = content.StateErrorStatusCode
		return &result, fmt.Errorf("StatusCode = %d", response.StatusCode)
	}

	contentType, ok := response.Header["Content-Type"]
	if !ok {
		result.MIME = "not found"
		result.State = content.StateUnsupportedFormat
		return &result, fmt.Errorf("Not found Content-Type in headers")
	}

	mediatype, _, err := mime.ParseMediaType(contentType[0])
	if err != nil {
		result.MIME = "parse error"
		result.State = content.StateUnsupportedFormat
		return &result, err
	}

	result.MIME = mediatype
	if mediatype != "text/html" {
		result.State = content.StateUnsupportedFormat
		return &result, fmt.Errorf("Unsupported mime format = %s", mediatype)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		result.State = content.StateAnswerError
		return &result, err
	}

	result.State = content.StateSuccess
	hash := md5.Sum(body)
	result.Content = content.Content{Hash: string(hash[:]), Data: content.Compressed{Data: body}}

	return &result, nil
}

// Send - load and parse the urlStr
// urlStr - valid URL
func (r *request) Process(urlStr string) (*content.Meta, map[string]*content.URL) {
	var urls map[string]*content.URL
	meta, err := r.get(urlStr)
	if err != nil {
		log.Printf("ERROR: Get URL %s, message: %s", urlStr, err)
	}

	if meta.State == content.StateSuccess {
		urls, err = r.parsePage(meta.Content.Data.Data, urlStr)
		if err != nil {
			meta.State = content.StateParseError
			log.Printf("ERROR: Parse URL %s, message: %s", urlStr, err)
		}
	}

	return meta, urls
}
