package content

import "database/sql"

// Link - links between pages
type Link struct {
	Master int64 `gorm:"type:integer REFERENCES url(id);index;not null"`
	Slave  int64 `gorm:"type:integer REFERENCES url(id);index;not null"`
}

// URL - struct for save all URLs in db
type URL struct {
	ID     int64         `gorm:"primary_key;not null"`
	URL    string        `gorm:"size:2048;not null;unique_index"`
	HostID sql.NullInt64 `gorm:"type:integer REFERENCES host(id);index"`
	Loaded bool          `gorm:"not null;index"`
}
