package parser

import (
	"bytes"
	"unicode/utf8"

	"github.com/cozees/cook/pkg/cook/ast"
	"github.com/cozees/cook/pkg/cook/token"
)

const (
	bom = 0xFEFF // byte order mark, only permitted as very first character
	eof = -1     // end of file
)

type ErrorHandler func(offset int, file *token.File, msg string) bool

type Scanner struct {
	file *token.File
	src  []byte
	err  ErrorHandler

	ch         rune
	offset     int
	rdOffset   int
	lineOffset int

	skipLineFeed bool
	prevTokens   [2]token.Token
	argumentMode bool
}

func NewScanner(file *token.File, src []byte, err ErrorHandler) *Scanner {
	s := &Scanner{
		file:         file,
		src:          src,
		err:          err,
		skipLineFeed: false,
	}
	s.reset(0)
	return s
}

func (s *Scanner) reset(offs int) {
	s.ch = eof
	s.offset = -1
	s.rdOffset = offs
	if len(s.src) > 0 {
		// find early line feed
		for i := offs; i >= 0; i-- {
			if s.src[i] == '\n' || s.src[i] == '\r' {
				s.lineOffset = i
			}
		}
		s.skipLineFeed = true
		s.prevTokens[0], s.prevTokens[1] = token.ILLEGAL, token.ILLEGAL
		s.argumentMode = false
		s.next()
		if s.ch == bom {
			s.next() // ignore BOM at file beginning
		}
	}
}

func (s *Scanner) error(offset int, msg string) {
	if s.err != nil {
		s.err(offset, s.file, msg)
	}
}

func (s *Scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		r, w := rune(s.src[s.rdOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "illegal character NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(s.src[s.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error(s.offset, "illegal UTF-8 encoding")
			} else if r == bom && s.offset > 0 {
				s.error(s.offset, "illegal byte order mark")
			}
		case r == '\r':
			// handle \r, \r\n and \n
			nOffs := s.rdOffset + w
			// shift next if it was newline
			if nOffs < len(s.src) && s.src[nOffs] == '\n' {
				s.rdOffset = nOffs
			}
			// treat \r as \n regardless whether line feed is \r or \r\n
			// so subsequence check only need to check \n
			r = '\n'
		}
		s.rdOffset += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		s.ch = eof
	}
}

func (s *Scanner) peek() byte {
	if s.rdOffset < len(s.src) {
		return s.src[s.rdOffset]
	}
	return 0
}

