package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"unicode"
	"unicode/utf8"
)

type tokenType uint8

const (
	tokenText   tokenType = 0
	tokenNumber tokenType = 1
	tokenEmpty  tokenType = 2
	tokenEnd    tokenType = 3
)

type token struct {
	text   string
	number int32
	ttype  tokenType
}

type dictParser struct {
	number  int32
	finish  bool
	scanner *bufio.Scanner
	words   []string
}

func (p *dictParser) nextLine() (bool, []byte, error) {
	done := !p.scanner.Scan()

	if done {
		return done, nil, nil
	}

	return done, p.scanner.Bytes(), p.scanner.Err()
}

func (p *dictParser) nextToken() (token, error) {
	done, data, err := p.nextLine()

	if err != nil {
		return token{}, err
	}

	if done {
		return token{ttype: tokenEnd}, nil
	}

	if len(data) == 0 {
		return token{ttype: tokenEmpty}, nil
	}

	var buffer bytes.Buffer
	var rRaw, r rune
	var size int
	for len(data) > 0 {
		rRaw, size = utf8.DecodeRune(data)
		r = unicode.ToLower(rRaw)
		if r == '\t' {
			return token{text: buffer.String(), ttype: tokenText}, nil
		}
		_, _ = buffer.WriteRune(r)
		data = data[size:]
	}

	text := buffer.String()
	number, err := strconv.ParseInt(text, 10, 32)
	if err != nil {
		return token{}, err
	}

	return token{number: int32(number), ttype: tokenNumber}, nil
}

func (p *dictParser) nextWords() ([]string, error) {
	if p.finish {
		return nil, nil
	}
	t, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if t.ttype != tokenNumber {
		return nil, errors.New("unknown token")
	}

	if p.number >= t.number {
		return nil, fmt.Errorf("unknown number %d", t.number)
	}
	p.number = t.number

	var result []string
	for {
		t, err = p.nextToken()
		if err != nil {
			return nil, err
		} else if t.ttype == tokenText {
			result = append(result, t.text)
		} else if t.ttype == tokenEnd {
			p.finish = true
			break
		} else if t.ttype == tokenEmpty {
			break
		} else {
			return nil, errors.New("unknown token")
		}
	}

	if len(result) == 0 {
		return nil, errors.New("empty array")
	}

	return result, nil
}

func (p *dictParser) parse() error {
	for {
		words, err := p.nextWords()

		if err != nil {
			return err
		}
		if words == nil {
			break
		}

		for _, word := range words {
			p.words = append(p.words, word)
			// fmt.Println(word)
		}

		// fmt.Println("!")
	}

	fmt.Println("success end")
	return nil
}

func (p *dictParser) Run(path string) error {
	p.number = 0
	p.finish = false
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	p.scanner = bufio.NewScanner(file)

	err = p.parse()
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	return err
}

func runImportDict() error {
	var d dictParser
	return d.Run("dict.opcorpora.txt")
}
