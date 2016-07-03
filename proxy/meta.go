package proxy

import (
	"database/sql"

	"github.com/ReanGD/go-web-search/database"
)

// Meta - proxy struct for database.Meta
type Meta struct {
	content         *Content
	redirectReferer *Meta
	redirectCnt     int
	origin          sql.NullInt64
	hostName        string
	urlStr          string
	state           database.State
	statusCode      sql.NullInt64
}

// NewMeta - create Meta
func NewMeta(hostName string, urlStr string, referer *Meta) *Meta {
	it := referer
	for it != nil {
		it.redirectCnt++
		it = it.redirectReferer
	}

	return &Meta{
		content:         nil,
		redirectReferer: referer,
		redirectCnt:     0,
		origin:          sql.NullInt64{Valid: false},
		hostName:        hostName,
		urlStr:          urlStr,
		state:           database.StateSuccess,
		statusCode:      sql.NullInt64{Valid: false}}
}

// SetState - set new state
func (in *Meta) SetState(state database.State) {
	if state != database.StateSuccess {
		in.content = nil
	}
	in.state = state
}

// SetStatusCode - set new status code
func (in *Meta) SetStatusCode(statusCode int) {
	in.statusCode = sql.NullInt64{Int64: int64(statusCode), Valid: true}
}

// SetContent - set new content
func (in *Meta) SetContent(content *Content) {
	in.content = content
}

// SetOrigin - set new content
func (in *Meta) SetOrigin(origin sql.NullInt64) {
	in.origin = origin
	if origin.Valid {
		in.content = nil
		in.state = database.StateDublicate
	}
}

// GetURL - get current URL
func (in *Meta) GetURL() string {
	return in.urlStr
}

// GetHostName - get field hostName
func (in *Meta) GetHostName() string {
	return in.hostName
}

// GetContent - get field content
func (in *Meta) GetContent() *Content {
	return in.content
}

// GetState - get field state
func (in *Meta) GetState() database.State {
	return in.state
}

// GetHash - get content hash
func (in *Meta) GetHash() string {
	if in.content == nil || in.state != database.StateSuccess {
		return ""
	}

	return in.content.hash
}

// GetReferer - get field redirectReferer
func (in *Meta) GetReferer() *Meta {
	return in.redirectReferer
}

// NeedWaitAfterRequest - Check is need wait after request by state
func (in *Meta) NeedWaitAfterRequest() bool {
	return in.state != database.StateDisabledByRobotsTxt
}

// GetMeta - get field meta converted for Db
func (in *Meta) GetMeta(urlID int64) *database.Meta {
	return &database.Meta{
		URL:         urlID,
		State:       in.state,
		Origin:      in.origin,
		RedirectCnt: in.redirectCnt,
		StatusCode:  in.statusCode}
}
