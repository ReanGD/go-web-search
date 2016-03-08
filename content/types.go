package content

import "time"

// State - current state of page
type State uint8

const (
	//StateSuccess - page without error
	StateSuccess State = iota
	//StateErrorURLFormat - can not parse URL
	StateErrorURLFormat = iota
	//StateDisabledByRobotsTxt - URL disabled in robots.txt
	StateDisabledByRobotsTxt = iota
	//StateConnectError - can not connect to server
	StateConnectError = iota
	//StateErrorStatusCode - status code != 200
	StateErrorStatusCode = iota
	//StateUnsupportedFormat - MIME type != text/html
	StateUnsupportedFormat = iota
	//StateDublicate - dublicate see "Origin" field for origin URL
	StateDublicate = iota
)

// Host - host information
type Host struct {
	ID               string    `gorm:"size:255;primary_key;not null"`
	RobotsStatusCode int       `gorm:"not null"`
	Timestamp        time.Time `gorm:"not null"`
	RobotsData       []byte
}

// Content - store page content
// Hash - hash of uncompressed content
type Content struct {
	ID   uint64     `gorm:"primary_key;not null"`
	Hash string     `gorm:"size:16;not null;index"`
	Data Compressed `gorm:"not null"`
}

// Meta - meta information about processed URL
// Parent - first parent URL
// Origin - link to origin document (for State == CtStateDublicate)
type Meta struct {
	ID         uint64    `gorm:"primary_key;not null"`
	URL        string    `gorm:"size:2048;not null;unique_index"`
	Parent     uint64    `gorm:"type:integer REFERENCES meta(id);not null"`
	Origin     uint64    `gorm:"type:integer REFERENCES meta(id)"`
	State      State     `gorm:"not null;index"`
	MIME       string    `gorm:"size:100"`
	Timestamp  time.Time `gorm:"not null;index"`
	ContentID  uint64    `gorm:"type:integer REFERENCES content(id)"`
	Content    Content
	StatusCode int32
}

// URL - struct for save all URLs in db
type URL struct {
	ID     string `gorm:"size:2048;primary_key;not null"`
	Loaded bool   `gorm:"not null;index"`
}
