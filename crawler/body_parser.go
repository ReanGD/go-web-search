package crawler

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/url"

	"github.com/ReanGD/go-web-search/proxy"
	"github.com/ReanGD/go-web-search/werrors"
	"github.com/uber-go/zap"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

func isHTML(content []byte) bool {
	isHTML := false
	if len(content) == 0 {
		return isHTML
	}
	if len(content) > 1024 {
		content = content[:1024]
	}

	z := html.NewTokenizer(bytes.NewReader(content))
	isFinish := false
	for !isFinish {
		switch z.Next() {
		case html.ErrorToken:
			isFinish = true
		case html.StartTagToken:
			tagName, _ := z.TagName()
			if bytes.Equal(tagName, []byte("html")) {
				isHTML = true
				isFinish = true
			}
		}
	}

	return isHTML
}

func bodyToUTF8(body []byte, contentType []string) (*transform.Reader, error) {
	contentTypeStr := ""
	if len(contentType) != 0 {
		contentTypeStr = contentType[0]
	}
	enc, _, _ := charset.DetermineEncoding(body, contentTypeStr)
	if enc == encoding.Nop {
		return nil, werrors.New(ErrEncodingNotFound)
	}

	return transform.NewReader(bytes.NewReader(body), enc.NewDecoder()), nil
}

func readBody(contentEncoding string, body io.Reader) ([]byte, error) {
	var err error
	result := []byte{}
	if contentEncoding == "gzip" {
		reader, err := gzip.NewReader(body)
		if err != nil {
			return result, werrors.NewDetails(ErrReadGZipResponse, err)
		}
		result, err = ioutil.ReadAll(reader)
		if err == nil {
			err = reader.Close()
		} else {
			_ = reader.Close()
		}
		if err != nil {
			return result, werrors.NewDetails(ErrReadGZipResponse, err)
		}
	} else if contentEncoding == "identity" {
		result, err = ioutil.ReadAll(body)
		if err != nil {
			return result, werrors.NewDetails(ErrReadResponse, err)
		}
	} else {
		return result, werrors.NewFields(ErrUnknownContentEncoding, "encoding", contentEncoding)
	}

	return result, nil
}

// ProcessBody - check and minimize request body
func ProcessBody(logger zap.Logger, body []byte, contentType []string, u *url.URL) (*HTMLMetadata, proxy.State, error) {
	state := proxy.StateSuccess

	if !isHTML(body) {
		return nil, proxy.StateParseError, werrors.New(ErrBodyNotHTML)
	}

	bodyReader, err := bodyToUTF8(body, contentType)
	if err != nil {
		return nil, proxy.StateEncodingError, err
	}

	node, err := html.Parse(bodyReader)
	if err != nil {
		return nil, proxy.StateParseError, werrors.NewDetails(ErrHTMLParse, err)
	}

	parser, err := RunDataExtrator(node, u)
	if err != nil {
		return nil, proxy.StateParseError, err
	}
	if !parser.MetaTagIndex {
		return nil, proxy.StateNoFollow, werrors.NewLevel(WarnPageNotIndexed, werrors.WarningLevel)
	}

	for url, error := range parser.WrongURLs {
		logger.Warn("Error parse URL",
			zap.String("err_url", url),
			zap.String("error", error),
		)
	}

	err = RunMinificationHTML(node)
	if err != nil {
		return nil, proxy.StateParseError, err
	}

	err = RunMinificationText(node)
	if err != nil {
		return nil, proxy.StateParseError, err
	}

	return parser, state, nil
}
