package dict

import (
	"bufio"
	"io"
	"os"
)

type lineReader struct {
	file    *os.File
	scanner *bufio.Scanner
}

func (r *lineReader) Read(p []byte) (int, error) {
	if !r.scanner.Scan() {
		p = nil
		return 0, nil
	}

	err := r.scanner.Err()
	if err != nil {
		return 0, err
	}

	p = r.scanner.Bytes()
	return len(p), err
}

func (r *lineReader) Close() error {
	return r.file.Close()
}

func createLineReader(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	result := &lineReader{
		file:    file,
		scanner: bufio.NewScanner(file)}

	return result, nil
}
