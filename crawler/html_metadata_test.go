package crawler

import (
	"bytes"
	"database/sql"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/uber-go/zap"
)

func helpNewHTMLMetadata() *HTMLMetadata {
	hostMng := &hostsManager{hosts: map[string]int64{"testhost1": 1}}
	h, err := NewHTMLMetadata(hostMng, "http://testhost1/test/")
	So(err, ShouldBeNil)

	return h
}

// TestTitle ...
func TestTitle(t *testing.T) {
	Convey("Title as url", t, func() {
		So(helpNewHTMLMetadata().GetTitle(), ShouldEqual, "http://testhost1/test/")
	})

	Convey("Set title with rewrite", t, func() {
		h := helpNewHTMLMetadata()
		h.SetTitle("title1", true)
		So(h.GetTitle(), ShouldEqual, "title1")
		h.SetTitle("title2", true)
		So(h.GetTitle(), ShouldEqual, "title2")
	})

	Convey("Set empty title with rewrite", t, func() {
		h := helpNewHTMLMetadata()
		h.SetTitle("title1", true)
		h.SetTitle("", true)
		So(h.GetTitle(), ShouldEqual, "title1")
	})

	Convey("Set title without rewrite", t, func() {
		h := helpNewHTMLMetadata()
		h.SetTitle("title1", true)
		h.SetTitle("title2", false)
		So(h.GetTitle(), ShouldEqual, "title1")
	})

	Convey("Long title", t, func() {
		h := helpNewHTMLMetadata()
		t := "!0000000000111111111122222222223333333333444444444455555555556666666666777777777788888888889999999999"
		h.SetTitle(t, true)
		out := "!000000000011111111112222222222333333333344444444445555555555666666666677777777778888888888999999..."
		So(h.GetTitle(), ShouldEqual, out)
	})
}

// TestAddURL ...
func TestAddURL(t *testing.T) {
	Convey("Add URLs", t, func() {
		h := helpNewHTMLMetadata()

		h.AddURL("/link1")
		h.AddURL("")
		h.AddURL("link2")
		h.AddURL("link3/link3")
		h.AddURL("http://testhost2/link4")
		h.AddURL("  link5  ")
		h.AddURL("/wrong%9")

		hostIDValid := sql.NullInt64{Int64: 1, Valid: true}
		hostIDInvalid := sql.NullInt64{Valid: false}

		expectedURLs := make(map[string]sql.NullInt64)
		expectedURLs["http://testhost1/link1"] = hostIDValid
		expectedURLs["http://testhost1/test/link2"] = hostIDValid
		expectedURLs["http://testhost1/test/link3/link3"] = hostIDValid
		expectedURLs["http://testhost2/link4"] = hostIDInvalid
		expectedURLs["http://testhost1/test/link5"] = hostIDValid
		So(h.URLs, ShouldResemble, expectedURLs)

		expectedWrongURLs := make(map[string]string)
		expectedWrongURLs["/wrong%9"] = `parse /wrong%9: invalid URL escape "%9"`
		So(h.wrongURLs, ShouldResemble, expectedWrongURLs)
	})

	Convey("Wrong URLs to log", t, func() {
		buf := &bytes.Buffer{}
		logger := zap.NewJSON(zap.DebugLevel, zap.Output(zap.AddSync(buf)))
		logger.StubTime()

		h := helpNewHTMLMetadata()
		h.AddURL("/link1")
		h.AddURL("/wrong%2")
		h.WrongURLsToLog(logger)

		expected := `{"msg":"Error parse URL","level":"warn","ts":0,"fields":{"err_url":"/wrong%2","details":"parse /wrong%2: invalid URL escape \"%2\""}}
`
		So(string(buf.Bytes()), ShouldEqual, expected)
	})
}

// TestInitError ...
func TestInitError(t *testing.T) {
	Convey("Error parse base URL", t, func() {
		hostMng := &hostsManager{}
		_, err := NewHTMLMetadata(hostMng, "/wrong%2")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrParseBaseURL)
	})
}

// TestClearURLs ...
func TestClearURLs(t *testing.T) {
	Convey("Clear not empty URLs", t, func() {
		urls := map[string]sql.NullInt64{"hostName": sql.NullInt64{Valid: false}}
		h := &HTMLMetadata{URLs: urls}
		So(len(h.URLs), ShouldEqual, 1)
		h.ClearURLs()
		So(len(h.URLs), ShouldEqual, 0)
	})

	Convey("Clear empty URLs", t, func() {
		h := &HTMLMetadata{URLs: make(map[string]sql.NullInt64)}
		So(len(h.URLs), ShouldEqual, 0)
		h.ClearURLs()
		So(len(h.URLs), ShouldEqual, 0)
	})
}
