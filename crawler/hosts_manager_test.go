package crawler

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ReanGD/go-web-search/proxy"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/temoto/robotstxt-go"
)

type fakeDbHost struct {
	host         *proxy.Host
	baseURL      string
	robotTxtData string
	getHostErr   string
}

func (f *fakeDbHost) GetHosts() (map[int64]*proxy.Host, error) {
	result := make(map[int64]*proxy.Host)
	if f.getHostErr != "" {
		return result, errors.New(f.getHostErr)
	}
	if f.robotTxtData != "" {
		result[1] = proxy.NewHost("hostName", 200, []byte(f.robotTxtData))
	}

	return result, nil
}

func (f *fakeDbHost) AddHost(host *proxy.Host, baseURL string) (int64, error) {
	f.host = host
	f.baseURL = baseURL

	return 1, nil
}

// TestResolveHost ...
func TestResolveHost(t *testing.T) {
	Convey("Check resolve hosts", t, func() {
		h := &hostsManager{hosts: map[string]int64{"hostname1": 1, "hostname2": 2}}
		So(h.ResolveHost("www.hostName1"), ShouldResemble, sql.NullInt64{Int64: 1, Valid: true})
		So(h.ResolveHost("HoStNaMe2"), ShouldResemble, sql.NullInt64{Int64: 2, Valid: true})
		So(h.ResolveHost("hostName3"), ShouldResemble, sql.NullInt64{Valid: false})
	})
}

// TestCheckURL ...
func TestCheckURL(t *testing.T) {
	Convey("Check resolve hosts", t, func() {
		robotstxtBody := `
User-agent: Yandex
Disallow: /search2/

User-agent: *
Disallow: /search3/

User-agent: Googlebot
Disallow: /search1/`
		robotGoogle, err := robotstxt.FromStatusAndBytes(200, []byte(robotstxtBody))
		So(err, ShouldBeNil)

		robotstxtBody = `User-agent: Yandex
Disallow: /search/`
		robotYandex, err := robotstxt.FromStatusAndBytes(200, []byte(robotstxtBody))
		So(err, ShouldBeNil)

		robotstxtBody = `User-agent: *
Disallow: /search/`
		robotAll, err := robotstxt.FromStatusAndBytes(200, []byte(robotstxtBody))
		So(err, ShouldBeNil)

		hosts := map[string]int64{
			"google": 1,
			"yandex": 2,
			"all":    3}
		robotsTxt := map[int64]*robotstxt.Group{
			1: robotGoogle.FindGroup("Googlebot"),
			2: robotYandex.FindGroup("Googlebot"),
			3: robotAll.FindGroup("Googlebot")}

		h := &hostsManager{hosts: hosts, robotsTxt: robotsTxt}

		u := &url.URL{Host: "google", Path: "/search1/"}
		_, access := h.CheckURL(u)
		So(access, ShouldBeFalse)

		u = &url.URL{Host: "google", Path: "/search2/"}
		_, access = h.CheckURL(u)
		So(access, ShouldBeTrue)

		u = &url.URL{Host: "google", Path: "/search3/"}
		_, access = h.CheckURL(u)
		So(access, ShouldBeTrue)

		u = &url.URL{Host: "yandex", Path: "/search/"}
		_, access = h.CheckURL(u)
		So(access, ShouldBeTrue)

		u = &url.URL{Host: "all", Path: "/search/"}
		_, access = h.CheckURL(u)
		So(access, ShouldBeFalse)

		u = &url.URL{Host: "all", Path: "/allowed/"}
		_, access = h.CheckURL(u)
		So(access, ShouldBeTrue)

		u = &url.URL{Host: "unknown", Path: "/allowed/"}
		_, access = h.CheckURL(u)
		So(access, ShouldBeFalse)
	})
}

// TestGetHosts ...
func TestGetHosts(t *testing.T) {
	Convey("Success resolve", t, func() {
		hosts := map[string]int64{"hostName": 1}
		h := &hostsManager{hosts: hosts}
		So(h.GetHosts(), ShouldResemble, hosts)
	})
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
				http.Redirect(w, r, "/error", http.StatusFound)
			}
		}))
		defer ts.Close()

		parsedURL, err := url.Parse(ts.URL)
		So(err, ShouldBeNil)

		h := &hostsManager{}
		db := &fakeDbHost{}

		err = h.initByHostName(db, parsedURL.Host)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrGetRequest)
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

	Convey("Failed by db.GetHosts", t, func() {
		h := &hostsManager{}
		db := &fakeDbHost{getHostErr: "get host error"}

		err := h.initByDb(db)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, db.getHostErr)
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
