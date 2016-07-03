package crawler

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ReanGD/go-web-search/database"
	"github.com/temoto/robotstxt-go"
)

type robotTxt struct {
	Group *robotstxt.Group
}

// FromHost - init by db element content.Host
func (r *robotTxt) FromHost(host *database.Host) error {
	robot, err := robotstxt.FromStatusAndBytes(host.RobotsStatusCode, host.RobotsData)
	if err != nil {
		return err
	}
	r.Group = robot.FindGroup("Googlebot")

	return nil
}

// FromHostName - init by hostName
// hostName - normalized host name
func (r *robotTxt) FromHostName(hostName string) (*database.Host, error) {
	robotsURL := NormalizeURL(&url.URL{Scheme: "http", Host: hostName, Path: "robots.txt"})
	response, err := http.Get(robotsURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	host := &database.Host{
		Name:             hostName,
		RobotsStatusCode: response.StatusCode,
		RobotsData:       body}

	return host, r.FromHost(host)
}

func (r *robotTxt) Test(u *url.URL) bool {
	copyURL := *u
	copyURL.Scheme = ""
	copyURL.Host = ""
	return r.Group.Test(copyURL.String())
}
