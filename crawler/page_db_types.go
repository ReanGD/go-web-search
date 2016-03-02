package crawler

// DbName - name of database pages.db
const DbName = "pages.db"

// DbBucketContents - name contents bucket in database pages.db
const DbBucketContents = "Contents"

// DbBucketWrongURLs - name bucket with info abount wrong pages in database pages.db
const DbBucketWrongURLs = "WrongURLs"

// DbBucketURLs - name links bucket in database pages.db
const DbBucketURLs = "URLs"

// DbBucketMeta - name meta bucket in database pages.db
const DbBucketMeta = "Meta"

// DbKeyMeta - name first meta bucket key in database pages.db
const DbKeyMeta = "0"

const (
	//PageTypeSuccess - page without error
	PageTypeSuccess uint8 = iota
	//PageType404 - page 404
	PageType404 = iota
)

//go:generate msgp

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
