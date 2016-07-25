package proxy

import "github.com/ReanGD/go-web-search/database"

// Host - proxy struct for database.Host
type Host struct {
	name             string
	robotsStatusCode int
	robotsData       []byte
}

// DbHost - database interface for work with host
type DbHost interface {
	// GetHosts - return map[id]Host
	GetHosts() (map[int64]*Host, error)
	// AddHost - baseURL: init url for host
	AddHost(host *Host, baseURL string) (int64, error)
}

// NewHost - create Host
func NewHost(name string, robotsStatusCode int, robotsData []byte) *Host {
	return &Host{
		name:             name,
		robotsStatusCode: robotsStatusCode,
		robotsData:       robotsData}
}

// GetName - get name
func (h *Host) GetName() string {
	return h.name
}

// GetRobotsTxt - get data for init robots.txt
func (h *Host) GetRobotsTxt() (int, []byte) {
	return h.robotsStatusCode, h.robotsData
}

// GetTable - get field host converted for Db
func (h *Host) GetTable() *database.Host {
	return &database.Host{
		Name:             h.name,
		RobotsStatusCode: h.robotsStatusCode,
		RobotsData:       h.robotsData}
}
