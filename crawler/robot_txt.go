package crawler

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/ReanGD/go-web-search/content"
	"github.com/temoto/robotstxt-go"
)

type robotTxt struct {
	Group *robotstxt.Group
	Host  *content.Host
}

// FromHost - init by db element content.Host
func (r *robotTxt) FromHost(host *content.Host) error {
	robot, err := robotstxt.FromStatusAndBytes(host.RobotsStatusCode, host.RobotsData)
	if err != nil {
		return err
	}
	r.Host = host
	r.Group = robot.FindGroup("Googlebot")

	return nil
}

// FromHostName - init by hostName
// hostName - normalized host name
func (r *robotTxt) FromHostName(hostName string) error {
	robotsURL := NormalizeURL(&url.URL{Scheme: "http", Host: hostName, Path: "robots.txt"})
	response, err := http.Get(robotsURL)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	host := &content.Host{Name: hostName,
		RobotsStatusCode: response.StatusCode,
		RobotsData:       body,
		Timestamp:        time.Now()}

	return r.FromHost(host)
}

func (r *robotTxt) Test(urlStr string) (bool, error) {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false, err
	}

	parsed.Scheme = ""
	parsed.Host = ""
	return r.Group.Test(parsed.String()), nil
}
