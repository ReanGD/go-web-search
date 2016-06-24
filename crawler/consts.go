package crawler

const (
	// ErrUnexpectedNodeType - found unexpected node type
	ErrUnexpectedNodeType = "Unexpected node type"
	// ErrUnexpectedTag - found unexpected tag
	ErrUnexpectedTag = "Unexpected tag"
	// ErrEncodingNotFound - not found encoding for content
	ErrEncodingNotFound = "Not found encoding for content"
	// ErrBodyNotHTML - not found HTML in body
	ErrBodyNotHTML = "Body not HTML"
	// ErrHTMLParse - HTML parse error
	ErrHTMLParse = "HTML parse error"
	// WarnPageNotIndexed - Page not indexed
	WarnPageNotIndexed = "Page not indexed (meta tag noindex)"
)
