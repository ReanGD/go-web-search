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

func helperNormalizeURL(in string, out string) {
	u, err := url.Parse(in)
	So(err, ShouldBeNil)
	So(NormalizeURL(u), ShouldEqual, out)
}

func TestNormalizeURL(t *testing.T) {
	Convey("Normalize URL", t, func() {
		helperNormalizeURL("", "")
		helperNormalizeURL("http://SeRvEr1", "http://server1")
	})

	Convey("Remove utm", t, func() {
		helperNormalizeURL(
			"http://s/utm_campaign",
			"http://s/utm_campaign")

		helperNormalizeURL(
			"http://s/?utm_source",
			"http://s/")

		helperNormalizeURL(
			"http://s?utm_source=1&utm_source1=1",
			"http://s?utm_source1=1")

		helperNormalizeURL(
			"http://s?utm_medium1=1&utm_medium=1&utm_medium2=1",
			"http://s?utm_medium1=1&utm_medium2=1")

		helperNormalizeURL(
			"http://s?utm_term=1&utm_content=1",
			"http://s")

		helperNormalizeURL(
			"http://s?utm_campaign=1",
			"http://s")
	})
}
