package token

import (
	"strconv"
)

type Token int

const (
	ILLEGAL Token = iota
	EOF
	COMMENT

	literal_beg
	IDENT
	INTEGER
	FLOAT
	STRING
	BOOLEAN
	literal_end

	operator_beg
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	ADD_ASSIGN // +=
	SUB_ASSIGN // -=
	MUL_ASSIGN // *=
	QUO_ASSIGN // /=
	REM_ASSIGN // %=

	AND // &
	OR  // |
	XOR // ^
	SHL // <<
	SHR // >>

	LAND    // &&
	LOR     // ||
	INC     // ++
	DEC     // --
	AND_NOT // &^

	EQL // ==
	LSS // <
	GTR // >
	NEQ // !=
	LEQ // <=
	GEQ // >=

	NOT // !

	ASSIGN    // =, assign or define
	READ_FROM // <, special case from redirect content from file
	ASSIGN_TO // >, special case for redirect content to file
	APPEND_TO // >>, special case for redirect content to append into exit file or create a new file if not exist

	LBRACK // [
	LBRACE // {
	RBRACK // ]
	RBRACE // }
	LPAREN // (
	RPAREN // )

	COMMA // ,
	COLON // :
	AT    // @
	HASH  // #
	LF    // \n, \r, \r\n : line feed
	VAR   // $
	QES   // ?
	DQES  // ??
	operator_end

	keyword_beg
	IF
	ELSE
	EXIT
	IS
	IN
	FOR
	RANGE
	SIZEOF
	CONTINUE
	BREAK
	INCLUDE
	// type keyword
	keyword_type_beg
	TINTEGER
	TFLOAT
	TSTRING
	TBOOLEAN
	TARRAY
	TMAP
	TOBJECT
	keyword_type_end
	keyword_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "COMMENT",

	IDENT:   "IDENT",
	INTEGER: "INTEGER",
	FLOAT:   "FLOAT",
	STRING:  "STRING",
	BOOLEAN: "BOOLEAN",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "%",

	ADD_ASSIGN: "+=",
	SUB_ASSIGN: "-=",
	MUL_ASSIGN: "*=",
	QUO_ASSIGN: "/=",
	REM_ASSIGN: "%=",

	AND:     "&",
	OR:      "|",
	XOR:     "^",
	SHL:     "<<",
	SHR:     ">>",
	AND_NOT: "&^",

	LAND: "&&",
	LOR:  "||",
	INC:  "++",
	DEC:  "--",

	EQL:       "==",
	LSS:       "<",
	GTR:       ">",
	ASSIGN:    "=",
	READ_FROM: "<",
	ASSIGN_TO: ">",
	APPEND_TO: ">>",
	NOT:       "!",

	NEQ: "!=",
	LEQ: "<=",
	GEQ: ">=",

	LBRACK: "[",
	LBRACE: "{",
	LPAREN: "(",
	RBRACK: "]",
	RBRACE: "}",
	RPAREN: ")",

	COMMA: ",",
	COLON: ":",
	AT:    "@",
	HASH:  "#",
	VAR:   "$",
	QES:   "?",
	DQES:  "??",
	LF:    "\n",

	IF:       "if",
	ELSE:     "else",
	EXIT:     "exit",
	IS:       "is",
	IN:       "in",
	FOR:      "for",
	RANGE:    "range",
	SIZEOF:   "sizeof",
	CONTINUE: "continue",
	BREAK:    "break",
	INCLUDE:  "include",

	TINTEGER: "integer",
	TFLOAT:   "float",
	TSTRING:  "string",
	TBOOLEAN: "boolean",
	TARRAY:   "array",
	TMAP:     "map",
	TOBJECT:  "object",
}

func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keyword_beg + 1; i < keyword_end; i++ {
		if i != keyword_type_beg && i != keyword_type_end { // skip mark of type keyword
			keywords[tokens[i]] = i
		}
	}
}

func Lookup(ident string) Token {
	if tok, is_keyword := keywords[ident]; is_keyword {
		return tok
	}
	return IDENT
}

func IsKeyword(name string) bool {
	_, ok := keywords[name]
	return ok
}

func IsIdentifier(name string) bool {
	for i, c := range name {
		if !('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z') && c != '_' && (i == 0 || !('0' <= c && c <= '9')) {
			return false
		}
	}
	return name != "" && !IsKeyword(name)
}

func (tok Token) IsLiteral() bool { return literal_beg < tok && tok < literal_end }

func (tok Token) IsOperator() bool {
	return (operator_beg < tok && tok < operator_end)
}

func (tok Token) IsKeyword() bool { return keyword_beg < tok && tok < keyword_end }

func (tok Token) Type() int {
	if keyword_type_beg < tok && tok < keyword_type_end {
		return 1 << (int(tok-keyword_type_beg) - 1)
	}
	return 0
}

const (
	LowestPrec  = 0 // non-operators
	UnaryPrec   = 6
	HighestPrec = 7
)

func (op Token) Precedence() int {
	switch op {
	case LOR:
		return 1
	case LAND:
		return 2
	case EQL, NEQ, LSS, LEQ, GTR, GEQ, QES, DQES, IS:
		return 3
	case ADD, SUB, OR, XOR:
		return 4
	case MUL, QUO, REM, SHL, SHR, AND, AND_NOT:
		return 5
	}
	return LowestPrec
}
