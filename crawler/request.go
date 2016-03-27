package crawler

import (
	"compress/gzip"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/ReanGD/go-web-search/content"
)

type request struct {
	Robot  *robotTxt
	client *http.Client
	meta   *content.Meta
	urls   map[string]string
}

func (r *request) get(u *url.URL) error {
	urlStr := u.String()
	r.urls = make(map[string]string)
	r.meta = &content.Meta{
		URL:             urlStr,
		Timestamp:       time.Now(),
		RedirectReferer: nil,
		RedirectCnt:     0,
		HostName:        NormalizeHostName(u.Host),
		StatusCode:      sql.NullInt64{Valid: false}}

	if !r.Robot.Test(u) {
		r.meta.State = content.StateDisabledByRobotsTxt
		log.Printf("INFO: URL %s blocked by robot.txt", urlStr)
		return nil
	}

	request := &http.Request{
		Method:     "GET",
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: map[string][]string{
			"User-Agent":      {"Mozilla/5.0 (compatible; GoWebSearch/0.1)"},
			"Accept":          {"text/html;q=0.9,*/*;q=0.1"},
			"Accept-Encoding": {"gzip"},
			"Accept-Language": {"ru-RU,ru;q=0.9,en-US;q=0.2,en;q=0.1"},
			"Accept-Charset":  {"utf-8;q=0.9,windows-1251;q=0.8,koi8-r;q=0.7,*;q=0.1"},
		},
		Body: nil,
		Host: u.Host,
	}

	response, err := r.client.Do(request)
	if err != nil {
		r.meta.State = content.StateConnectError
		return err
	}
	defer response.Body.Close()
	if r.meta.URL != urlStr {
		urlStr = urlStr + "->" + r.meta.URL
	}

	r.meta.StatusCode = sql.NullInt64{Int64: int64(response.StatusCode), Valid: true}
	if response.StatusCode != 200 {
		r.meta.State = content.StateErrorStatusCode
		return fmt.Errorf("StatusCode = %d", response.StatusCode)
	}

	contentType, ok := response.Header["Content-Type"]
	if !ok {
		r.meta.State = content.StateUnsupportedFormat
		return fmt.Errorf("Not found Content-Type in headers")
	}

	mediatype, _, err := mime.ParseMediaType(contentType[0])
	if err != nil {
		r.meta.State = content.StateUnsupportedFormat
		return fmt.Errorf("Parse Content-Type, error %s", err)
	}

	if mediatype != "text/html" {
		r.meta.State = content.StateUnsupportedFormat
		log.Printf("INFO: URL %s has unsupported mime format = %s", urlStr, mediatype)
		return nil
	}

	var body []byte
	contentEncoding, ok := response.Header["Content-Encoding"]
	if ok && contentEncoding[0] == "gzip" {
		reader, err := gzip.NewReader(response.Body)
		if err != nil {
			r.meta.State = content.StateAnswerError
			return err
		}
		body, err = ioutil.ReadAll(reader)
		reader.Close()
	} else {
		body, err = ioutil.ReadAll(response.Body)
	}
	if err != nil {
		r.meta.State = content.StateAnswerError
		return err
	}

	parser := new(HTMLParser)
	err = parser.Parse(body, u)
	if err != nil {
		r.meta.State = content.StateParseError
		return fmt.Errorf("Html parse error: %s", err)
	}
	if !parser.MetaTagIndex {
		r.meta.State = content.StateNoFollow
		return nil
	}

	hash := md5.Sum(body)
	r.urls = parser.URLs
	r.meta.State = content.StateSuccess
	r.meta.Content = content.Content{
		Hash: string(hash[:]),
		Body: content.Compressed{Data: body}}

	return nil
}

// Send - load and parse the urlStr
// urlStr - valid URL
func (r *request) Process(u *url.URL) *content.PageData {
	err := r.get(u)
	if err != nil {
		log.Printf("ERROR: Get URL %s, message: %s", u.String(), err)
	}

	return &content.PageData{MetaItem: r.meta, URLs: r.urls}
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

		r.meta.State = content.StateDublicate
		r.meta.StatusCode = sql.NullInt64{Int64: 301, Valid: true}

		copyURL := *req.URL
		r.meta = &content.Meta{
			URL:             NormalizeURL(&copyURL),
			Timestamp:       time.Now(),
			RedirectReferer: r.meta,
			RedirectCnt:     0,
			HostName:        NormalizeHostName(req.URL.Host),
			StatusCode:      sql.NullInt64{Valid: false}}

		currentMeta := r.meta.RedirectReferer
		for currentMeta != nil {
			currentMeta.RedirectCnt++
			currentMeta = currentMeta.RedirectReferer
		}

		for attr, val := range via[0].Header {
			if _, ok := req.Header[attr]; !ok {
				req.Header[attr] = val
			}
		}

		return nil
	}
}
