package crawler

import (
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"

	"github.com/ReanGD/go-web-search/proxy"
	"github.com/uber-go/zap"
)

type request struct {
	Robot  *robotTxt
	client *http.Client
	meta   *proxy.InMeta
	urls   map[string]string
}

func (r *request) get(u *url.URL) error {
	urlStr := u.String()
	r.urls = make(map[string]string)
	r.meta = proxy.NewMeta(NormalizeHostName(u.Host), urlStr, nil)

	if !r.Robot.Test(u) {
		r.meta.SetState(proxy.StateDisabledByRobotsTxt)
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
			"Accept-Encoding": {"gzip;q=0.9,identity;q=0.5,*;q=0.1"},
			"Accept-Language": {"ru-RU,ru;q=0.9,en-US;q=0.2,en;q=0.1"},
			"Accept-Charset":  {"utf-8;q=0.9,windows-1251;q=0.8,koi8-r;q=0.7,*;q=0.1"},
		},
		Body: nil,
		Host: u.Host,
	}

	response, err := r.client.Do(request)
	if err != nil {
		r.meta.SetState(proxy.StateConnectError)
		return err
	}
	defer response.Body.Close()
	if r.meta.GetURL() != urlStr {
		urlStr = urlStr + "->" + r.meta.GetURL()
	}

	r.meta.SetStatusCode(response.StatusCode)
	if response.StatusCode != 200 {
		r.meta.SetState(proxy.StateErrorStatusCode)
		return fmt.Errorf("StatusCode = %d", response.StatusCode)
	}

	contentType, ok := response.Header["Content-Type"]
	if !ok {
		r.meta.SetState(proxy.StateUnsupportedFormat)
		return fmt.Errorf("Not found Content-Type in headers")
	}

	mediatype, _, err := mime.ParseMediaType(contentType[0])
	if err != nil {
		r.meta.SetState(proxy.StateUnsupportedFormat)
		return fmt.Errorf("Parse Content-Type, error %s", err)
	}

	if mediatype != "text/html" {
		r.meta.SetState(proxy.StateUnsupportedFormat)
		log.Printf("INFO: URL %s has unsupported mime format = %s", urlStr, mediatype)
		return nil
	}

	contentEncoding := ""
	contentEncodingArr, ok := response.Header["Content-Encoding"]
	if ok && len(contentEncodingArr) != 0 {
		contentEncoding = contentEncodingArr[0]
	}

	body, err := readBody(contentEncoding, response.Body)
	if err != nil {
		r.meta.SetState(proxy.StateAnswerError)
		return err
	}

	parser, state, err := ProcessBody(zap.NewJSON(), body, contentType, u)
	r.meta.SetState(state)
	if err != nil {
		return err
	}

	r.urls = parser.URLs
	r.meta.SetContent(proxy.NewContent(body, parser.Title))

	return nil
}

// Send - load and parse the urlStr
// urlStr - valid URL
func (r *request) Process(u *url.URL) *proxy.PageData {
	err := r.get(u)
	if err != nil {
		log.Printf("ERROR: Get URL %s, message: %s", u.String(), err)
	}

	return proxy.NewPageData(r.meta, r.urls)
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

		r.meta.SetState(proxy.StateDublicate)
		r.meta.SetStatusCode(301)

		copyURL := *req.URL
		r.meta = proxy.NewMeta(NormalizeHostName(req.URL.Host), NormalizeURL(&copyURL), r.meta)

		for attr, val := range via[0].Header {
			if _, ok := req.Header[attr]; !ok {
				req.Header[attr] = val
			}
		}

		return nil
	}
}
