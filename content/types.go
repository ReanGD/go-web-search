package content

import (
	"database/sql"
	"time"
)

// State - current state of page
type State uint8

const (
	//StateSuccess - page without error
	StateSuccess State = iota
	//StateDisabledByRobotsTxt - URL disabled in robots.txt
	StateDisabledByRobotsTxt = iota
	//StateConnectError - can not connect to server
	StateConnectError = iota
	//StateErrorStatusCode - status code != 200
	StateErrorStatusCode = iota
	//StateUnsupportedFormat - MIME type != text/html
	StateUnsupportedFormat = iota
	//StateAnswerError - can not read body
	StateAnswerError = iota
	//StateParseError - can not parse body
	StateParseError = iota
	//StateDublicate - dublicate see "Origin" field for origin URL
	StateDublicate = iota
	//StateExternal - found URL for external (body not save)
	StateExternal = iota
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
	ID   int64      `gorm:"primary_key;not null"`
	Hash string     `gorm:"size:16;not null"`
	Data Compressed `gorm:"not null"`
}

// Meta - meta information about processed URL
// Parent - first parent URL
// Origin - link to origin document (for State == CtStateDublicate)
type Meta struct {
	ID              int64          `gorm:"primary_key;not null"`
	URL             string         `gorm:"size:2048;not null;unique_index"`
	State           State          `gorm:"not null"`
	MIME            sql.NullString `gorm:"size:100"`
	Timestamp       time.Time      `gorm:"not null"`
	Parent          sql.NullInt64  `gorm:"type:integer REFERENCES meta(id)"`
	Origin          sql.NullInt64  `gorm:"type:integer REFERENCES meta(id)"`
	ContentID       sql.NullInt64  `gorm:"type:integer REFERENCES content(id)"`
	RedirectReferer *Meta          `sql:"-"`
	HostName        string         `sql:"-"`
	RedirectCnt     int
	Content         Content
	StatusCode      sql.NullInt64
}

// URL - struct for save all URLs in db
// Parent - first parent URL
type URL struct {
	ID     string        `gorm:"size:2048;primary_key;not null"`
	Parent sql.NullInt64 `gorm:"type:integer REFERENCES meta(id)"`
	HostID sql.NullInt64 `gorm:"type:integer REFERENCES host(id);index"`
	Loaded bool          `gorm:"not null;index"`
}

// IsValidHash - Check is valid hash by state
func (m *Meta) IsValidHash() bool {
	return (m.State == StateSuccess || m.State == StateParseError)
}

// NeedWaitAfterRequest - Check is need wait after request by state
func (m *Meta) NeedWaitAfterRequest() bool {
	return m.State != StateDisabledByRobotsTxt
}
