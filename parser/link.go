package parser

import (
	"container/list"
	"net/url"
	"strings"
)

// ParseLinks - normalize list of links
func ParseLinks(baseURL string, linkList list.List) (map[string]uint32, error) {
	result := make(map[string]uint32)

	base, err := url.Parse(baseURL)
	if err != nil {
		return result, err
	}

	for it := linkList.Front(); it != nil; it = it.Next() {
		relative, err := url.Parse(strings.TrimSpace(it.Value.(string)))
		if err != nil {
			return result, err
		}
		link := base.ResolveReference(relative)
		if link.Scheme == "http" || link.Scheme == "https" {
			link.Fragment = ""
			linkStr := link.String()
			if val, ok := result[linkStr]; ok {
				result[linkStr] = val + 1
			} else {
				result[linkStr] = 1
			}
		}
	}

	return result, nil
}
