package crawler

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/purell"
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
