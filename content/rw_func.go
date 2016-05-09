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
		if err == gorm.ErrRecordNotFound {
			newItem := &URL{
				URL:    baseURL,
				HostID: sql.NullInt64{Int64: host.ID, Valid: true},
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

		db.hosts[host.Name] = host
		return nil
	})
}

// GetNewURLs - get URLs for downloads for host
func (db *DBrw) GetNewURLs(hostName string, cnt int) ([]URL, error) {
	var urls []URL
	host, exists := db.hosts[hostName]
	if !exists {
		return urls, fmt.Errorf("host name %s not found in db", hostName)
	}

	err := db.Where("host_id = ? and loaded = ?", host.ID, false).Limit(cnt).Find(&urls).Error
	if err != nil {
		return urls, fmt.Errorf("find not loaded pages in 'URL' table for host %s, message: %s", hostName, err)
	}

	return urls, nil
}

// FindOrigin - get origin row id in table 'Meta'
func (db *DBrw) FindOrigin(meta *Meta) (sql.NullInt64, error) {
	null := sql.NullInt64{Valid: false}
	if meta.Content.Body.IsNull() {
		return null, nil
	}

	ids, exists := db.hashes[meta.Content.Hash]
	if !exists {
		return null, nil
	}

	for _, id := range ids {
		var content Content
		err := db.Where("id = ?", id.ContentID).Find(&content).Error
		if err != nil {
			return null, fmt.Errorf("find in 'Content' with id %d, message: %s", id.ContentID, err)
		}
		if meta.Content.Body.Equals(content.Body) {
			return sql.NullInt64{Int64: id.MetaID, Valid: true}, nil
		}
	}
	return null, nil
}

// AddHash - add new hash to hash storage
func (db *DBrw) AddHash(meta *Meta) error {
	if !meta.IsValidHash() {
		return nil
	}
	if !meta.ContentID.Valid || meta.Content.Body.IsNull() {
		return fmt.Errorf("ContentID is null in item 'Meta' for URL %s", meta.URL)
	}

	hash := meta.Content.Hash
	item := hashVal{MetaID: meta.ID, ContentID: meta.ContentID.Int64}
	_, exists := db.hashes[hash]
	if exists {
		db.hashes[hash] = append(db.hashes[hash], item)
	} else {
		db.hashes[hash] = []hashVal{item}
	}

	return nil
}
