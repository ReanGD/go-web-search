package crawler

import (
	"bytes"
	"net/url"
	"testing"

	"golang.org/x/net/html"

	. "github.com/smartystreets/goconvey/convey"
)

func helperRunDataExtrator(htmlStr string) *HTMLMetadata {
	baseURL, _ := url.Parse("http://testhost1/test/")
	node, err := html.Parse(bytes.NewReader([]byte(htmlStr)))
	So(err, ShouldBeNil)

	meta, err := RunDataExtrator(node, baseURL)
	So(err, ShouldBeNil)

	return meta
}

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

		expectedURLs := make(map[string]string)
		expectedURLs["http://testhost1/link1"] = "testhost1"
		expectedURLs["http://testhost1/test/link2"] = "testhost1"
		expectedURLs["http://testhost1/test/link3/link3"] = "testhost1"
		expectedURLs["http://testhost1/link4"] = "testhost1"
		expectedURLs["http://testhost1/test/link5"] = "testhost1"
		expectedURLs["http://testhost2/link6"] = "testhost2"
		expectedURLs["http://testhost1/test/link7"] = "testhost1"
		So(meta.URLs, ShouldResemble, expectedURLs)

		expectedWrongURLs := make(map[string]string)
		expectedWrongURLs["/wrong%9"] = `parse /wrong%9: invalid URL escape "%9"`
		So(meta.WrongURLs, ShouldResemble, expectedWrongURLs)
	})

	Convey("TestFrameURLs", t, func() {
		meta := helperRunDataExtrator(`<html><head></head>
  <frameset><frame src="link1">
    <frameset>
      <frame src="/link2">
    </frameset>
  </frameset>
</html>`)

		expectedURLs := make(map[string]string)
		expectedURLs["http://testhost1/test/link1"] = "testhost1"
		expectedURLs["http://testhost1/link2"] = "testhost1"
		So(meta.URLs, ShouldResemble, expectedURLs)

		So(meta.WrongURLs, ShouldBeEmpty)
	})

	Convey("TestIFrameURLs", t, func() {
		meta := helperRunDataExtrator(`<html><head></head>
<body>
 <iframe src="link1">
 </iframe>
</body>
</html>`)

		expectedURLs := make(map[string]string)
		expectedURLs["http://testhost1/test/link1"] = "testhost1"
		So(meta.URLs, ShouldResemble, expectedURLs)

		So(meta.WrongURLs, ShouldBeEmpty)
	})
}

func TestTitle(t *testing.T) {
	Convey("Title as url", t, func() {
		meta := helperRunDataExtrator(`<html><head></head><body><div></div></body></html>`)
		So(meta.Title, ShouldEqual, "http://testhost1/test/")
	})

	Convey("Title from tag title", t, func() {
		meta := helperRunDataExtrator(`<html><head><title>Title text</title></head><body><div></div></body></html>`)
		So(meta.Title, ShouldEqual, "Title text")
	})

	Convey("Title from tag meta", t, func() {
		meta := helperRunDataExtrator(`<html><head>
<meta name="title" content="Title text">
</head><body><div></div></body></html>`)
		So(meta.Title, ShouldEqual, "Title text")
	})

	Convey("Title from tags meta and title", t, func() {
		meta := helperRunDataExtrator(`<html><head>
<meta name="title" content="Title text 1">
<title>Title text 2</title>
</head><body><div></div></body></html>`)
		So(meta.Title, ShouldEqual, "Title text 2")
	})

	Convey("Long title", t, func() {
		meta := helperRunDataExtrator(`<html><head>
<title>!0000000000111111111122222222223333333333444444444455555555556666666666777777777788888888889999999999</title>
</head><body><div></div></body></html>`)
		So(meta.Title, ShouldEqual, "!000000000011111111112222222222333333333344444444445555555555666666666677777777778888888888999999...")
	})
}

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

func TestErrorDataExtrator(t *testing.T) {
	Convey("Test error node type", t, func() {
		baseURL, err := url.Parse("http://testhost1/test/")
		So(err, ShouldBeNil)

		node, err := html.Parse(bytes.NewReader([]byte(`<html><head></head><body></body></html>`)))
		So(err, ShouldBeNil)

		node.FirstChild.Type = html.ErrorNode
		_, err = RunDataExtrator(node, baseURL)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, ErrUnexpectedNodeType)
	})
}
