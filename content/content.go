package content

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

func createTables(db *gorm.DB, values ...interface{}) error {
	for _, value := range values {
		if !db.HasTable(value) {
			err := db.CreateTable(value).Error
			if err != nil {
				db.Close()
				return err
			}
		}
	}
	return nil
}

// DBrw - content database (enable read/write operations)
type DBrw struct {
	*gorm.DB
	hosts map[string]*Host
}

// GetDBrw - create/open content.db
func GetDBrw() (*DBrw, error) {
	db, err := gorm.Open("sqlite3", "file:content.db?cache=shared")
	// db, err := gorm.Open("sqlite3", "content.db")
	if err != nil {
		return nil, err
	}

	// db.DB().SetMaxIdleConns(1)
	// db.DB().SetMaxOpenConns(2)
	db.SingularTable(true)
	db.SetLogger(defaultLogger)
	db.LogMode(false)

	err = createTables(&db, Host{}, Content{}, Meta{}, URL{})
	if err != nil {
		db.Close()
		return nil, err
	}

	var hosts []Host
	err = db.Find(&hosts).Error
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("Get hosts list from db, message: %s", err)
	}

	result := &DBrw{DB: &db, hosts: make(map[string]*Host, len(hosts))}

	for i := 0; i != len(hosts); i++ {
		result.hosts[hosts[i].Name] = &hosts[i]
	}

	return result, nil
}
