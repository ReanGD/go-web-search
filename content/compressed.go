package content

import (
	"bytes"
	"compress/zlib"
	"database/sql/driver"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
)

// Compressed - field compressed by zlib
type Compressed struct {
	Data []byte
}

// CompressedHeaders - field with http headers compressed by zlib
type CompressedHeaders struct {
	Compressed
}

// Compress - compress c.Data value
func (c *Compressed) Compress() ([]byte, error) {
	if len(c.Data) == 0 {
		return nil, nil
	}

	var zContent bytes.Buffer
	w, _ := zlib.NewWriterLevelDict(&zContent, 6, nil)
	_, err := w.Write(c.Data)
	w.Close()
	if err != nil {
		return nil, fmt.Errorf("zlib write error: %s", err)
	}

	return zContent.Bytes(), nil
}

// Value - prepare value for save to DB
func (c Compressed) Value() (driver.Value, error) {
	return c.Compress()
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
		return errors.New("Scan source was not string")
	}

	r, err := zlib.NewReader(bytes.NewReader(obj))
	if err != nil {
		return fmt.Errorf("zlib new reader error: %s", err)
	}
	result, err := ioutil.ReadAll(r)
	r.Close()
	if err != nil {
		return fmt.Errorf("zlib read all error: %s", err)
	}
	c.Data = result

	return nil
}

// IsNull - check is null
func (c *Compressed) IsNull() bool {
	return len(c.Data) == 0
}

// Equals - check is equals
func (c *Compressed) Equals(other Compressed) bool {
	return bytes.Equal([]byte(c.Data), []byte(other.Data))
}

// Set - convert headers to bytes
func (c *CompressedHeaders) Set(headers map[string]string) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(headers)
	if err != nil {
		return fmt.Errorf("serialize headers to bytes, error: %s", err)
	}
	c.Data = buf.Bytes()

	return nil
}

// Get - convert bytes to headers
func (c *CompressedHeaders) Get() (map[string]string, error) {
	decoder := gob.NewDecoder(bytes.NewReader(c.Data))
	var headers map[string]string
	err := decoder.Decode(&headers)
	if err != nil {
		return headers, fmt.Errorf("deserialize bytes to headers, error: %s", err)
	}

	return headers, nil
}
