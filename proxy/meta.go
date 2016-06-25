package proxy

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

// InMeta - inner data for Meta
type InMeta struct {
	content         *InContent
	redirectReferer *InMeta
	redirectCnt     int
	origin          sql.NullInt64
	hostName        string
	urlStr          string
	state           State
	statusCode      sql.NullInt64
}

// NewMeta - create Meta
func NewMeta(hostName string, urlStr string, referer *InMeta) *InMeta {
	it := referer
	for it != nil {
		it.redirectCnt++
		it = it.redirectReferer
	}

	return &InMeta{
		content:         nil,
		redirectReferer: referer,
		redirectCnt:     0,
		origin:          sql.NullInt64{Valid: false},
		hostName:        hostName,
		urlStr:          urlStr,
		state:           StateSuccess,
		statusCode:      sql.NullInt64{Valid: false}}
}

// SetState - set new state
func (in *InMeta) SetState(state State) {
	if state != StateSuccess {
		in.content = nil
	}
	in.state = state
}

// SetStatusCode - set new status code
func (in *InMeta) SetStatusCode(statusCode int) {
	in.statusCode = sql.NullInt64{Int64: int64(statusCode), Valid: true}
}

// SetContent - set new content
func (in *InMeta) SetContent(content *InContent) {
	in.content = content
}

// SetOrigin - set new content
func (in *InMeta) SetOrigin(origin sql.NullInt64) {
	in.origin = origin
	if origin.Valid {
		in.content = nil
		in.state = StateDublicate
	}
}

// GetURL - get current URL
func (in *InMeta) GetURL() string {
	return in.urlStr
}

// GetHostName - get field hostName
func (in *InMeta) GetHostName() string {
	return in.hostName
}

// GetContent - get field content
func (in *InMeta) GetContent() *InContent {
	return in.content
}

// GetState - get field state
func (in *InMeta) GetState() State {
	return in.state
}

// GetHash - get content hash
func (in *InMeta) GetHash() string {
	if in.content == nil || in.state != StateSuccess {
		return ""
	}

	return in.content.hash
}

// GetReferer - get field redirectReferer
func (in *InMeta) GetReferer() *InMeta {
	return in.redirectReferer
}

// NeedWaitAfterRequest - Check is need wait after request by state
func (in *InMeta) NeedWaitAfterRequest() bool {
	return in.state != StateDisabledByRobotsTxt
}

// GetMeta - get field meta converted for Db
func (in *InMeta) GetMeta(urlID int64) *Meta {
	return &Meta{
		URL:         urlID,
		State:       in.state,
		Origin:      in.origin,
		RedirectCnt: in.redirectCnt,
		StatusCode:  in.statusCode}
}
