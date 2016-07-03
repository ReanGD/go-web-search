package content

import (
	"fmt"

	"github.com/ReanGD/go-web-search/database"
	"github.com/jinzhu/gorm"
)

func createTables(db *gorm.DB, values ...interface{}) error {
	for _, value := range values {
		if !db.HasTable(value) {
			err := db.CreateTable(value).Error
			if err != nil {
				errClose := db.Close()
				if errClose != nil {
					fmt.Printf("%s", errClose)
				}
				return err
			}
		}
	}
	return nil
}

// DBrw - content database (enable read/write operations)
type DBrw struct {
	*gorm.DB
	// map[Host.Name]Host
	hosts map[string]*database.Host
	// map[Content.Hash]Content.URL
	hashes map[string]int64
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

	err = createTables(db, &database.Host{}, &database.Content{}, &database.Meta{}, &Link{}, &URL{})
	if err != nil {
		errClose := db.Close()
		if errClose != nil {
			fmt.Printf("%s", errClose)
		}
		return nil, err
	}

	var hosts []database.Host
	err = db.Find(&hosts).Error
	if err != nil {
		errClose := db.Close()
		if errClose != nil {
			fmt.Printf("%s", errClose)
		}
		return nil, fmt.Errorf("Get hosts list from db, message: %s", err)
	}

	var hashes []database.Content
	err = db.Select("url, hash").Find(&hashes).Error
	if err != nil {
		errClose := db.Close()
		if errClose != nil {
			fmt.Printf("%s", errClose)
		}
		return nil, fmt.Errorf("Get content list from db, message: \"%s\"", err)
	}

	result := &DBrw{
		DB:     db,
		hosts:  make(map[string]*database.Host, len(hosts)),
		hashes: make(map[string]int64, len(hashes))}

	for i := 0; i != len(hosts); i++ {
		result.hosts[hosts[i].Name] = &hosts[i]
	}

	for _, item := range hashes {
		result.hashes[item.Hash] = item.URL
	}

	return result, nil
}
