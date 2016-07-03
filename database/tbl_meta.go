package database

import "database/sql"

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

// Meta - meta information about processed URL
// Origin - link to origin document (for State == CtStateDublicate)
type Meta struct {
	URL         int64         `gorm:"type:integer REFERENCES url(id);unique_index;not null"`
	State       State         `gorm:"not null"`
	Origin      sql.NullInt64 `gorm:"type:integer REFERENCES url(id)"`
	RedirectCnt int
	StatusCode  sql.NullInt64
}
