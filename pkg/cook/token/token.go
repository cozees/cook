package token

import "reflect"

type Token int

const (
	ILLEGAL Token = iota
	EOF
	// type specific
	COMMENT
	literal_beg
	// order of the below token is important to has the same order as
	// TINTEGER, TFLOAT ...etc
	INTEGER // 125
	FLOAT   // 1.23
	BOOLEAN // true, false
	STRING  // 'text', "text", `text`, text
	ARRAY   // {1:2, 3:2}
	MAP     // [1, 2, 3]
	// none order needed
	IDENT      // name, variable
	STRING_ITP // sample $A text
	literal_end

	operator_beg
	ADD            // +
	SUB            // -
	MUL            // *
	QUO            // /
	REM            // %
	ADD_ASSIGN     // +=
	SUB_ASSIGN     // -=
	MUL_ASSIGN     // *=
	QUO_ASSIGN     // /=
	REM_ASSIGN     // %=
	INC            // ++
	DEC            // --
	AND            // &
	LAND           // &&
	OR             // |
	LOR            // ||
	XOR            // ^
	SHL            // <<
	SHR            // >>
	AND_NOT        // &^
	AND_ASSIGN     // &=
	OR_ASSIGN      // |=
	AND_NOT_ASSIGN // &^=
	ASSIGN         // =
	LAMBDA         // =>
	NOT            // !
	EQL            // ==
	NEQ            // !=
	LSS            // <
	LEQ            // <=
	GTR            // >
	GEQ            // >=
	READ_FROM      // <, it work exclusively in argument scanning
	WRITE_TO       // >, it work exclusively in argument scanning
	APPEND_TO      // >>, it work exclusively in argument scanning
	AT             // @
	HASH           // #
	VAR            // $
	QES            // ?
	DQS            // ??
	FD             // ~
	PIPE           // |, it work exclusively in argument scanning

	LBRACK // [
	LBRACE // {
	RBRACK // ]
	RBRACE // }
	LPAREN // (
	RPAREN // )
	COMMA  // ,
	COLON  // :
	LF     // \n,\r,\r\n: linefeed
	operator_end

	keyword_beg
	FOR
	IS
	IF
	ELSE
	RETURN
	IN
	EXIT
	RANGE
	SIZEOF
	BREAK
	CONTINUE
	INCLUDE
	DELETE
	ON
	EXISTS

	// operating system keyword
	LINUX
	MACOS
	WINDOWS

	// type keyword specific
	type_rep_beg
	TINTEGER
	TFLOAT
	TBOOLEAN
	TSTRING
	TARRAY
	TMAP
	TOBJECT
	type_rep_end
	// end of keyword session
	keyword_end
)

var tokens = [...]string{
	ILLEGAL:        "ILLEGAL",
	COMMENT:        "COMMENT",
	INTEGER:        "INTEGER",
	FLOAT:          "FLOAT",
	BOOLEAN:        "BOOLEAN",
	STRING:         "STRING",
	STRING_ITP:     "STRING_ITP",
	MAP:            "MAP",
	ARRAY:          "ARRAY",
	IDENT:          "IDENT",
	ADD:            "+",
	SUB:            "-",
	MUL:            "*",
	QUO:            "/",
	REM:            "%",
	ADD_ASSIGN:     "+=",
	SUB_ASSIGN:     "-=",
	MUL_ASSIGN:     "*=",
	QUO_ASSIGN:     "/=",
	REM_ASSIGN:     "%=",
	AND_ASSIGN:     "&=",
	OR_ASSIGN:      "|=",
	AND_NOT_ASSIGN: "&^=",
	AND:            "&",
	OR:             "|",
	INC:            "++",
	DEC:            "--",
	LAND:           "&&",
	LOR:            "||",
	XOR:            "^",
	SHL:            "<<",
	SHR:            ">>",
	AND_NOT:        "&^",
	ASSIGN:         "=",
	EQL:            "==",
	LAMBDA:         "=>",
	NOT:            "!",
	NEQ:            "!=",
	LSS:            "<",
	LEQ:            "<=",
	GTR:            ">",
	GEQ:            ">=",
	READ_FROM:      "<",
	WRITE_TO:       ">",
	APPEND_TO:      ">>",
	AT:             "@",
	HASH:           "#",
	VAR:            "$",
	QES:            "?",
	DQS:            "??",
	FD:             "~",
	PIPE:           "|",
	LBRACK:         "[",
	LBRACE:         "{",
	RBRACK:         "]",
	RBRACE:         "}",
	LPAREN:         "(",
	RPAREN:         ")",
	COMMA:          ",",
	COLON:          ":",
	LF:             "\n",
	FOR:            "for",
	IS:             "is",
	IF:             "if",
	ELSE:           "else",
	RETURN:         "return",
	IN:             "in",
	EXIT:           "exit",
	RANGE:          "..",
	SIZEOF:         "sizeof",
	BREAK:          "break",
	CONTINUE:       "continue",
	INCLUDE:        "include",
	DELETE:         "delete",
	ON:             "on",
	EXISTS:         "exists",
	LINUX:          "linux",
	MACOS:          "darwin",
	WINDOWS:        "windows",
	TINTEGER:       "integer",
	TFLOAT:         "float",
	TBOOLEAN:       "boolean",
	TSTRING:        "string",
	TARRAY:         "array",
	TMAP:           "map",
	TOBJECT:        "object",
}

func (t Token) String() string { return tokens[t] }

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keyword_beg; i < keyword_end; i++ {
		if i != type_rep_beg && i != type_rep_end {
			keywords[tokens[i]] = i
		}
	}
}

func Lookup(ident string, tok Token) Token {
	if tok, is_keyword := keywords[ident]; is_keyword {
		return tok
	}
	return tok
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

func (tok Token) Kind() reflect.Kind {
	if literal_beg < tok && tok < IDENT {
		switch tok {
		case INTEGER:
			return reflect.Int64
		case FLOAT:
			return reflect.Float64
		case BOOLEAN:
			return reflect.Bool
		case STRING:
			return reflect.String
		case MAP:
			return reflect.Map
		case ARRAY:
			return reflect.Slice
		}
	} else if type_rep_beg < tok && tok < TOBJECT {
		switch tok {
		case TINTEGER:
			return reflect.Int64
		case TFLOAT:
			return reflect.Float64
		case TBOOLEAN:
			return reflect.Bool
		case TSTRING:
			return reflect.String
		case TMAP:
			return reflect.Map
		case TARRAY:
			return reflect.Slice
		}
	}
	return reflect.Invalid
}

func (tok Token) Type() int {
	if type_rep_beg < tok && tok < type_rep_end {
		return 1 << (int(tok-type_rep_beg) - 1)
	}
	return 0
}

func (tok Token) IsComparison() bool {
	return EQL <= tok && tok <= GEQ || tok == IS
}

func (tok Token) IsLogicOperator() bool {
	return EQL <= tok && tok <= GEQ || tok == IS || tok == LAND || tok == LOR
}

const (
	LowestPrec  = 0 // non-operators
	UnaryPrec   = 7
	HighestPrec = 8
)

func (op Token) Precedence() int {
	switch op {
	case LOR:
		return 1
	case LAND:
		return 2
	case QES, DQS:
		return 3
	case EQL, NEQ, LSS, LEQ, GTR, GEQ, IS:
		return 4
	case ADD, SUB, OR, XOR:
		return 5
	case MUL, QUO, REM, SHL, SHR, AND, AND_NOT:
		return 6
	}
	return LowestPrec
}
