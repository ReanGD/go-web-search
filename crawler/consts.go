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
	// ErrStatusCode - Wrong status code
	ErrStatusCode = "Wrong status code"
	// ErrNotFountContentType - Not found Content-Type in headers
	ErrNotFountContentType = "Not found Content-Type in headers"
	// ErrParseContentType - Error parse Content-Type header
	ErrParseContentType = "Parse Content-Type header"
	// ErrRenderHTML - Error Render HTML
	ErrRenderHTML = "Render HTML"
	// ErrParseBaseURL - Parse base URL
	ErrParseBaseURL = "Parse base URL"
	// ErrResolveBaseURL - Resolve base URL by host name
	ErrResolveBaseURL = "Resolve base URL by hostname: StatusCode != 200"
	// ErrGetRequest - Error Get request
	ErrGetRequest = "Get request"
	// ErrReadResponseBody - Error read response body
	ErrReadResponseBody = "Read response body"
	// ErrCloseResponseBody - Close response body
	ErrCloseResponseBody = "Close response body"
	// ErrCreateRobotsTxtFromDb - Error create robots.txt from db data
	ErrCreateRobotsTxtFromDb = "Create robots.txt from db data"
	// ErrCreateRobotsTxtFromURL - Error create robots.txt from url
	ErrCreateRobotsTxtFromURL = "Create robots.txt from url"
	// WarnPageNotIndexed - Page not indexed
	WarnPageNotIndexed = "Page not indexed (meta tag noindex)"
	// InfoUnsupportedMimeFormat - Unsupported mime format
	InfoUnsupportedMimeFormat = "Unsupported mime format"
	// DbgRequestDuration - Request duration
	DbgRequestDuration = "Request duration"
	// DbgBodyProcessingDuration - Body processing duration
	DbgBodyProcessingDuration = "Body processing duration"
	// DbgBodySize - Body size after processing
	DbgBodySize = "Body size"
)
