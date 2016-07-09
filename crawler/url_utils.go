package crawler

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/purell"
	"github.com/opennota/urlesc"
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

func removeUtmFromQuery(u *url.URL) {
	q := u.Query()

	if len(q) > 0 {
		buf := new(bytes.Buffer)
		for key, value := range q {
			if key != "utm_source" &&
				key != "utm_medium" &&
				key != "utm_term" &&
				key != "utm_content" &&
				key != "utm_campaign" {
				for _, v := range value {
					if buf.Len() > 0 {
						_, _ = buf.WriteRune('&')
					}
					_, _ = buf.WriteString(fmt.Sprintf("%s=%s", key, urlesc.QueryEscape(v)))
				}
			}
		}

		// Rebuild the raw query string
		u.RawQuery = buf.String()
	}
}

// NormalizeURL - nomalize URL
func NormalizeURL(u *url.URL) string {
	removeUtmFromQuery(u)
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
