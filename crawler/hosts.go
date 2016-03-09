package crawler

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/ReanGD/go-web-search/content"
	"github.com/jinzhu/gorm"
)

func fillHost(hostName string) (*content.Host, error) {
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

	result := &content.Host{ID: hostName,
		RobotsStatusCode: response.StatusCode,
		RobotsData:       body,
		Timestamp:        time.Now()}

	return result, nil
}

func fillHosts(baseHosts map[string]bool, db *gorm.DB) error {
	for hostName, isNew := range baseHosts {
		if isNew {
			host, err := fillHost(hostName)
			if err != nil {
				return err
			}

			err = db.Create(host).Error
			if err != nil {
				return err
			}

			urlKey := URLFromHost(hostName)
			var url content.URL
			err = db.Where("id = ?", urlKey).First(&url).Error
			if err == gorm.RecordNotFound {
				url = content.URL{ID: urlKey, Loaded: false}
				err = db.Create(&url).Error
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}

func initHosts(baseHosts map[string]bool, db *gorm.DB) (int, error) {
	var hosts []content.Host
	err := db.Find(&hosts).Error
	if err != nil {
		return 0, nil
	}

	totalCnt := len(hosts)

	needFill := false
	for hostName := range baseHosts {
		isNew := true
		for _, host := range hosts {
			if host.ID == hostName {
				isNew = false
				break
			}
		}
		if isNew {
			needFill = true
			totalCnt++
		}
		baseHosts[hostName] = isNew
	}

	if needFill {
		tr := db.Begin()
		err = tr.Error
		if err == nil {
			err = fillHosts(baseHosts, tr)
			if err != nil {
				tr.Rollback()
			} else {
				err = tr.Commit().Error
			}
		}
	}

	return totalCnt, err
}
