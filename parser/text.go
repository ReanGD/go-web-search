package parser

import (
	"container/list"
	"regexp"
	"strings"

	"github.com/kljensen/snowball"
)

var regexRus = regexp.MustCompile(`[а-яё][0-9а-яё]*`)

// ParseText - convert list of strings to list of words
func ParseText(stringList list.List) (map[string]uint32, error) {
	result := make(map[string]uint32)

	for it := stringList.Front(); it != nil; it = it.Next() {
		text := it.Value.(string)
		if len(text) > 2 {
			words := regexRus.FindAllString(strings.ToLower(text), -1)
			for _, word := range words {
				if len(word) > 2 {
					stemmed, err := snowball.Stem(word, "russian", true)
					if err != nil {
						return result, err
					}
					if val, ok := result[stemmed]; ok {
						result[stemmed] = val + 1
					} else {
						result[stemmed] = 1
					}
				}
			}
		}
	}

	return result, nil
}
