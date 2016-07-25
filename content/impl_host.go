package content

import (
	"database/sql"
	"fmt"

	"github.com/ReanGD/go-web-search/database"
	"github.com/ReanGD/go-web-search/proxy"
	"github.com/jinzhu/gorm"
)

// GetHosts - get rows from table 'Host'
func (db *DBrw) GetHosts() (map[int64]*proxy.Host, error) {
	result := make(map[int64]*proxy.Host)

	var hosts []database.Host
	err := db.Find(&hosts).Error
	if err != nil {
		return result, fmt.Errorf("Get hosts list from db, message: %s", err)
	}

	for _, host := range hosts {
		result[host.ID] = proxy.NewHost(host.Name, host.RobotsStatusCode, host.RobotsData)
	}

	return result, nil
}

// AddHost - add new host
func (db *DBrw) AddHost(host *proxy.Host, baseURL string) (int64, error) {
	var id int64
	err := db.Transaction(func(tr *DBrw) error {
		tblHost := host.GetTable()
		err := tr.Create(tblHost).Error
		if err != nil {
			return fmt.Errorf("add new 'Host' record for host %s, message: %s", host.GetName(), err)
		}

		id = tblHost.ID

		var dbItem URL
		err = tr.Where("id = ?", baseURL).First(&dbItem).Error
		if err == gorm.ErrRecordNotFound {
			newItem := &URL{
				URL:    baseURL,
				HostID: sql.NullInt64{Int64: id, Valid: true},
				Loaded: false}
			err = tr.Create(newItem).Error
			if err != nil {
				return fmt.Errorf("add new 'URL' record for URL %s, message: %s", baseURL, err)
			}
		} else if err != nil {
			return fmt.Errorf("find in 'URL' table for URL %s, message: %s", baseURL, err)
		} else {
			// nothing to update
		}

		return nil
	})

	return id, err
}
