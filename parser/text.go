package parser

import (
	"container/list"
	"regexp"
	"strings"

	"github.com/kljensen/snowball"
)

var regexRus = regexp.MustCompile(`[а-яё][0-9а-яё]*`)

// ParseText - convert list of strings to list of words
func ParseText(stringList list.List) (list.List, error) {
	var result list.List

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
					result.PushBack(stemmed)
				}
			}
		}
	}

	return result, nil
}