func (s *Scanner) Scan() (offs int, tok token.Token, lit string, ci ast.Expr) {
readArgument:
	if s.argumentMode {
		return s.scanArgument()
	}

	if s.prevTokens[1] == token.IDENT &&
		(s.prevTokens[0] == token.HASH || s.prevTokens[0] == token.AT) {
		switch s.ch {
		case ':':
			goto nonArgument
		case '\n', eof:
			// calling external command or target with no argument
			offs, tok, lit = s.offset, token.LF, "\n"
			s.argumentMode = false
			s.skipLineFeed = true
			s.next()
			s.prevTokens[0], s.prevTokens[1] = token.ILLEGAL, token.ILLEGAL
			return
		default:
			s.argumentMode = true
			s.prevTokens[0], s.prevTokens[1] = token.ILLEGAL, token.ILLEGAL
			goto readArgument
		}
	}

nonArgument:
	// add token to the queue, determine next whether we should read argument instead
	defer func() {
		s.prevTokens[0], s.prevTokens[1] = s.prevTokens[1], tok
	}()

revisit:
	isBeginLine := s.lineOffset == s.offset
	skipLineFeed := true
	s.skipWhitespace()

	offs = s.offset
	switch ch := s.ch; {
	case ch == eof:
		if !s.skipLineFeed {
			tok = token.LF
			lit = "\n"
			s.skipLineFeed = true
			return
		}
		tok = token.EOF
		return
	case isLineFeed(ch):
		s.next()
		if isBeginLine {
			// empty line let scan next line
			goto revisit
		}
		tok = token.LF
		lit = "\n"
		s.skipLineFeed = true
		return
	case isLetter(ch):
		lit = s.scanIdentifier()
		tok = token.Lookup(lit)
		switch tok {
		case token.IDENT, token.BREAK, token.CONTINUE:
			skipLineFeed = false
		}
	case isDecimal(ch):
		tok, lit = s.scanNumber()
		skipLineFeed = false
	default:
		s.next()
		switch ch {
		case '\'', '"':
			lit, ci = s.scanString(ch)
			tok = token.STRING
			skipLineFeed = false
		case '@':
			tok = token.AT
		case '$':
			tok = token.VAR
		case '?':
			tok = s.ternary(s.ch == '?', token.DQES, token.QES)
		case '!':
			tok = s.ternary(s.ch == '=', token.NEQ, token.NOT)
		case '#':
			tok = token.HASH
		case '^':
			tok = token.XOR
		case '&':
			tok = s.ternary(s.ch == '&', token.LAND, token.AND)
		case '%':
			tok = s.ternary(s.ch == '=', token.REM_ASSIGN, token.REM)
		case '|':
			tok = s.ternary(s.ch == '|', token.LOR, token.OR)
		case '=':
			tok = s.ternary(s.ch == '=', token.EQL, token.ASSIGN)
		case '+':
			tok = s.ternary(s.ch == '+', token.INC, s.ternary(s.ch == '=', token.ADD_ASSIGN, token.ADD))
			if tok == token.INC {
				skipLineFeed = false
			}
		case '-':
			tok = s.ternary(s.ch == '-', token.DEC, s.ternary(s.ch == '=', token.SUB_ASSIGN, token.SUB))
			if tok == token.DEC {
				skipLineFeed = false
			}
		case '/':
			if s.ch == '/' || s.ch == '*' {
				if !s.skipLineFeed && s.ch == '/' {
					s.ch = '/'
					s.offset = offs
					s.rdOffset = offs + 1
					s.skipLineFeed = true
					tok, lit = token.LF, "\n"
					return
				}
				s.scanComment()
				goto revisit
			}
			tok = s.ternary(s.ch == '=', token.QUO_ASSIGN, token.QUO)
		case '*':
			tok = s.ternary(s.ch == '=', token.MUL_ASSIGN, token.MUL)
		case ':':
			tok = token.COLON
		case ',':
			tok = token.COMMA
		case '.':
			if s.ch != '.' {
				s.error(s.offset, "invalid symbol .")
				return
			}
			s.next()
			tok = token.RANGE
			skipLineFeed = true
		case '[':
			tok = token.LBRACK
			skipLineFeed = true
		case ']':
			tok = token.RBRACK
			skipLineFeed = false
		case '{':
			tok = token.LBRACE
			skipLineFeed = true
		case '}':
			tok = token.RBRACE
			skipLineFeed = false
		case '(':
			tok = token.LPAREN
			skipLineFeed = true
		case ')':
			tok = token.RPAREN
			skipLineFeed = false
		case '≥':
			tok = token.GEQ
		case '≤':
			tok = token.LEQ
		case '>':
			tok = s.ternary(s.ch == '=', token.GEQ, s.ternary(s.ch == '>', token.SHR, token.GTR))
		case '<':
			tok = s.ternary(s.ch == '=', token.GEQ, s.ternary(s.ch == '<', token.SHL, token.LSS))
		default:
			// defualt illegal token
			lit = string(ch)
		}
	}
	s.skipLineFeed = skipLineFeed
	return
}

func (s *Scanner) skip(test func(r rune) bool) (count int) {
	for test(s.ch) && s.ch != eof {
		s.next()
		count++
	}
	return
}

func (s *Scanner) skipWhitespace() {
	for (isSpace(s.ch) || (s.skipLineFeed && isLineFeed(s.ch))) && s.ch != eof {
		s.next()
	}
}

func (s *Scanner) scanNumber() (tok token.Token, lit string) {
	offs := s.offset
	s.next()
	s.skip(isDecimal)
	tok = token.INTEGER
	if s.ch == '.' && isDecimal(rune(s.peek())) {
		s.next()
		s.skip(isDecimal)
		tok = token.FLOAT
	}
	lit = string(s.src[offs:s.offset])
	return
}

