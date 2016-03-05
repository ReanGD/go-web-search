package crawler

// DbName - name of database pages.db
const DbName = "pages.db"

// DbBucketHosts - bucket in database pages.db with base host
const DbBucketHosts = "Hosts"

// DbBucketContents - bucket in database pages.db with page contents
const DbBucketContents = "Contents"

// DbBucketWrongURLs - bucket in database pages.db with wrong URLs
const DbBucketWrongURLs = "WrongURLs"

// DbBucketURLs - bucket in database pages.db with URLs meta info
const DbBucketURLs = "URLs"

// DbBucketMeta - bucket in database pages.db with db settings
const DbBucketMeta = "Meta"

// DbKeyMeta - name first meta bucket key in database pages.db
const DbKeyMeta = "0"

const (
	//PageTypeNone - not loaded page
	PageTypeNone uint8 = iota
	//PageTypeSuccess - page without error
	PageTypeSuccess = iota
	//PageType404 - page 404
	PageType404 = iota
)

//go:generate msgp -tests=false
//msgp:encode ignore DbHost DbContent DbWrongURL DbURL DbMeta
//msgp:decode ignore DbHost DbContent DbWrongURL DbURL DbMeta

// DbHost - host information in page_db
type DbHost struct {
	RobotsTxt  []byte `msg:"robots"`
	StatusCode int    `msg:"status"`
}

// DbContent - page content in page_db
type DbContent struct {
	ID      uint64   `msg:"id"`
	Content []byte   `msg:"content"`
	Hash    [16]byte `msg:"hash"`
}

// DbWrongURL - info about wrong URL in page_db
// ErrorType: see PageTypeSuccess, PageType404, etc.
type DbWrongURL struct {
	ErrorType uint8 `msg:"etype"`
}

// DbURL - page url meta info in page_db
// ErrorType: see PageTypeSuccess, PageType404, etc.
type DbURL struct {
	ID        uint64 `msg:"id"`
	Count     uint32 `msg:"count"`
	ErrorType uint8  `msg:"etype"`
}

// DbMeta - meta data for page_db
type DbMeta struct {
	LastID uint64 `msg:"lastId"`
}
