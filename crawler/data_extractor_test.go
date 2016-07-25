package crawler

import (
	"bytes"
	"database/sql"
	"testing"

	"golang.org/x/net/html"

	. "github.com/smartystreets/goconvey/convey"
)

func helperRunDataExtrator(htmlStr string) *HTMLMetadata {
	node, err := html.Parse(bytes.NewReader([]byte(htmlStr)))
	So(err, ShouldBeNil)

	hostMng := &hostsManager{hosts: map[string]int64{"testhost1": 1}}
	meta, err := RunDataExtrator(hostMng, node, "http://testhost1/test/")
	So(err, ShouldBeNil)

	return meta
}

// TestURLs ...
func TestURLs(t *testing.T) {
	Convey("TestURLs", t, func() {
		meta := helperRunDataExtrator(`<html><head>
<link rel="next" href="/link1">
<link rel="prev" href="link2">
<link rel="previous" href="link3/link3">
</head><body>
<a href="/link4">text</a>
<div><a href="link5">text</a></div>
<div><a href="http://testhost2/link6">text
                        <a href="  link7  ">text</a>
                        <a href="">text</a>
     </a>
<noindex>
<a href="/link8">text</a>
</noindex>
<a href="/wrong%9">text</a>
</div>
</body></html>`)

		hostIDValid := sql.NullInt64{Int64: 1, Valid: true}
		hostIDInvalid := sql.NullInt64{Valid: false}

		expectedURLs := make(map[string]sql.NullInt64)
		expectedURLs["http://testhost1/link1"] = hostIDValid
		expectedURLs["http://testhost1/test/link2"] = hostIDValid
		expectedURLs["http://testhost1/test/link3/link3"] = hostIDValid
		expectedURLs["http://testhost1/link4"] = hostIDValid
		expectedURLs["http://testhost1/test/link5"] = hostIDValid
		expectedURLs["http://testhost2/link6"] = hostIDInvalid
		expectedURLs["http://testhost1/test/link7"] = hostIDValid
		So(meta.URLs, ShouldResemble, expectedURLs)

		expectedWrongURLs := make(map[string]string)
		expectedWrongURLs["/wrong%9"] = `parse /wrong%9: invalid URL escape "%9"`
		So(meta.wrongURLs, ShouldResemble, expectedWrongURLs)
	})

	Convey("TestFrameURLs", t, func() {
		meta := helperRunDataExtrator(`<html><head></head>
  <frameset><frame src="link1">
    <frameset>
      <frame src="/link2">
    </frameset>
  </frameset>
</html>`)

		hostIDValid := sql.NullInt64{Int64: 1, Valid: true}

		expectedURLs := make(map[string]sql.NullInt64)
		expectedURLs["http://testhost1/test/link1"] = hostIDValid
		expectedURLs["http://testhost1/link2"] = hostIDValid
		So(meta.URLs, ShouldResemble, expectedURLs)

		So(meta.wrongURLs, ShouldBeEmpty)
	})

	Convey("TestIFrameURLs", t, func() {
		meta := helperRunDataExtrator(`<html><head></head>
<body>
 <iframe src="link1">
 </iframe>
</body>
</html>`)

		hostIDValid := sql.NullInt64{Int64: 1, Valid: true}

		expectedURLs := make(map[string]sql.NullInt64)
		expectedURLs["http://testhost1/test/link1"] = hostIDValid
		So(meta.URLs, ShouldResemble, expectedURLs)

		So(meta.wrongURLs, ShouldBeEmpty)
	})
}

