package crawler

import (
	"bytes"
	"net/http"
	"time"

	"github.com/ReanGD/go-web-search/database"
	"github.com/ReanGD/go-web-search/proxy"
	"github.com/ReanGD/go-web-search/werrors"
	"github.com/uber-go/zap"
	"golang.org/x/net/html"
)

type responseParser struct {
	logger         zap.Logger
	meta           *proxy.Meta
	URLs           map[string]string
	BodyDurationMs int64
}

// newResponseParser - create responseParser struct
func newResponseParser(logger zap.Logger, meta *proxy.Meta) *responseParser {
	return &responseParser{
		logger:         logger,
		meta:           meta,
		URLs:           make(map[string]string),
		BodyDurationMs: 0}
}

func (r *responseParser) processMeta(statusCode int, header *http.Header) (string, string, error) {
	r.meta.SetStatusCode(statusCode)
	err := checkStatusCode(statusCode)
	if err != nil {
		r.meta.SetState(database.StateErrorStatusCode)
		return "", "", err
	}

	contentType, err := checkContentType(header)
	if err != nil {
		r.meta.SetState(database.StateUnsupportedFormat)
		return "", "", err
	}

	return contentType, getContentEncoding(header), nil
}

func (r *responseParser) processBody(body []byte, contentType string) error {
	if !isHTML(body) {
		r.meta.SetState(database.StateParseError)
		return werrors.New(ErrBodyNotHTML)
	}

	bodyReader, err := bodyToUTF8(body, contentType)
	if err != nil {
		r.meta.SetState(database.StateEncodingError)
		return err
	}

	node, err := html.Parse(bodyReader)
	if err != nil {
		r.meta.SetState(database.StateParseError)
		return werrors.NewDetails(ErrHTMLParse, err)
	}

	parser, err := RunDataExtrator(node, r.meta.GetURL())
	if err != nil {
		r.meta.SetState(database.StateParseError)
		return err
	}
	if !parser.MetaTagIndex {
		r.meta.SetState(database.StateNoFollow)
		return werrors.NewLevel(zap.InfoLevel, WarnPageNotIndexed)
	}
	parser.WrongURLsToLog(r.logger)

	var buf bytes.Buffer
	err = bodyMinification(node, &buf)
	if err != nil {
		r.meta.SetState(database.StateParseError)
		return err
	}

	r.URLs = parser.URLs
	r.meta.SetContent(proxy.NewContent(buf.Bytes(), parser.GetTitle()))

	r.logger.Debug(DbgBodySize, zap.Int("size", buf.Len()))
	return nil
}

func (r *responseParser) timeTrack(start time.Time) {
	r.BodyDurationMs = int64(time.Since(start) / time.Millisecond)
	r.logger.Debug(DbgBodyProcessingDuration, zap.Int64("duration", r.BodyDurationMs))
}

func (r *responseParser) Run(response *http.Response) error {
	contentType, contentEncoding, err := r.processMeta(response.StatusCode, &response.Header)
	if err != nil {
		_ = response.Body.Close()
		return err
	}

	body, err := readBody(contentEncoding, response.Body)
	closeErr := response.Body.Close()
	defer r.timeTrack(time.Now())
	if err != nil {
		r.meta.SetState(database.StateAnswerError)
		return err
	}
	if closeErr != nil {
		r.meta.SetState(database.StateAnswerError)
		return werrors.NewDetails(ErrCloseResponseBody, closeErr)
	}

	err = r.processBody(body, contentType)
	if err != nil {
		return err
	}

	r.meta.SetState(database.StateSuccess)
	return nil
}
