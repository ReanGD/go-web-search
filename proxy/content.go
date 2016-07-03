package proxy

import (
	"crypto/sha512"

	"github.com/ReanGD/go-web-search/database"
)

// Content - proxy struct for database.Content
type Content struct {
	hash  string
	body  database.Compressed
	title string
}

// NewContent - create Content
func NewContent(body []byte, title string) *Content {
	hash := sha512.Sum512(body)
	return &Content{
		hash:  string(hash[:]),
		body:  database.Compressed{Data: body},
		title: title}
}

// GetContent - convert to content.Content
func (in *Content) GetContent(urlID int64) *database.Content {
	return &database.Content{
		URL:   urlID,
		Hash:  in.hash,
		Body:  in.body,
		Title: in.title}
}
