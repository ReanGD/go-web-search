package crawler

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ReanGD/go-web-search/content"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

var (
	// ErrBodyParserEncodingNotFound - not found encoding for content
	ErrBodyParserEncodingNotFound = errors.New("body_parser.bodyToUTF8: not found encoding for content")
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
		return nil, ErrBodyParserEncodingNotFound
	}

	return transform.NewReader(bytes.NewReader(body), enc.NewDecoder()), nil
}

// ProcessBody - check and minimize request body
func ProcessBody(body []byte, contentType []string) (*html.Node, content.State, error) {
	state := content.StateSuccess

	if !isHTML(body) {
		return nil, content.StateParseError, fmt.Errorf("Body not html")
	}

	bodyReader, err := bodyToUTF8(body, contentType)
	if err != nil {
		return nil, content.StateEncodingError, err
	}

	node, err := html.Parse(bodyReader)
	if err != nil {
		return nil, content.StateParseError, fmt.Errorf("Html parse error: %s", err)
	}

	return node, state, nil
}
