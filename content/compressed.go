package content

import (
	"bytes"
	"compress/zlib"
	"database/sql/driver"
	"errors"
	"fmt"
	"io/ioutil"
)

// Compressed - field compressed by zlib
type Compressed struct {
	Data []byte
}

// Value - compress value in field
func (c Compressed) Value() (driver.Value, error) {
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

// Scan - uncompress field to value
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
func (c Compressed) IsNull() bool {
	return len(c.Data) == 0
}

// Equals - check is equals
func (c Compressed) Equals(other Compressed) bool {
	return bytes.Equal([]byte(c.Data), []byte(other.Data))
}
