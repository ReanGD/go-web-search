package dict

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type tokenType uint8

const (
	tokenEmptyLine tokenType = 0
	tokenNumber    tokenType = 1
	tokenData      tokenType = 2
	tokenEnd       tokenType = 3
)

type token struct {
	data   []string
	number int32
	ttype  tokenType
}

type tokenReader interface {
	nextTokens() (*token, error)
}

type tokenRead struct {
	scanner *bufio.Scanner
}

func splitFunc(c rune) bool {
	return c == ' ' || c == ',' || c == '\t'
}

func (t *tokenRead) nextTokens() (*token, error) {
	if !t.scanner.Scan() {
		return &token{ttype: tokenEnd}, nil
	}

	err := t.scanner.Err()
	if err != nil {
		return nil, err
	}

	data := t.scanner.Text()
	if len(data) == 0 {
		return &token{ttype: tokenEmptyLine}, nil
	}

	words := strings.FieldsFunc(data, splitFunc)
	if len(words) == 0 {
		return &token{ttype: tokenEmptyLine}, nil
	}

	if len(words) == 1 {
		number, err := strconv.ParseInt(words[0], 10, 32)
		if err != nil {
			return nil, err
		}
		return &token{number: int32(number), ttype: tokenNumber}, nil
	}

	return &token{data: words, ttype: tokenData}, nil
}

func createTokenReader(r io.Reader) tokenReader {
	return &tokenRead{scanner: bufio.NewScanner(r)}
}
