package content

import (
	"database/sql"
	"time"
)

// State - current state of page
type State uint8

const (
	//StateSuccess - page without error
	StateSuccess State = 0
	//StateDisabledByRobotsTxt - URL disabled in robots.txt
	StateDisabledByRobotsTxt = 1
	//StateConnectError - can not connect to server
	StateConnectError = 2
	//StateErrorStatusCode - status code != 200
	StateErrorStatusCode = 3
	//StateUnsupportedFormat - MIME type != text/html
	StateUnsupportedFormat = 4
	//StateAnswerError - can not read body
	StateAnswerError = 5
	//StateParseError - can not parse body or body not html (body not save)
	StateParseError = 6
	//StateEncodingError - can not find or apply encoding (body not save)
	StateEncodingError = 7
	//StateDublicate - dublicate see "Origin" field for origin URL
	StateDublicate = 8
	//StateExternal - after redirect - host is external (body not save)
	StateExternal = 9
	//StateNoFollow - found meta tag nofollow (body not save)
	StateNoFollow = 10
)

// Host - host information
type Host struct {
	ID               int64     `gorm:"primary_key;not null"`
	Name             string    `gorm:"size:255;unique_index;not null"`
	Timestamp        time.Time `gorm:"not null"`
	RobotsStatusCode int       `gorm:"not null"`
	RobotsData       []byte
}

// Content - store page content
// Hash - hash of uncompressed content
type Content struct {
	URL   int64      `gorm:"type:integer REFERENCES url(id);unique_index;not null"`
	Hash  string     `gorm:"size:16;not null"`
	Body  Compressed `gorm:"not null"`
	Title string     `gorm:"size:100;not null"`
}

// Meta - meta information about processed URL
// Origin - link to origin document (for State == CtStateDublicate)
type Meta struct {
	URL             int64         `gorm:"type:integer REFERENCES url(id);unique_index;not null"`
	State           State         `gorm:"not null"`
	Origin          sql.NullInt64 `gorm:"type:integer REFERENCES url(id)"`
	RedirectReferer *Meta         `sql:"-"`
	HostName        string        `sql:"-"`
	URLForResolve   string        `sql:"-"`
	Content         *Content      `sql:"-"`
	RedirectCnt     int
	StatusCode      sql.NullInt64
}

// Link - links between pages
type Link struct {
	Master int64 `gorm:"type:integer REFERENCES url(id);index;not null"`
	Slave  int64 `gorm:"type:integer REFERENCES url(id);index;not null"`
}

// URL - struct for save all URLs in db
type URL struct {
	ID     int64         `gorm:"primary_key;not null"`
	URL    string        `gorm:"size:2048;not null;unique_index"`
	HostID sql.NullInt64 `gorm:"type:integer REFERENCES host(id);index"`
	Loaded bool          `gorm:"not null;index"`
}

// IsValidHash - Check is valid hash by state
func (m *Meta) IsValidHash() bool {
	return m.State == StateSuccess
}

// NeedWaitAfterRequest - Check is need wait after request by state
func (m *Meta) NeedWaitAfterRequest() bool {
	return m.State != StateDisabledByRobotsTxt
}
