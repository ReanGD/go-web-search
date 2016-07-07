package crawler

// status: ok
import (
	"net/url"
	"strings"

	"github.com/ReanGD/go-web-search/werrors"
	"github.com/uber-go/zap"
)

// HTMLMetadata extracted meta data from HTML
type HTMLMetadata struct {
	// [URL]hostname
	URLs         map[string]string
	MetaTagIndex bool
	title        string
	// [URL]error
	wrongURLs map[string]string
	baseURL   *url.URL
}

// NewHTMLMetadata - create new HTMLMetadata struct
func NewHTMLMetadata(urlStr string) (*HTMLMetadata, error) {
	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, werrors.NewFields(ErrParseBaseURL,
			zap.String("details", err.Error()),
			zap.String("parsed_url", urlStr))
	}

	return &HTMLMetadata{
		URLs:         make(map[string]string),
		wrongURLs:    make(map[string]string),
		title:        "",
		MetaTagIndex: true,
		baseURL:      baseURL,
	}, nil
}

// SetTitle - set title
func (h *HTMLMetadata) SetTitle(title string, rewrite bool) {
	if title != "" && (h.title == "" || rewrite) {
		h.title = title
	}
}

// GetTitle - get title
func (h *HTMLMetadata) GetTitle() string {
	if h.title == "" {
		return h.baseURL.String()
	}

	runeTitle := []rune(h.title)
	if len(runeTitle) > 100 {
		return string(runeTitle[0:97]) + "..."
	}

	return h.title
}

// AddURL - add not parsed URL
func (h *HTMLMetadata) AddURL(link string) {
	if link == "" {
		return
	}
	relative, err := url.Parse(strings.TrimSpace(link))
	if err != nil {
		h.wrongURLs[link] = err.Error()
		return
	}

	parsed := h.baseURL.ResolveReference(relative)
	urlStr := NormalizeURL(parsed)
	parsed, _ = url.Parse(urlStr)

	if (parsed.Scheme == "http" || parsed.Scheme == "https") && urlStr != h.baseURL.String() {
		h.URLs[urlStr] = NormalizeHostName(parsed.Host)
	}
}

// WrongURLsToLog - write to log add wrong URLs
func (h *HTMLMetadata) WrongURLsToLog(logger zap.Logger) {
	for url, error := range h.wrongURLs {
		logger.Warn("Error parse URL",
			zap.String("err_url", url),
			zap.String("details", error),
		)
	}
}
