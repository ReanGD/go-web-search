package crawler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ReanGD/go-web-search/proxy"
	"github.com/ReanGD/go-web-search/werrors"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/temoto/robotstxt-go"
	"github.com/uber-go/zap"
)

type fakeDbHost struct {
	host         *proxy.Host
	baseURL      string
	robotTxtData string
}

func (f *fakeDbHost) GetHosts() map[int64]*proxy.Host {
	result := make(map[int64]*proxy.Host)
	if f.robotTxtData != "" {
		result[1] = proxy.NewHost("hostName", 200, []byte(f.robotTxtData))
	}

	return result
}

func (f *fakeDbHost) AddHost(host *proxy.Host, baseURL string) (int64, error) {
	f.host = host
	f.baseURL = baseURL

	return 1, nil
}

// TestResolveURL ...
func TestResolveURL(t *testing.T) {
	Convey("Success resolve", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.FormValue("n") == "" {
				http.Redirect(w, r, "/?n=1", http.StatusFound)
			}
		}))
		defer ts.Close()

		parsedURL, err := url.Parse(ts.URL)
		So(err, ShouldBeNil)

		h := &hostsManager{}
		resolvedURL, err := h.resolveURL(parsedURL.Host)
		So(err, ShouldBeNil)
		So(resolvedURL, ShouldEqual, ts.URL+"/?n=1")
	})
}

// TestReadRobotTxt ...
func TestReadRobotTxt(t *testing.T) {
	Convey("Success read robot txt", t, func() {
		expected := "robot txt data"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(expected))
		}))
		defer ts.Close()

		parsedURL, err := url.Parse(ts.URL)
		So(err, ShouldBeNil)

		h := &hostsManager{}
		statusCode, body, err := h.readRobotTxt(parsedURL.Host)
		So(err, ShouldBeNil)
		So(statusCode, ShouldEqual, 200)
		So(string(body), ShouldEqual, expected)
	})
}

// TestInitByHostName ...
func TestInitByHostName(t *testing.T) {
	Convey("Success", t, func() {
		robotstxtBody := []byte("User-agent: *")
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(robotstxtBody)
		}))
		defer ts.Close()

		var hostID int64
		hostID = 1
		baseURL := ts.URL
		parsedURL, err := url.Parse(baseURL)
		So(err, ShouldBeNil)

		h := &hostsManager{
			robotsTxt: make(map[int64]*robotstxt.Group),
			hosts:     make(map[string]int64)}
		db := &fakeDbHost{}
		hostName := parsedURL.Host
		hostsExpected := make(map[string]int64)
		hostsExpected[hostName] = hostID
		robot, err := robotstxt.FromStatusAndBytes(200, robotstxtBody)
		So(err, ShouldBeNil)
		robotsTxtExpected := make(map[int64]*robotstxt.Group)
		robotsTxtExpected[hostID] = robot.FindGroup("Googlebot")
		host := proxy.NewHost(hostName, 200, robotstxtBody)

		err = h.initByHostName(db, hostName)
		So(err, ShouldBeNil)
		So(h.hosts, ShouldResemble, hostsExpected)
		So(h.robotsTxt, ShouldResemble, robotsTxtExpected)
		So(db.host, ShouldResemble, host)
		So(db.baseURL, ShouldEqual, baseURL)
	})

	Convey("Failed resolve base URL by status code", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "error status", 500)
		}))
		defer ts.Close()

		parsedURL, err := url.Parse(ts.URL)
		So(err, ShouldBeNil)

		h := &hostsManager{}
		db := &fakeDbHost{}

		err = h.initByHostName(db, parsedURL.Host)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrResolveBaseURL)
	})

	Convey("Failed read robot txt", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.Error(w, "error", 10)
			}
		}))
		defer ts.Close()

		parsedURL, err := url.Parse(ts.URL)
		So(err, ShouldBeNil)

		h := &hostsManager{}
		db := &fakeDbHost{}

		err = h.initByHostName(db, parsedURL.Host)
		So(err, ShouldNotBeNil)
		werr := err.(*werrors.ErrorEx)
		So(werr.Fields, ShouldEqual, []zap.Field{})
		So(err.Error(), ShouldEqual, ErrCreateRobotsTxtFromURL)
	})

	Convey("Failed create robots txt", t, func() {
		robotstxtBody := []byte("Disallow:without_user_agent")
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(robotstxtBody)
		}))
		defer ts.Close()

		parsedURL, err := url.Parse(ts.URL)
		So(err, ShouldBeNil)

		h := &hostsManager{}
		db := &fakeDbHost{}

		err = h.initByHostName(db, parsedURL.Host)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrCreateRobotsTxtFromURL)
	})
}

// TestInitByDb ...
func TestInitByDb(t *testing.T) {
	Convey("Success", t, func() {
		h := &hostsManager{
			robotsTxt: make(map[int64]*robotstxt.Group),
			hosts:     make(map[string]int64)}
		db := &fakeDbHost{robotTxtData: "message"}

		var hostID int64
		hostID = 1
		hostsExpected := make(map[string]int64)
		hostsExpected["hostName"] = hostID
		robot, err := robotstxt.FromStatusAndBytes(200, []byte(db.robotTxtData))
		So(err, ShouldBeNil)
		robotsTxtExpected := make(map[int64]*robotstxt.Group)
		robotsTxtExpected[hostID] = robot.FindGroup("Googlebot")

		err = h.initByDb(db)
		So(err, ShouldBeNil)
		So(h.hosts, ShouldResemble, hostsExpected)
		So(h.robotsTxt, ShouldResemble, robotsTxtExpected)
	})
}

// TestInitHostManager ...
func TestInitHostManager(t *testing.T) {
	Convey("Failed init by many redirect", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/error", http.StatusFound)
		}))
		defer ts.Close()

		parsedURL, err := url.Parse(ts.URL)
		So(err, ShouldBeNil)

		h := &hostsManager{}
		db := &fakeDbHost{}

		err = h.Init(db, []string{parsedURL.Host, "error_host"})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrGetRequest)
	})

	Convey("Failed init error robots txt data", t, func() {
		h := &hostsManager{}
		db := &fakeDbHost{robotTxtData: "Disallow:without_user_agent"}

		err := h.Init(db, []string{})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrCreateRobotsTxtFromDb)
	})

	Convey("Empty hosts", t, func() {
		h := &hostsManager{}
		db := &fakeDbHost{}

		err := h.Init(db, []string{})
		So(err, ShouldBeNil)
		So(h.hosts, ShouldBeEmpty)
		So(h.robotsTxt, ShouldBeEmpty)
	})
}
