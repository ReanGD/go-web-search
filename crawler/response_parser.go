package crawler

import (
	"bytes"
	"mime"
	"net/http"

	"github.com/ReanGD/go-web-search/proxy"
	"github.com/ReanGD/go-web-search/werrors"
	"github.com/uber-go/zap"
	"golang.org/x/net/html"
)

type responseParser struct {
	logger zap.Logger
	meta   *proxy.InMeta
	URLs   map[string]string
}

// newResponseParser - create responseParser struct
func newResponseParser(logger zap.Logger, meta *proxy.InMeta) *responseParser {
	return &responseParser{
		logger: logger,
		meta:   meta}
}

func (r *responseParser) processStatusCode(statusCode int) error {
	r.meta.SetStatusCode(statusCode)
	if statusCode != 200 {
		r.meta.SetState(proxy.StateErrorStatusCode)
		return werrors.NewFields(ErrStatusCode, zap.Int("status_code", statusCode))
	}

	return nil
}

func (r *responseParser) processContentType(header *http.Header) (string, error) {
	contentType := ""
	contentTypeArr, ok := (*header)["Content-Type"]
	if !ok || len(contentTypeArr) == 0 {
		r.meta.SetState(proxy.StateUnsupportedFormat)
		return "", werrors.New(ErrNotFountContentType)
	}
	contentType = contentTypeArr[0]

	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		r.meta.SetState(proxy.StateUnsupportedFormat)
		return "", werrors.NewFields(ErrParseContentType,
			zap.String("detail", err.Error()),
			zap.String("content_type", contentType))
	}

	if mediatype != "text/html" {
		r.meta.SetState(proxy.StateUnsupportedFormat)
		return "", werrors.NewEx(zap.InfoLevel, InfoUnsupportedMimeFormat,
			zap.String("content_type", contentType))
	}

	return contentType, nil
}

func (r *responseParser) ProcessBody(body []byte, contentType string) error {
	if !isHTML(body) {
		r.meta.SetState(proxy.StateParseError)
		return werrors.New(ErrBodyNotHTML)
	}

	bodyReader, err := bodyToUTF8(body, contentType)
	if err != nil {
		r.meta.SetState(proxy.StateEncodingError)
		return err
	}

	node, err := html.Parse(bodyReader)
	if err != nil {
		r.meta.SetState(proxy.StateParseError)
		return werrors.NewDetails(ErrHTMLParse, err)
	}

	parser, err := RunDataExtrator(node, r.meta.GetURL())
	if err != nil {
		r.meta.SetState(proxy.StateParseError)
		return err
	}
	if !parser.MetaTagIndex {
		r.meta.SetState(proxy.StateNoFollow)
		return werrors.NewLevel(zap.WarnLevel, WarnPageNotIndexed)
	}

	for url, error := range parser.WrongURLs {
		r.logger.Warn("Error parse URL",
			zap.String("err_url", url),
			zap.String("error", error),
		)
	}

	var buf bytes.Buffer
	err = bodyMinification(node, &buf)
	if err != nil {
		r.meta.SetState(proxy.StateParseError)
		return err
	}

	r.URLs = parser.URLs
	r.meta.SetContent(proxy.NewContent(buf.Bytes(), parser.Title))

	return nil
}

func (r *responseParser) Run(response *http.Response) error {
	defer response.Body.Close()

	err := r.processStatusCode(response.StatusCode)
	if err != nil {
		return err
	}

	contentType, err := r.processContentType(&response.Header)
	if err != nil {
		return err
	}

	contentEncoding := ""
	contentEncodingArr, ok := response.Header["Content-Encoding"]
	if ok && len(contentEncodingArr) != 0 {
		contentEncoding = contentEncodingArr[0]
	}

	body, err := readBody(contentEncoding, response.Body)
	if err != nil {
		r.meta.SetState(proxy.StateAnswerError)
		return err
	}

	err = r.ProcessBody(body, contentType)
	if err != nil {
		return err
	}

	r.meta.SetState(proxy.StateSuccess)
	return nil
}
