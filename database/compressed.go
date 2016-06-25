package database

import (
	"bytes"
	"compress/zlib"
	"database/sql/driver"
	"errors"
	"fmt"
	"io/ioutil"
)

const (
	// ErrScanArgument - Scan argument is not array of bytes
	ErrScanArgument = "Compressed.Scan: argument is not array of bytes"
)

// Compressed - field compressed by zlib
type Compressed struct {
	Data []byte
}

// Compress - compress c.Data value
func (c *Compressed) Compress() []byte {
	if len(c.Data) == 0 {
		return nil
	}

	var zContent bytes.Buffer
	w, _ := zlib.NewWriterLevelDict(&zContent, 6, nil)
	_, _ = w.Write(c.Data)
	_ = w.Close()

	return zContent.Bytes()
}

// Value - prepare value for save to DB
func (c Compressed) Value() (driver.Value, error) {
	return c.Compress(), nil
}

// Scan - load data from DB to value
func (c *Compressed) Scan(value interface{}) error {
	if value == nil {
		var nilResult []byte
		c.Data = nilResult
		return nil
	}

	obj, ok := value.([]byte)
	if !ok {
		return errors.New(ErrScanArgument)
	}

	r, err := zlib.NewReader(bytes.NewReader(obj))
	if err != nil {
		return fmt.Errorf("Compressed.Scan: %s", err)
	}
	result, err := ioutil.ReadAll(r)
	_ = r.Close()
	if err != nil {
		return fmt.Errorf("Compressed.Scan: %s", err)
	}
	c.Data = result

	return nil
}
