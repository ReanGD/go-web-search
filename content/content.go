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
				errClose := db.Close()
				if errClose != nil {
					fmt.Printf("%s", err)
				}
				return err
			}
		}
	}
	return nil
}

type hashVal struct {
	MetaID    int64
	ContentID int64
}

// DBrw - content database (enable read/write operations)
type DBrw struct {
	*gorm.DB
	// map[Host.Name]Host
	hosts map[string]*Host
	// map[Content.Hash][]hashVal
	hashes map[string][]hashVal
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

	err = createTables(db, &Host{}, &Content{}, &Meta{}, &URL{})
	if err != nil {
		errClose := db.Close()
		if errClose != nil {
			fmt.Printf("%s", err)
		}
		return nil, err
	}

	var hosts []Host
	err = db.Find(&hosts).Error
	if err != nil {
		errClose := db.Close()
		if errClose != nil {
			fmt.Printf("%s", err)
		}
		return nil, fmt.Errorf("Get hosts list from db, message: %s", err)
	}

	type hashResult struct {
		MetaID    int64
		ContentID int64
		Hash      string
	}

	var hashResults []hashResult
	sql := "SELECT meta.id as meta_id, content.id as content_id, content.hash as hash"
	sql += " FROM meta JOIN content ON content.id == meta.content_id"
	sql += " WHERE meta.content_id IS NOT NULL AND meta.state IN (?, ?)"
	err = db.Raw(sql, StateSuccess, StateParseError).Scan(&hashResults).Error
	if err != nil {
		errClose := db.Close()
		if errClose != nil {
			fmt.Printf("%s", err)
		}
		return nil, fmt.Errorf("Get content list from db, message: %s", err)
	}

	result := &DBrw{
		DB:     db,
		hosts:  make(map[string]*Host, len(hosts)),
		hashes: make(map[string][]hashVal, len(hashResults))}

	for i := 0; i != len(hosts); i++ {
		result.hosts[hosts[i].Name] = &hosts[i]
	}

	for _, item := range hashResults {
		hash := item.Hash
		item := hashVal{MetaID: item.MetaID, ContentID: item.ContentID}
		_, exists := result.hashes[hash]
		if exists {
			result.hashes[hash] = append(result.hashes[hash], item)
		} else {
			result.hashes[hash] = []hashVal{item}
		}
	}

	return result, nil
}
