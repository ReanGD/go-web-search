package content

import (
	"database/sql"
	"fmt"
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

// GetNewURLs - get URLs for downloads for host
func (db *DBrw) GetNewURLs(hostID int64, cnt int) ([]URL, error) {
	var urls []URL
	err := db.Where("host_id = ? and loaded = ?", hostID, false).Limit(cnt).Find(&urls).Error
	if err != nil {
		return urls, fmt.Errorf("find not loaded pages in 'URL' table for host %d, message: %s", hostID, err)
	}

	return urls, nil
}

// FindOrigin - find origin url id in table 'URL'
func (db *DBrw) FindOrigin(hash string) sql.NullInt64 {
	id, exists := db.hashes[hash]
	if !exists {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: id, Valid: true}
}

// AddHash - add new hash to hash storage
func (db *DBrw) AddHash(hash string, urlID int64) {
	db.hashes[hash] = urlID
}
