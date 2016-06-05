package crawler

import (
	"net/http"
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

// + purell.FlagRemoveTrailingSlash
const usuallySafeNormalizationFlags purell.NormalizationFlags = purell.FlagRemoveDotSegments

// +purell.FlagForceHTTP + purell.FlagRemoveWWW
const unsafeNormalizationFlags purell.NormalizationFlags = purell.FlagRemoveFragment |
	purell.FlagRemoveDuplicateSlashes |
	purell.FlagSortQuery

const defaultNormalizationFlags purell.NormalizationFlags = safeNormalizationFlags |
	usuallySafeNormalizationFlags |
	unsafeNormalizationFlags

// NormalizeURL - nomalize URL
func NormalizeURL(u *url.URL) string {
	return purell.NormalizeURL(u, defaultNormalizationFlags)
}

// NormalizeHostName - normalize host name
func NormalizeHostName(hostName string) string {
	result := hostName
	if len(result) > 0 {
		result = strings.ToLower(result)
	}
	if len(result) > 0 && strings.HasPrefix(strings.ToLower(result), "www.") {
		result = result[4:]
	}

	return result
}

// GenerateURLByHostName - generate URL by hostName
func GenerateURLByHostName(hostName string) (string, error) {
	response, err := http.Get(NormalizeURL(&url.URL{Scheme: "http", Host: hostName}))
	if err != nil {
		return "", err
	}
	err = response.Body.Close()

	return response.Request.URL.String(), err
}
