package dict

import (
	"errors"
	"fmt"
	"strings"
)

// POST - часть речи
type POST uint8

const (
	// NOUN - имя существительное
	NOUN POST = 0
	// ADJF - имя прилагательное (полное)
	ADJF POST = 1
	// ADJS - имя прилагательное (краткое)
	ADJS POST = 2
	// COMP - компаратив
	COMP POST = 3
	// VERB - глагол (личная форма)
	VERB POST = 4
	// INFN - глагол (инфинитив)
	INFN POST = 5
	// PRTF - причастие (полное)
	PRTF POST = 6
	// PRTS - причастие (краткое)
	PRTS POST = 7
	// GRND - деепричастие
	GRND POST = 8
	// NUMR - числительное
	NUMR POST = 9
	// ADVB - наречие
	ADVB POST = 10
	// NPRO - местоимение-существительное
	NPRO POST = 11
	// PRED - предикатив
	PRED POST = 12
	// PREP - предлог
	PREP POST = 13
	// CONJ - союз
	CONJ POST = 14
	// PRCL - частица
	PRCL POST = 15
	// INTJ - междометие
	INTJ POST = 16
)

var postMap = map[string]POST{
	"NOUN": NOUN,
	"ADJF": ADJF,
	"ADJS": ADJS,
	"COMP": COMP,
	"VERB": VERB,
	"INFN": INFN,
	"PRTF": PRTF,
	"PRTS": PRTS,
	"GRND": GRND,
	"NUMR": NUMR,
	"ADVB": ADVB,
	"NPRO": NPRO,
	"PRED": PRED,
	"PREP": PREP,
	"CONJ": CONJ,
	"PRCL": PRCL,
	"INTJ": INTJ}

type word struct {
	name string
	post POST
}

type wordGroup struct {
	normalForm string
	words      []word
}

type dictReader interface {
	nextGroup() (*wordGroup, error)
	isDone() bool
}

type dictRead struct {
	tokens tokenReader
	number int32
	done   bool
}

func (d *dictRead) nextGroup() (*wordGroup, error) {
	if d.done {
		return nil, errors.New("done")
	}
	t, err := d.tokens.nextToken()
	if err != nil {
		return nil, err
	}

	if t.ttype != tokenNumber {
		return nil, errors.New("first can be number line")
	}

	if d.number >= t.number {
		return nil, fmt.Errorf("unknown number %d", t.number)
	}

	d.number = t.number

	result := &wordGroup{}
	for {
		t, err = d.tokens.nextToken()
		if err != nil {
			return nil, err
		}
		if t.ttype == tokenEmptyLine {
			break
		}
		if t.ttype == tokenEnd {
			d.done = true
			break
		}
		if t.ttype == tokenData {
			name := strings.ToLower(t.data[0])
			postStr := t.data[1]
			post, ok := postMap[postStr]
			if !ok {
				return nil, fmt.Errorf("unknown POST value %s", postStr)
			}
			w := word{name: name, post: post}
			if len(result.words) == 0 {
				result.normalForm = name
			}
			result.words = append(result.words, w)
		} else {
			return nil, errors.New("unknown tag")
		}
	}

	return result, nil
}

func (d *dictRead) isDone() bool {
	return d.done
}

func createDictReader(tokens tokenReader) dictReader {
	return &dictRead{tokens: tokens, number: 0, done: false}
}
