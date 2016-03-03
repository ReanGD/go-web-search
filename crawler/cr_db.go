package crawler

import (
	"time"

	"github.com/boltdb/bolt"
	"github.com/tinylib/msgp/msgp"
)

// DB - bolt.DB for page.db
type DB struct {
	*bolt.DB
}

// Tx - bolt.Tx for page.db
type Tx struct {
	*bolt.Tx
}

// Bucket - bolt.Bucket for page.db
type Bucket struct {
	*bolt.Bucket
}

// Open - open or create database
func (db *DB) Open() error {
	var err error
	db.DB, err = bolt.Open(DbName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	err = db.DB.Update(func(tx *bolt.Tx) error {
		var err error
		_, err = tx.CreateBucketIfNotExists([]byte(DbBucketContents))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(DbBucketWrongURLs))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(DbBucketURLs))
		if err != nil {
			return err
		}
		bMeta, err := tx.CreateBucketIfNotExists([]byte(DbBucketMeta))
		if err != nil {
			return err
		}

		metaVal := DbMeta{LastID: 0}
		bytes, err := metaVal.MarshalMsg(nil)
		if err != nil {
			return err
		}

		err = bMeta.Put([]byte(DbKeyMeta), bytes)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		db.Close()
	}

	return err
}

// Close - close database
func (db *DB) Close() {
	db.DB.Close()
}

// Update - update database
func (db *DB) Update(fn func(*Tx) error) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		childTx := &Tx{Tx: tx}
		return fn(childTx)
	})
}

// View - view database
func (db *DB) View(fn func(*Tx) error) error {
	return db.DB.View(func(tx *bolt.Tx) error {
		childTx := &Tx{Tx: tx}
		return fn(childTx)
	})
}

// Bucket - get wrapper of bucket
func (tx *Tx) Bucket(name string) *Bucket {
	return &Bucket{Bucket: tx.Tx.Bucket([]byte(name))}
}

// Put - put operation wrapper
func (b *Bucket) Put(key []byte, value msgp.Marshaler) error {
	bytes, err := value.MarshalMsg(nil)
	if err != nil {
		return err
	}
	return b.Bucket.Put(key, bytes)
}

// Get - get operation wrapper
func (b *Bucket) Get(key []byte, value msgp.Unmarshaler) (bool, error) {
	bytes := b.Bucket.Get(key)
	if bytes != nil {
		_, err := value.UnmarshalMsg(bytes)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}
