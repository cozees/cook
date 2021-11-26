package parser

import (
	"fmt"
	"unicode/utf8"
)

// end of file rune character
const (
	RuneEOF = -1
	bom     = 0xFEFF // unicode byte order
)

type Token uint

const (
	ILLEGAL Token = iota
	EOF
	COMMENT

	INTEGER // 123
	FLOAT   // 1.2
	BOOLEAN // true/false

	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	EQL // =
	NOT // !

	COMMA  // ,
	COLON  // :
	OPTION // ?

	CUSTOM // any custom register token should start from here
)

type Scanner interface {
	Init(src []byte)
	Finalize()
	Scan() (line, column, offset int, tok Token, lit string, err error)
	RegisterSingleCharacterToken(r rune, tok Token)

	Peek() byte
	Next() (rune, error)
	Offset() int
	NextOffset() int
}

type simpleScanner struct {
	src []byte

	ch         rune
	line       int
	offset     int
	rdOffset   int
	lineOffset int

	singleCharacter bool
	singleTables    []Token
}

func NewSimpleScanner(singleCharacter bool) Scanner {
	return &simpleScanner{singleCharacter: singleCharacter, line: 1}
}

func (ss *simpleScanner) RegisterSingleCharacterToken(r rune, tok Token) {
	if !ss.singleCharacter {
		panic("scanner is not a single character scanner")
	}
	i, tbSize := int(r), len(ss.singleTables)
	switch {
	case i == tbSize:
		ss.singleTables = append(ss.singleTables, tok)
	case i > tbSize:
		needed := i - tbSize
		for ; needed >= 0; needed-- {
			ss.singleTables = append(ss.singleTables, ILLEGAL)
		}
		fallthrough
	default:
		ss.singleTables[i] = tok
	}
}

func (ss *simpleScanner) Offset() int     { return ss.offset }
func (ss *simpleScanner) NextOffset() int { return ss.rdOffset }

func (ss *simpleScanner) Next() (rune, error) {
	if ss.rdOffset < len(ss.src) {
		ss.offset = ss.rdOffset
		if ss.ch == '\n' {
			ss.lineOffset = ss.offset
			ss.line++
		}
		r, w := rune(ss.src[ss.rdOffset]), 1
		switch {
		case r == 0:
			return -1, fmt.Errorf("illegal character NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(ss.src[ss.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				return -1, fmt.Errorf("illegal UTF-8 encoding")
			} else if r == bom && ss.offset > 0 {
				return -1, fmt.Errorf("illegal byte order mark")
			}
		case r == '\r':
			// handle \r, \r\n and \n
			nOffs := ss.rdOffset + w
			// shift next if it was newline
			if nOffs < len(ss.src) && ss.src[nOffs] == '\n' {
				ss.rdOffset = nOffs
			}
			// treat \r as \n regardless whether line feed is \r or \r\n
			// so subsequence check only need to check \n
			r = '\n'
		}
		ss.rdOffset += w
		ss.ch = r
	} else {
		ss.offset = len(ss.src)
		if ss.ch == '\n' {
			ss.lineOffset = ss.offset
			ss.line++
		}
		ss.ch = RuneEOF
	}
	return ss.ch, nil
}

func (ss *simpleScanner) Peek() byte {
	if ss.rdOffset < len(ss.src) {
		return ss.src[ss.rdOffset]
	}
	return 0
}

func (ss *simpleScanner) Init(src []byte) {
	ss.src = src
	ss.offset = 0
	ss.rdOffset = 0
	ss.line = 1
	ss.lineOffset = 0
	ss.ch = -1
}

func (ss *simpleScanner) Finalize() { ss.src = nil }

func (ss *simpleScanner) Scan() (line, column, offset int, tok Token, lit string, err error) {
	if _, err = ss.Next(); err != nil {
		return
	}
	if ss.singleCharacter {
		return ss.scanSimpleCharacter()
	}
	return
}

func (ss *simpleScanner) scanSimpleCharacter() (line, column, offset int, tok Token, lit string, err error) {
	switch ss.ch {
	case RuneEOF:
		tok = EOF
	case '+':
		tok = ADD
	case '-':
		tok = SUB
	case '*':
		tok = MUL
	case '/':
		tok = QUO
	case '%':
		tok = REM
	case '=':
		tok = EQL
	case ':':
		tok = COLON
	case ',':
		tok = COMMA
	case '?':
		tok = OPTION
	default:
		if int(ss.ch) < len(ss.singleTables) {
			tok = ss.singleTables[ss.ch]
		}
	}
	line = ss.line
	offset = ss.offset
	column = ss.offset - ss.lineOffset
	lit = string(ss.ch)
	return
}
