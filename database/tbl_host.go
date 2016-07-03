package database

// Host - host information
type Host struct {
	ID               int64  `gorm:"primary_key;not null"`
	Name             string `gorm:"size:255;unique_index;not null"`
	RobotsStatusCode int    `gorm:"not null"`
	RobotsData       []byte
}