// TestParseTitle ...
func TestParseTitle(t *testing.T) {
	Convey("Title as url", t, func() {
		meta := helperRunDataExtrator(`<html><head></head><body><div></div></body></html>`)
		So(meta.GetTitle(), ShouldEqual, "http://testhost1/test/")
	})

	Convey("Title from tag title", t, func() {
		meta := helperRunDataExtrator(`<html><head><title>Title text</title></head><body><div></div></body></html>`)
		So(meta.GetTitle(), ShouldEqual, "Title text")
	})

	Convey("Title from tag meta", t, func() {
		meta := helperRunDataExtrator(`<html><head>
<meta name="title" content="Title text">
</head><body><div></div></body></html>`)
		So(meta.GetTitle(), ShouldEqual, "Title text")
	})

	Convey("Title from tags meta and title", t, func() {
		meta := helperRunDataExtrator(`<html><head>
<meta name="title" content="Title text 1">
<title>Title text 2</title>
</head><body><div></div></body></html>`)
		So(meta.GetTitle(), ShouldEqual, "Title text 2")
	})

	Convey("Long title", t, func() {
		meta := helperRunDataExtrator(`<html><head>
<title>!0000000000111111111122222222223333333333444444444455555555556666666666777777777788888888889999999999</title>
</head><body><div></div></body></html>`)
		So(meta.GetTitle(), ShouldEqual, "!000000000011111111112222222222333333333344444444445555555555666666666677777777778888888888999999...")
	})
}

// TestMeta ...
func TestMeta(t *testing.T) {
	Convey("NoIndex robots", t, func() {
		meta := helperRunDataExtrator(`<html><head>
<meta name="robots" content="noIndex">
</head><body><div></div></body></html>`)
		So(meta.MetaTagIndex, ShouldEqual, false)
	})

	Convey("NoIndex googlebot", t, func() {
		meta := helperRunDataExtrator(`<html><head>
<meta name="googlebot" content="noindex">
</head><body><div></div></body></html>`)
		So(meta.MetaTagIndex, ShouldEqual, false)
	})

	Convey("NoIndex yandex", t, func() {
		meta := helperRunDataExtrator(`<html><head>
<meta name="yandex" content="Noindex">
</head><body><div></div></body></html>`)
		So(meta.MetaTagIndex, ShouldEqual, true)
	})

	Convey("Index robots", t, func() {
		meta := helperRunDataExtrator(`<html><head>
<meta name="robots" content="index">
</head><body><div></div></body></html>`)
		So(meta.MetaTagIndex, ShouldEqual, true)
	})

	Convey("Nofollow robots", t, func() {
		meta := helperRunDataExtrator(`<html><head>
<meta name="robots" content="Nofollow">
</head><body>
<a href="link1"></a>
</body></html>`)
		So(meta.MetaTagIndex, ShouldEqual, true)
		So(meta.URLs, ShouldBeEmpty)
	})

	Convey("Index and Nofollow robots", t, func() {
		meta := helperRunDataExtrator(`<html><head>
<meta name="robots" content="index, Nofollow">
</head><body>
<a href="link1"></a>
</body></html>`)
		So(meta.MetaTagIndex, ShouldEqual, true)
		So(meta.URLs, ShouldBeEmpty)
	})

	Convey("None robots", t, func() {
		meta := helperRunDataExtrator(`<html><head>
<meta name="robots" content="none">
</head><body>
<a href="link1"></a>
</body></html>`)
		So(meta.MetaTagIndex, ShouldEqual, false)
		So(meta.URLs, ShouldBeEmpty)
	})
}

// TestErrorDataExtrator ...
func TestErrorDataExtrator(t *testing.T) {
	Convey("Test error node type", t, func() {
		node, err := html.Parse(bytes.NewReader([]byte(`<html><head></head><body></body></html>`)))
		So(err, ShouldBeNil)

		node.FirstChild.Type = html.ErrorNode
		hostMng := &hostsManager{}
		_, err = RunDataExtrator(hostMng, node, "http://testhost1/test/")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrUnexpectedNodeType)
	})

	Convey("Test error URL", t, func() {
		node, err := html.Parse(bytes.NewReader([]byte(`<html><head></head><body></body></html>`)))
		So(err, ShouldBeNil)

		hostMng := &hostsManager{}
		_, err = RunDataExtrator(hostMng, node, "%1")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrParseBaseURL)
	})
}
