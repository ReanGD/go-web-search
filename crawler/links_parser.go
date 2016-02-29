package crawler

import (
	"log"
	"net/url"
	"strings"
)

func processLinks(baseURL string, links *PageLinks, hostFilter string) (map[string]uint32, error) {
	result := make(map[string]uint32)

	base, err := url.Parse(baseURL)
	if err != nil {
		log.Printf("ERROR: Parse URL message: %s", err)
		return result, err
	}

	for it := links.LinkList.Front(); it != nil; it = it.Next() {
		relative, err := url.Parse(strings.TrimSpace(it.Value.(string)))
		if err != nil {
			log.Printf("ERROR: Parse URL on page %s, message: %s", baseURL, err)
			continue
		}
		link := base.ResolveReference(relative)
		if (link.Scheme == "http" || link.Scheme == "https") && link.Host == hostFilter {
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
