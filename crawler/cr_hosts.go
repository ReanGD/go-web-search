package crawler

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/temoto/robotstxt.go"
)

func getRobotsTxt(host string) (*DbHost, error) {
	robotsURL := NormalizeURL(&url.URL{Scheme: "http", Host: host, Path: "robots.txt"})
	response, err := http.Get(robotsURL)

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	result := &DbHost{RobotsTxt: body, StatusCode: response.StatusCode}
	return result, nil
}

func getHostsRobotsTxt(db *DB, baseHosts map[string]int) (map[string]*robotstxt.Group, error) {
	var result map[string]*robotstxt.Group

	newHosts, err := db.readNotLoadedHosts(baseHosts)
	if err != nil {
		return result, err
	}

	if len(newHosts) != 0 {
		loadHosts := make(map[string]*DbHost, len(newHosts))
		for _, host := range newHosts {
			dbHost, err := getRobotsTxt(host)
			if err != nil {
				return result, err
			}
			loadHosts[host] = dbHost
		}

		err = db.writeHosts(loadHosts)
		if err != nil {
			return result, err
		}
	}

	hostsData, err := db.readHosts(baseHosts)
	if err != nil {
		return result, err
	}

	result = make(map[string]*robotstxt.Group, len(baseHosts))
	for host, hostData := range hostsData {
		robot, err := robotstxt.FromStatusAndBytes(hostData.StatusCode, hostData.RobotsTxt)
		if err != nil {
			return result, err
		}
		result[host] = robot.FindGroup("Googlebot")
	}

	return result, err
}
