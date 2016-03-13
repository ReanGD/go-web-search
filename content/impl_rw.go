package content

import (
	"database/sql"
	"fmt"

	"github.com/jinzhu/gorm"
)

// Transaction - execute fn in transaction
func (db *DBrw) Transaction(fn func(*DBrw) error) error {
	tr := db.DB.Begin()
	if tr.Error != nil {
		return fmt.Errorf("create transaction, message: %s", tr.Error)
	}
	originDB := db.DB
	db.DB = tr
	userErr := fn(db)
	db.DB = originDB
	if userErr != nil {
		err := tr.Rollback().Error
		if err != nil {
			return fmt.Errorf("rollback transaction, message: %s", err)
		}
		return userErr
	}

	err := tr.Commit().Error
	if err != nil {
		return fmt.Errorf("commit transaction, message: %s", err)
	}

	return nil
}

// GetHosts - get rows from table 'Host'
func (db *DBrw) GetHosts() map[string]*Host {
	return db.hosts
}

// GetHostID - get host ID by host name
func (db *DBrw) GetHostID(hostName string) sql.NullInt64 {
	host, exists := db.hosts[hostName]
	if exists {
		return sql.NullInt64{Int64: host.ID, Valid: true}
	}
	return sql.NullInt64{Valid: false}
}

// AddHost - add new host
func (db *DBrw) AddHost(host *Host, baseURL string) error {
	return db.Transaction(func(tr *DBrw) error {
		err := tr.Create(host).Error
		if err != nil {
			return fmt.Errorf("add new 'Host' record for host %s, message: %s", host.Name, err)
		}

		var dbItem URL
		err = tr.Where("id = ?", baseURL).First(&dbItem).Error
		if err == gorm.RecordNotFound {
			newItem := &URL{ID: baseURL, HostID: sql.NullInt64{Int64: host.ID, Valid: true}, Loaded: false}
			err = tr.Create(newItem).Error
			if err != nil {
				return fmt.Errorf("add new 'URL' record for URL %s, message: %s", baseURL, err)
			}
		} else if err != nil {
			return fmt.Errorf("find in 'URL' table for URL %s, message: %s", baseURL, err)
		} else {
			// nothing to update
		}

		db.hosts[host.Name] = host
		return nil
	})
}
