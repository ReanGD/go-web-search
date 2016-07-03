package database

// Content - store page content
// Hash - hash of uncompressed content
type Content struct {
	URL   int64      `gorm:"type:integer REFERENCES url(id);unique_index;not null"`
	Hash  string     `gorm:"size:64;not null"`
	Body  Compressed `gorm:"not null"`
	Title string     `gorm:"size:100;not null"`
}
