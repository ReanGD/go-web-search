package crawler

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/ReanGD/go-web-search/content"
	"github.com/temoto/robotstxt-go"
)

type request struct {
	RobotTxt *robotstxt.Group
}

// Get - get urlStr and save page to content.Meta
// urlStr - valid URL
func (r *request) Get(parentID uint64, urlStr string) (*content.Meta, error) {
	result := content.Meta{URL: urlStr, Parent: parentID, Timestamp: time.Now()}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		result.State = content.StateErrorURLFormat
		return &result, err
	}

	parsed.Scheme = ""
	parsed.Host = ""
	if !r.RobotTxt.Test(parsed.String()) {
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
