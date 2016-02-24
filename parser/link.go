package parser

import (
	"container/list"
	"net/url"
	"strings"
)

// ParseLinks - normalize list of links
func ParseLinks(baseURL string, linkList list.List) (map[string]bool, error) {
	result := make(map[string]bool)

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
			result[link.String()] = true
		}
	}

	return result, nil
}
