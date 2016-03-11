package content

import "github.com/jinzhu/gorm"

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

// GetDB - create/open content.db
func GetDB() (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", "content.db")
	if err != nil {
		return nil, err
	}

	err = db.DB().Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	db.SingularTable(true)
	db.SetLogger(defaultLogger)
	db.LogMode(true)

	err = createTables(&db, Host{}, Content{}, Meta{}, URL{})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &db, nil
}
