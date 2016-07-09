package crawler

import (
	"database/sql"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ReanGD/go-web-search/proxy"
	"github.com/ReanGD/go-web-search/werrors"
	"github.com/temoto/robotstxt-go"
	"github.com/uber-go/zap"
)

type hostsManager struct {
	robotsTxt map[int64]*robotstxt.Group
	hosts     map[string]int64
}

// ResolveHost - find host id
func (m *hostsManager) ResolveHost(hostName string) sql.NullInt64 {
	hostID, ok := m.hosts[NormalizeHostName(hostName)]
	return sql.NullInt64{Int64: hostID, Valid: ok}
}

// CheckURL - check URL by robots.txt
func (m *hostsManager) CheckURL(u *url.URL) (sql.NullInt64, bool) {
	hostID := m.ResolveHost(u.Host)
	if !hostID.Valid {
		return hostID, false
	}

	copyURL := *u
	copyURL.Scheme = ""
	copyURL.Host = ""
	return hostID, m.robotsTxt[hostID.Int64].Test(copyURL.String())
}

// GetHosts - get list of hosts
func (m *hostsManager) GetHosts() []string {
	i := 0
	result := make([]string, len(m.hosts))
	for host := range m.hosts {
		result[i] = host
		i++
	}

	return result
}

func (m *hostsManager) initByDb(db proxy.DbHost) error {
	for id, host := range db.GetHosts() {
		hostName := host.GetName()
		robot, err := robotstxt.FromStatusAndBytes(host.GetRobotsTxt())
		if err != nil {
			return werrors.NewFields(ErrCreateRobotsTxtFromDb,
				zap.String("host", hostName),
				zap.String("details", err.Error()))
		}
		m.hosts[hostName] = id
		m.robotsTxt[id] = robot.FindGroup("Googlebot")
	}

	return nil
}

func (m *hostsManager) resolveURL(hostName string) (string, error) {
	hostURL := NormalizeURL(&url.URL{Scheme: "http", Host: hostName})
	response, err := http.Get(hostURL)
	if err == nil {
		err = response.Body.Close()
		if response.StatusCode != 200 {
			return "", werrors.NewFields(ErrResolveBaseURL,
				zap.Int("status_code", response.StatusCode),
				zap.String("url", hostURL))

		}
	}
	if err != nil {
		return "", werrors.NewFields(ErrGetRequest,
			zap.String("details", err.Error()),
			zap.String("url", hostURL))
	}

	return response.Request.URL.String(), nil
}

func (m *hostsManager) readRobotTxt(hostName string) (int, []byte, error) {
	var body []byte
	robotsURL := NormalizeURL(&url.URL{Scheme: "http", Host: hostName, Path: "robots.txt"})
	response, err := http.Get(robotsURL)
	if err == nil {
		body, err = ioutil.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if err == nil {
			err = closeErr
		}
	}

	if err != nil {
		return 0, body, werrors.NewFields(ErrGetRequest,
			zap.String("details", err.Error()),
			zap.String("url", robotsURL))
	}

	return response.StatusCode, body, nil
}

func (m *hostsManager) initByHostName(db proxy.DbHost, hostName string) error {
	baseURL, err := m.resolveURL(hostName)
	if err != nil {
		return err
	}

	statusCode, body, err := m.readRobotTxt(hostName)
	if err != nil {
		return err
	}

	robot, err := robotstxt.FromStatusAndBytes(statusCode, body)
	if err != nil {
		return werrors.NewDetails(ErrCreateRobotsTxtFromURL, err)
	}

	host := proxy.NewHost(hostName, statusCode, body)
	hostID, err := db.AddHost(host, baseURL)

	m.hosts[hostName] = hostID
	m.robotsTxt[hostID] = robot.FindGroup("Googlebot")

	return nil
}

// Init - init host manager
func (m *hostsManager) Init(db proxy.DbHost, baseHosts []string) error {
	m.robotsTxt = make(map[int64]*robotstxt.Group, len(baseHosts))
	m.hosts = make(map[string]int64, len(baseHosts))

	err := m.initByDb(db)
	if err != nil {
		return err
	}

	for _, hostNameRaw := range baseHosts {
		hostName := NormalizeHostName(hostNameRaw)
		_, exists := m.hosts[hostName]
		if !exists {
			err = m.initByHostName(db, hostName)
			if err != nil {
				return werrors.AddFields(err, zap.String("host", hostName))
			}
		}
	}

	return nil
}
