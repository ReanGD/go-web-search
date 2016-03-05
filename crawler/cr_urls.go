package crawler

import (
	"container/list"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/purell"
	"github.com/temoto/robotstxt.go"
)

const safeNormalizationFlags purell.NormalizationFlags = purell.FlagLowercaseScheme |
	purell.FlagLowercaseHost |
	purell.FlagUppercaseEscapes |
	purell.FlagDecodeUnnecessaryEscapes |
	purell.FlagEncodeNecessaryEscapes |
	purell.FlagRemoveDefaultPort |
	purell.FlagRemoveEmptyQuerySeparator

const usuallySafeNormalizationFlags purell.NormalizationFlags = purell.FlagRemoveTrailingSlash |
	purell.FlagRemoveDotSegments

const unsafeNormalizationFlags purell.NormalizationFlags = purell.FlagRemoveFragment |
	purell.FlagForceHTTP |
	purell.FlagRemoveDuplicateSlashes |
	purell.FlagRemoveWWW |
	purell.FlagSortQuery

const defaultNormalizationFlags purell.NormalizationFlags = safeNormalizationFlags |
	usuallySafeNormalizationFlags |
	unsafeNormalizationFlags

// NormalizeURL - nomalize URL
func NormalizeURL(u *url.URL) string {
	return purell.NormalizeURL(u, defaultNormalizationFlags)
}

// NormalizeRawURL - normalize URL
func NormalizeRawURL(rawURL string) (string, error) {
	var result string
	u, err := url.Parse(rawURL)
	if err != nil {
		return result, err
	}
	result = NormalizeURL(u)
	return result, nil
}

// NormalizeHost - normalize host name
func NormalizeHost(host string) string {
	var result string
	if len(host) > 0 {
		result = strings.ToLower(host)
	}

	return result
}

// URLFromHost - create normalized URL from hostname
func URLFromHost(host string) string {
	return NormalizeURL(&url.URL{Scheme: "http", Host: NormalizeHost(host)})
}

type handlerURLs struct {
	Robot     *robotstxt.Group
	BaseHosts map[string]int
}

func (h *handlerURLs) handle(baseURL string, urls *list.List) (map[string]bool, error) {
	result := make(map[string]bool)

	base, err := url.Parse(baseURL)
	if err != nil {
		log.Printf("ERROR: Parse URL message: %s", err)
		return result, err
	}

	for it := urls.Front(); it != nil; it = it.Next() {
		relative, err := url.Parse(strings.TrimSpace(it.Value.(string)))
		if err != nil {
			log.Printf("ERROR: Parse URL on page %s, message: %s", baseURL, err)
			continue
		}
		parsed := base.ResolveReference(relative)
		urlStr := NormalizeURL(parsed)
		parsed, err = url.Parse(urlStr)

		_, isBaseHost := h.BaseHosts[parsed.Host]
		isHTTP := (parsed.Scheme == "http" || parsed.Scheme == "https")

		if isHTTP && isBaseHost && urlStr != baseURL {
			parsed.Scheme = ""
			parsed.Host = ""
			if h.Robot.Test(parsed.String()) {
				result[urlStr] = true
			}
		}
	}

	return result, nil
}
