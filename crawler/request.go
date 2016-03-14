package crawler

import (
	"crypto/md5"
	"database/sql"
	"errors"
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
	Robot       *robotTxt
	client      *http.Client
	redirectCnt int
}

func (r *request) parsePage(body []byte, baseURL string) (map[string]string, error) {
	result := make(map[string]string)

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

		if (parsed.Scheme == "http" || parsed.Scheme == "https") && urlStr != baseURL {
			result[urlStr] = NormalizeHostName(parsed.Host)
		}
	}

	return result, nil
}

func (r *request) get(urlStr string) (*content.Meta, error) {
	r.redirectCnt = 0
	result := content.Meta{
		URL:         urlStr,
		MIME:        sql.NullString{Valid: false},
		Timestamp:   time.Now(),
		RedirectCnt: sql.NullInt64{Valid: false},
		StatusCode:  sql.NullInt64{Valid: false}}

	allow, err := r.Robot.Test(urlStr)
	if err != nil {
		result.State = content.StateErrorURLFormat
		return &result, err
	}
	if !allow {
		result.State = content.StateDisabledByRobotsTxt
		log.Printf("INFO: URL %s blocked by robot.txt", urlStr)
		return &result, nil
	}

	request, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		result.State = content.StateErrorURLFormat
		return &result, err
	}

	// response, err := http.Get(urlStr)
	response, err := r.client.Do(request)
	if err != nil {
		result.State = content.StateConnectError
		return &result, err
	}
	result.RedirectCnt = sql.NullInt64{Int64: int64(r.redirectCnt), Valid: true}
	defer response.Body.Close()

	result.StatusCode = sql.NullInt64{Int64: int64(response.StatusCode), Valid: true}
	if response.StatusCode != 200 {
		result.State = content.StateErrorStatusCode
		return &result, fmt.Errorf("StatusCode = %d", response.StatusCode)
	}

	contentType, ok := response.Header["Content-Type"]
	if !ok {
		result.MIME = sql.NullString{String: "not found", Valid: true}
		result.State = content.StateUnsupportedFormat
		return &result, fmt.Errorf("Not found Content-Type in headers")
	}

	mediatype, _, err := mime.ParseMediaType(contentType[0])
	if err != nil {
		result.MIME = sql.NullString{String: "parse error", Valid: true}
		result.State = content.StateUnsupportedFormat
		return &result, err
	}

	result.MIME = sql.NullString{String: mediatype, Valid: true}
	if mediatype != "text/html" {
		result.State = content.StateUnsupportedFormat
		log.Printf("INFO: URL %s has unsupported mime format = %s", urlStr, mediatype)
		return &result, nil
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
func (r *request) Process(urlStr string) *content.PageData {
	meta, err := r.get(urlStr)
	if err != nil {
		log.Printf("ERROR: Get URL %s, message: %s", urlStr, err)
	}

	urls := make(map[string]string)
	if meta.State == content.StateSuccess {
		urls, err = r.parsePage(meta.Content.Data.Data, urlStr)
		if err != nil {
			meta.State = content.StateParseError
			log.Printf("ERROR: Parse URL %s, message: %s", urlStr, err)
		}
	}

	return &content.PageData{
		HostName: r.Robot.HostName,
		MetaItem: meta,
		URLs:     urls}
}

// Init - init request structure
func (r *request) Init() {
	r.client = new(http.Client)
	r.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		if len(via) == 0 {
			return nil
		}
		r.redirectCnt++
		return nil
	}
}
