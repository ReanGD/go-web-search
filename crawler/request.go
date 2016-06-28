package crawler

import (
	"errors"
	"log"
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
	logger zap.Logger
}

func (r *request) get(u *url.URL) (int64, error) {
	urlStr := u.String()
	r.urls = make(map[string]string)
	r.meta = proxy.NewMeta(NormalizeHostName(u.Host), urlStr, nil)

	if !r.Robot.Test(u) {
		r.meta.SetState(proxy.StateDisabledByRobotsTxt)
		log.Printf("INFO: URL %s blocked by robot.txt", urlStr)
		return 0, nil
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
		return 0, err
	}

	parser := newResponseParser(r.logger, r.meta)
	err = parser.Run(response)
	if err == nil {
		r.urls = parser.URLs
	}

	return parser.BodyDurationMs, err
}

// Send - load and parse the urlStr
// urlStr - valid URL
func (r *request) Process(u *url.URL) (*proxy.PageData, int64) {
	duration, err := r.get(u)
	if err != nil {
		log.Printf("ERROR: Get URL %s, message: %s", u.String(), err)
	}

	return proxy.NewPageData(r.meta, r.urls), duration
}

// Init - init request structure
func (r *request) Init(logger zap.Logger) {
	r.client = new(http.Client)
	r.logger = logger
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
