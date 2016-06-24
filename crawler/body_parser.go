package crawler

import (
	"bytes"
	"net/url"

	"github.com/ReanGD/go-web-search/content"
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

// ProcessBody - check and minimize request body
func ProcessBody(logger zap.Logger, body []byte, contentType []string, u *url.URL) (*HTMLMetadata, content.State, error) {
	state := content.StateSuccess

	if !isHTML(body) {
		return nil, content.StateParseError, werrors.New(ErrBodyNotHTML)
	}

	bodyReader, err := bodyToUTF8(body, contentType)
	if err != nil {
		return nil, content.StateEncodingError, err
	}

	node, err := html.Parse(bodyReader)
	if err != nil {
		return nil, content.StateParseError, werrors.NewDetails(ErrHTMLParse, err)
	}

	parser, err := RunDataExtrator(node, u)
	if err != nil {
		return parser, content.StateParseError, err
	}
	if !parser.MetaTagIndex {
		return parser, content.StateNoFollow, werrors.NewLevel(WarnPageNotIndexed, werrors.WarningLevel)
	}

	for url, error := range parser.WrongURLs {
		logger.Warn("Error parse URL",
			zap.String("err_url", url),
			zap.String("error", error),
		)
	}

	return parser, state, nil
}
