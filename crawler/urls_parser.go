package crawler

import (
	"log"
	"net/url"
	"strings"
)

func processURLs(baseURL string, urls *pageURLs, baseHosts map[string]bool) (map[string]bool, error) {
	result := make(map[string]bool)

	base, err := url.Parse(baseURL)
	if err != nil {
		log.Printf("ERROR: Parse URL message: %s", err)
		return result, err
	}

	for it := urls.LinkList.Front(); it != nil; it = it.Next() {
		relative, err := url.Parse(strings.TrimSpace(it.Value.(string)))
		if err != nil {
			log.Printf("ERROR: Parse URL on page %s, message: %s", baseURL, err)
			continue
		}
		url := base.ResolveReference(relative)
		url.Fragment = ""
		urlStr := url.String()
		isBaseHost, _ := baseHosts[url.Host]
		if (url.Scheme == "http" || url.Scheme == "https") && isBaseHost && urlStr != baseURL {
			result[urlStr] = true
		}
	}

	return result, nil
}
