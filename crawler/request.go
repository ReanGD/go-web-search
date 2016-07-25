package crawler

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/ReanGD/go-web-search/database"
	"github.com/ReanGD/go-web-search/proxy"
	"github.com/ReanGD/go-web-search/werrors"
	"github.com/uber-go/zap"
)

type request struct {
	hostMng *hostsManager
	client  *http.Client
	meta    *proxy.Meta
	urls    map[string]sql.NullInt64
	logger  zap.Logger
}

func (r *request) get(u *url.URL) (int64, error) {
	urlStr := u.String()
	r.urls = make(map[string]sql.NullInt64)
	hostID, robotOk := r.hostMng.CheckURL(u)
	r.meta = proxy.NewMeta(hostID, urlStr, nil)

	if !hostID.Valid {
		r.meta.SetState(database.StateExternal)
	}
	if !robotOk {
		r.meta.SetState(database.StateDisabledByRobotsTxt)
		log.Printf("INFO: URL %s blocked by robot.txt", urlStr)
		return 0, nil
	}

	startTime := time.Now()
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
		r.meta.SetState(database.StateConnectError)
		return 0, err
	}

	if r.meta.GetState() != database.StateSuccess {
		// here or early - logging!!!
		return 0, nil
	}

	// hostID, robotOk := r.hostMng.CheckURL(u)
	// 		if !robotOk {
	// 		r.meta.SetState(database.StateDisabledByRobotsTxt)
	// 		return fmt.Errorf("INFO: URL %s blocked by robot.txt", NormalizeURL(&copyURL))
	// 	}

	loggerURL := r.logger.With(zap.String("url", r.meta.GetURL()))

	RequestDurationMs := int64(time.Since(startTime) / time.Millisecond)
	loggerURL.Debug(DbgRequestDuration, zap.Int64("duration", RequestDurationMs))

	parser := newResponseParser(loggerURL, r.hostMng, r.meta)
	err = parser.Run(response)
	if err == nil {
		r.urls = parser.URLs
	} else {
		werrors.LogError(loggerURL, err)
		err = nil
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

		r.meta.SetState(database.StateDublicate)
		r.meta.SetStatusCode(301)

		hostID, _ := r.hostMng.CheckURL(req.URL)
		copyURL := *req.URL
		r.meta = proxy.NewMeta(hostID, NormalizeURL(&copyURL), r.meta)
		if !hostID.Valid {
			r.meta.SetState(database.StateExternal)
		}

		for attr, val := range via[0].Header {
			if _, ok := req.Header[attr]; !ok {
				req.Header[attr] = val
			}
		}

		return nil
	}
}
