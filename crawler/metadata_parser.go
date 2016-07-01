package crawler

import (
	"mime"
	"net/http"

	"github.com/ReanGD/go-web-search/werrors"
	"github.com/uber-go/zap"
)

func checkStatusCode(statusCode int) error {
	if statusCode != 200 {
		lvl := zap.ErrorLevel
		// 401 Unauthorized
		// 404 Not Found
		if statusCode == 404 || statusCode == 401 {
			lvl = zap.WarnLevel
		}
		return werrors.NewEx(lvl, ErrStatusCode, zap.Int("status_code", statusCode))
	}

	return nil
}

func checkContentType(header *http.Header) (string, error) {
	contentTypeArr, ok := (*header)["Content-Type"]
	if !ok || len(contentTypeArr) == 0 {
		return "", werrors.New(ErrNotFountContentType)
	}
	contentType := contentTypeArr[0]

	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", werrors.NewFields(ErrParseContentType,
			zap.String("detail", err.Error()),
			zap.String("content_type", contentType))
	}

	if mediatype != "text/html" {
		return "", werrors.NewEx(zap.InfoLevel, InfoUnsupportedMimeFormat,
			zap.String("content_type", contentType))
	}

	return contentType, nil
}

func getContentEncoding(header *http.Header) string {
	contentEncodingArr, ok := (*header)["Content-Encoding"]
	if ok && len(contentEncodingArr) != 0 {
		return contentEncodingArr[0]
	}

	return ""
}
