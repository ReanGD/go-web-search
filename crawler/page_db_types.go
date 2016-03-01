package crawler

// DbName - name of database pages.db
const DbName = "pages.db"

// DbBucketContents - name contents bucket in database pages.db
const DbBucketContents = "Contents"

// DbBucketURLs - name links bucket in database pages.db
const DbBucketURLs = "URLs"

// DbBucketMeta - name meta bucket in database pages.db
const DbBucketMeta = "Meta"

// DbKeyMeta - name first meta bucket key in database pages.db
const DbKeyMeta = "0"

//go:generate msgp

// DbContent - page content in page_db
type DbContent struct {
	ID      uint64   `msg:"id"`
	Content []byte   `msg:"content"`
	Hash    [16]byte `msg:"hash"`
}

// DbURL - page url meta info in page_db
type DbURL struct {
	ID    uint64 `msg:"id"`
	Count uint32 `msg:"count"`
}

// DbMeta - meta data for page_db
type DbMeta struct {
	LastID uint64 `msg:"lastId"`
}
