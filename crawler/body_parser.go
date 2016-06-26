package crawler

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"

	"github.com/ReanGD/go-web-search/werrors"
	"github.com/uber-go/zap"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

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
		return result, werrors.NewFields(ErrUnknownContentEncoding, zap.String("encoding", contentEncoding))
	}

	return result, nil
}

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

func bodyToUTF8(body []byte, contentType string) (*transform.Reader, error) {
	enc, _, _ := charset.DetermineEncoding(body, contentType)
	if enc == encoding.Nop {
		return nil, werrors.New(ErrEncodingNotFound)
	}

	return transform.NewReader(bytes.NewReader(body), enc.NewDecoder()), nil
}

func bodyMinification(node *html.Node, buf io.Writer) error {
	htmlMinification := minificationHTML{}
	err := htmlMinification.Run(node)

	if err == nil {
		textMinification := minificationText{}
		err = textMinification.Run(node)
	}

	if err == nil {
		wbuf := bufio.NewWriter(buf)
		err = html.Render(wbuf, node)
		if err == nil {
			err = wbuf.Flush()
		}
		if err != nil {
			return werrors.NewDetails(ErrRenderHTML, err)
		}
	}

	return err
}
