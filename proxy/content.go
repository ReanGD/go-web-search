package proxy

import (
	"crypto/sha512"

	"github.com/ReanGD/go-web-search/database"
)

// Content - store page content
// Hash - hash of uncompressed content
type Content struct {
	URL   int64               `gorm:"type:integer REFERENCES url(id);unique_index;not null"`
	Hash  string              `gorm:"size:64;not null"`
	Body  database.Compressed `gorm:"not null"`
	Title string              `gorm:"size:100;not null"`
}

// InContent - inner data for Content
type InContent struct {
	hash  string
	body  database.Compressed
	title string
}

// NewContent - create Content
func NewContent(body []byte, title string) *InContent {
	hash := sha512.Sum512(body)
	return &InContent{
		hash:  string(hash[:]),
		body:  database.Compressed{Data: body},
		title: title}
}

// GetContent - convert to content.Content
func (in *InContent) GetContent(urlID int64) *Content {
	return &Content{
		URL:   urlID,
		Hash:  in.hash,
		Body:  in.body,
		Title: in.title}
}