func (s *Scanner) scanString(ch rune) (lit string, ci ast.Expr) {
	isNotSingleQuote := ch != '\''
	raw := bytes.NewBufferString("")
	builder := ast.NewStringBindingBuilder("")
	exprOffs := s.offset - 1
	if ch == 0 {
		exprOffs = s.offset
	}
loop:
	for {
		switch {
		case s.ch == ch || (ch == 0 && isArgumentStringTerminate(s.ch)):
			if isNotLineFeed(s.ch) && s.ch != eof {
				s.next()
			}
			break loop
		case s.ch == '\\':
			offs := s.offset
			s.next()
			switch s.ch {
			case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', ch:
				raw.WriteString(string(s.src[offs:s.rdOffset]))
				builder.WriteString(string(s.src[offs:s.rdOffset]))
				s.next()
			default:
				msg := "unknown escape sequence"
				if s.ch < 0 {
					msg = "escape sequence not terminated"
				}
				s.error(s.offset, msg)
				return
			}
		case isNotSingleQuote && s.ch == '$':
			offs := s.offset
			expr := s.scanExpr()
			if expr == nil {
				return
			}
			raw.WriteString(string(s.src[offs:s.offset]))
			builder.AddExpression(expr)
			continue
		}
		raw.WriteRune(s.ch)
		builder.WriteRune(s.ch)
		s.next()
	}
	lit = raw.String()
	ci = builder.Build(exprOffs, s.file)
	return
}

// special case for string interpolate
func (s *Scanner) scanExpr() ast.Expr {
	offs := s.offset
	iden, rbrackNeed := s.scanVariable()
	vexpr := ast.NewIdentExpr(offs, s.file, iden)
	if s.ch != '[' {
		if rbrackNeed {
			s.error(offs, "syntax error expect }")
			return nil
		}
		return vexpr
	} else {
		s.next()
		var index ast.Expr
		var tok token.Token
		var lit string
		for s.ch != ']' {
			switch {
			case isLetter(s.ch):
				index = s.scanExpr()
			case isDecimal(s.ch):
				numOffs := s.offset
				tok, lit = s.scanNumber()
				if tok != token.INTEGER {
					s.error(numOffs, "indexing must an integer value")
					return nil
				}
				index = ast.NewBasicLiteral(numOffs, s.file, tok, lit)
			default:
				s.error(s.offset, "expect ]")
				return nil
			}
		}
		s.next()
		if rbrackNeed {
			if s.ch != '}' {
				s.error(s.offset, "expect }")
				return nil
			}
			s.next()
		}
		return ast.NewIndexExpr(offs, s.file, index, vexpr)
	}
}

func (s *Scanner) scanVariable() (string, bool) {
	s.next()
	if s.ch == '{' {
		s.next()
		lit := s.scanIdentifier()
		if s.ch == '[' {
			return lit, true
		}
		if s.ch != '}' {
			s.error(s.offset, "expect }")
			return "", false
		}
		s.next()
		return lit, false
	} else {
		return s.scanIdentifier(), false
	}
}

func (s *Scanner) scanIdentifier() string {
	offs := s.offset
	s.next()
	for isLetter(s.ch) || isDecimal(s.ch) {
		s.next()
	}
	return string(s.src[offs:s.offset])
}

// all comment will be skip and not take into account
func (s *Scanner) scanComment() {
	offs := s.offset - 1

	if s.ch == '/' {
		s.next()
		s.skip(isNotLineFeed)
		return
	}

	/*-style comment */
	s.next()
	for s.ch >= 0 {
		ch := s.ch
		s.next()
		if ch == '*' && s.ch == '/' {
			s.next()
			return
		}
	}
	s.error(offs, "comment not terminated")
}

func (s *Scanner) scanArgument() (offs int, tok token.Token, lit string, ci ast.Expr) {
	nextLine := false
revisit:
	s.skipWhitespace()
	offs = s.offset
	switch ch := s.ch; ch {
	case '\n', eof:
		if nextLine {
			s.error(offs, "expect argument after \\")
		} else {
			tok, lit = token.LF, "\n"
			s.argumentMode = false
			s.skipLineFeed = true
		}
	case '>':
		s.next()
		tok = s.ternary(s.ch == '>', token.APPEND_TO, token.ASSIGN_TO)
		lit = tok.String()
	case '<':
		s.next()
		tok, lit = token.READ_FROM, "<"
	case '\'', '"':
		s.next()
		lit, ci = s.scanString(ch)
		tok = token.STRING
	case '\\':
		s.next()
		s.skip(isSpace)
		if s.ch == '\n' {
			s.next()
			nextLine = true
			goto revisit
		} else {
			s.error(s.offset, "expected a newline")
		}
	default:
		lit, ci = s.scanString(0)
		tok = token.STRING
	}
	return
}

func (s *Scanner) ternary(cond bool, t token.Token, f token.Token) token.Token {
	if cond {
		s.next()
		return t
	} else {
		return f
	}
}
