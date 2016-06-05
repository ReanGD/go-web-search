package crawler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerateURLByHostName(t *testing.T) {
	Convey("Success found url by hostname", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.FormValue("n") == "" {
				http.Redirect(w, r, "/?n=1", http.StatusFound)
			}
		}))
		defer ts.Close()

		parsedURL, err := url.Parse(ts.URL)
		So(err, ShouldBeNil)
		baseURL, err := GenerateURLByHostName(parsedURL.Host)
		So(err, ShouldBeNil)
		So(baseURL, ShouldEqual, ts.URL+"/?n=1")
	})

	Convey("Not found url by hostname", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/error", http.StatusFound)
		}))
		defer ts.Close()

		parsedURL, err := url.Parse(ts.URL)
		So(err, ShouldBeNil)
		_, err = GenerateURLByHostName(parsedURL.Host)
		So(err.Error(), ShouldEqual, "Get /error: stopped after 10 redirects")
	})
}

func TestNormalizeHostName(t *testing.T) {
	Convey("Normalize host name", t, func() {
		So(NormalizeHostName(""), ShouldEqual, "")
		So(NormalizeHostName("server1"), ShouldEqual, "server1")
		So(NormalizeHostName("SeRvEr1"), ShouldEqual, "server1")
		So(NormalizeHostName("www.server1"), ShouldEqual, "server1")
		So(NormalizeHostName("www"), ShouldEqual, "www")
	})
}

func TestNormalizeURL(t *testing.T) {
	Convey("Normalize URL", t, func() {
		So(NormalizeHostName(""), ShouldEqual, "")
		So(NormalizeHostName("http://SeRvEr1"), ShouldEqual, "http://server1")
	})
}
