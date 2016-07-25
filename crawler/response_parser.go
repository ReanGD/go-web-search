package crawler

import (
	"bytes"
	"database/sql"
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
	hostMng        *hostsManager
	meta           *proxy.Meta
	URLs           map[string]sql.NullInt64
	BodyDurationMs int64
}

// newResponseParser - create responseParser struct
func newResponseParser(logger zap.Logger, hostMng *hostsManager, meta *proxy.Meta) *responseParser {
	return &responseParser{
		logger:         logger,
		hostMng:        hostMng,
		meta:           meta,
		URLs:           make(map[string]sql.NullInt64),
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

func (r *responseParser) processBody(body []byte, contentType string) (database.State, error) {
	if !isHTML(body) {
		return database.StateParseError, werrors.New(ErrBodyNotHTML)
	}

	bodyReader, err := bodyToUTF8(body, contentType)
	if err != nil {
		return database.StateEncodingError, err
	}

	node, err := html.Parse(bodyReader)
	if err != nil {
		return database.StateParseError, werrors.NewDetails(ErrHTMLParse, err)
	}

	parser, err := RunDataExtrator(r.hostMng, node, r.meta.GetURL())
	if err != nil {
		return database.StateParseError, err
	}
	if !parser.MetaTagIndex {
		return database.StateNoFollow, werrors.NewLevel(zap.InfoLevel, WarnPageNotIndexed)
	}
	parser.WrongURLsToLog(r.logger)

	var buf bytes.Buffer
	err = bodyMinification(node, &buf)
	if err != nil {
		return database.StateParseError, err
	}

	r.URLs = parser.URLs
	r.meta.SetContent(proxy.NewContent(buf.Bytes(), parser.GetTitle()))

	r.logger.Debug(DbgBodySize, zap.Int("size", buf.Len()))
	return database.StateSuccess, nil
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

	state, err := r.processBody(body, contentType)
	if err != nil {
		r.meta.SetState(state)
		return err
	}

	r.meta.SetState(database.StateSuccess)
	return nil
}
