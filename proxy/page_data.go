package proxy

// PageData - full info from page
type PageData struct {
	meta *Meta
	// map[URL]HostName
	urls      map[string]string
	parentURL int64
}

// NewPageData - create PageData
func NewPageData(meta *Meta, urls map[string]string) *PageData {
	return &PageData{
		meta: meta,
		urls: urls}
}

// SetParentURL - set field parentURL
func (in *PageData) SetParentURL(parentURL int64) {
	in.parentURL = parentURL
}

// GetMeta - get field meta
func (in *PageData) GetMeta() *Meta {
	return in.meta
}

// GetURLs - get field urls
func (in *PageData) GetURLs() map[string]string {
	return in.urls
}

// GetParentURL - get field parentURL
func (in *PageData) GetParentURL() int64 {
	return in.parentURL
}
