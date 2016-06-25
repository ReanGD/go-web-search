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
	// ErrReadGZipResponse - Can't read response body as gzip archive
	ErrReadGZipResponse = "Read response body as a gzip archive error"
	// ErrReadResponse - Can't read response body archive
	ErrReadResponse = "Read response body error"
	// ErrUnknownContentEncoding - Unknown http header Content-Encoding
	ErrUnknownContentEncoding = "Unknown http header \"Content-Encoding\""
	// WarnPageNotIndexed - Page not indexed
	WarnPageNotIndexed = "Page not indexed (meta tag noindex)"
)
