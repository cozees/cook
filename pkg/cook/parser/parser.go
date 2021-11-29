package parser

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"

	"github.com/cozees/cook/pkg/cook/ast"
	"github.com/cozees/cook/pkg/cook/token"
)

const maxError = 10

type Parser interface {
	Parse(file string) (ast.CookProgram, error)
}

type implParser struct {
	s        *Scanner // current scanner
	errCount int      // number of error count
	ignore   bool     // check all the syntax but not adding any instruction or target to the program

	simpleOperand bool

	// current token
	cOffs int
	cTok  token.Token
	cLit  string
	cExpr ast.Expr

	// ahead token by 1 step
	nOffs int
	nTok  token.Token
	nLit  string
	nExpr ast.Expr
}

func NewParser() Parser { return &implParser{} }

func (p *implParser) errorHandler(offset int, f *token.File, msg string) bool {
	n, l, c := f.Position(offset)
	fmt.Printf("%s:%d:%d %s\n", n, l, c, msg)
	p.errCount++
	p.toBeginStatement()
	// false terminiate as soon as error occurred otherwise true
	return p.errCount > maxError
}

func (p *implParser) expect(require token.Token) (offs int) {
	if p.cTok != require {
		p.errorHandler(p.cOffs, p.s.file, fmt.Sprintf("expect %s but got %s", require, p.cTok))
		offs = -1
	} else {
		offs = p.cOffs
	}
	p.next()
	return
}

func (p *implParser) parseIncludeDirective(files token.Files, ms map[string]*Scanner, fname string) error {
	// if file has been included
	if files.IsExisted(fname) {
		return nil
	}
	// scan only include directive
	f, err := files.AddFile(fname)
	if err != nil {
		return err
	}
	raw, err := f.ReadFile()
	if err != nil {
		return err
	}
	p.s = NewScanner(f, raw, p.errorHandler)
	ms[f.Abs()] = p.s

	for {
		offs, tok, _, _ := p.s.Scan()
		if tok == token.INCLUDE {
			offs, tok, lit, _ := p.s.Scan()
			if tok != token.STRING {
				p.errorHandler(offs, p.s.file, "include token must be a string")
			} else {
				s := p.s
				p.parseIncludeDirective(files, ms, filepath.Join(f.Dir(), lit))
				p.s = s
				offs, tok, _, _ = p.s.Scan()
				if tok != token.LF {
					n, l, c := p.s.file.Position(offs)
					return fmt.Errorf("%s:%d:%d expect linefeed but go %s", n, l, c, tok)
				}
			}
		} else {
			// we already scan include directive therefore we don't need to scan it again.
			p.s.reset(offs)
			p.s = nil
			return nil
		}
	}
}

func (p *implParser) addMainFile(mainFile string) (token.Files, map[string]*Scanner, error) {
	files := token.NewFiles()
	ms := make(map[string]*Scanner)
	err := p.parseIncludeDirective(files, ms, mainFile)
	if err != nil {
		return nil, nil, err
	}
	return files, ms, nil
}

//
func (p *implParser) init() {
	p.s.skipLineFeed = true
	p.nOffs, p.nTok, p.nLit, p.nExpr = p.s.Scan()
	p.ignore = false
}

func (p *implParser) next() {
	p.cOffs, p.cTok, p.cLit, p.cExpr = p.nOffs, p.nTok, p.nLit, p.nExpr
	if p.nTok != token.EOF {
		p.nOffs, p.nTok, p.nLit, p.nExpr = p.s.Scan()
	}
}

func (p *implParser) toBeginStatement() {
	for {
		switch p.cTok {
		case token.EOF:
			// reach the end of file
			return
		case token.AT, token.HASH:
			if p.nTok == token.IDENT {
				return
			}
		case token.IDENT:
			if p.nTok == token.AT || p.nTok == token.COLON {
				return
			}
		case token.IF, token.FOR:
			return
		}
		p.next()
	}
}

func (p *implParser) Parse(file string) (ast.CookProgram, error) {
	files, ss, err := p.addMainFile(file)
	if err != nil {
		return nil, err
	}
	cook := ast.NewCookProgram()
	totalErr := 0
	for name, f := range files {
		p.s = ss[name]
		p.parseProgram(f, cook)
		totalErr += p.errCount
		// reset error count on each file
		p.errCount = 0
	}
	if totalErr > 0 {
		return nil, fmt.Errorf("parse encouter %d error(s)", totalErr)
	}
	return cook, nil
}

//
func (p *implParser) parseProgram(f *token.File, cp ast.CookProgram) {
	var block ast.BlockStatement = cp
	p.init()
	p.next()
	inTarget := false
	for p.cTok != token.EOF {
		switch {
		case p.cTok == token.INCLUDE:
			p.errorHandler(p.cOffs, f, "include directive must define at the top of the file.")
		case p.cTok == token.IDENT:
			lit := p.cLit
			if p.nTok == token.COLON || p.nTok == token.AT {
				p.next()
				if block = p.tryParseTarget(lit, cp); block != nil || p.ignore {
					inTarget = true
					continue
				}
			} else if p.next(); p.tryParseAssignStatement(lit, block) {
				continue
			} else {
				p.errorHandler(p.cOffs, f, fmt.Sprintf("invalid token %s", p.cTok))
			}
		case inTarget:
			switch p.cTok {
			case token.FOR:
				p.parseForStatment(block)
			case token.IF:
				p.parseIfStatement(block)
			case token.AT, token.HASH:
				p.parseInvocationStatement(block)
			case token.MUL:
				t, ok := block.(ast.Target)
				if ok && t.Name() == "all" {
					p.next()
					lit := p.cLit
					if p.expect(token.IDENT) != -1 {
						if block = p.tryParseTarget(lit, cp); block != nil || p.ignore {
							inTarget = true
							continue
						} else {
							p.errorHandler(p.cOffs, f, fmt.Sprintf("invalid token %s", p.cTok))
						}
					}
				}
			}
		}
	}
}

func (p *implParser) parseSimpleStatement(block ast.BlockStatement) bool {
	for p.cTok != token.LF {
		switch p.cTok {
		case token.INCLUDE:
			p.errorHandler(p.cOffs, p.s.file, "include directive must define at the top of the file")
			return false
		case token.IDENT:
			offs, lit := p.cOffs, p.cLit
			if p.nTok == token.AT || p.nTok == token.COLON {
				// target declaraton only target is follow by @ or :
				return false
			} else if p.next(); p.cTok == token.INC || p.cTok == token.DEC {
				opOffs, op := p.cOffs, p.cTok
				if p.next(); p.expect(token.LF) != -1 {
					x := ast.NewIncDecExpr(opOffs, p.s.file, op, ast.NewIdentExpr(offs, p.s.file, lit))
					block.AddStatement(ast.NewWrapExprStatement(x))
					return true
				} else {
					return false
				}
			} else if p.tryParseAssignStatement(lit, block) {
				return true
			} else {
				p.errorHandler(p.cOffs, p.s.file, fmt.Sprintf("invalid token %s", p.cTok))
				return false
			}
		case token.FOR:
			p.parseForStatment(block)
			return true
		case token.IF:
			p.parseIfStatement(block)
			return true
		case token.AT, token.HASH:
			p.parseInvocationStatement(block)
			return true
		case token.CONTINUE, token.BREAK:
			kind := p.cTok
			p.next()
			label := ""
			if p.cTok == token.AT {
				p.next()
				label = p.cLit
				if p.expect(token.IDENT) == -1 {
					continue
				}
			}
			if p.expect(token.LF) != -1 {
				block.AddStatement(ast.NewBreakContinueStatement(kind, label))
				return true
			}
		}
	}
	p.next()
	return true
}

func (p *implParser) tryParseTarget(name string, cook ast.CookProgram) (t ast.BlockStatement) {
	var err error
	var ignore = false
revisit:
	switch p.cTok {
	case token.AT:
		p.next()
		ignore = p.cLit != runtime.GOOS
		p.expect(token.IDENT)
		goto revisit
	case token.COLON:
		if !ignore {
			if t, err = cook.AddTarget(name); err != nil {
				p.errorHandler(p.cOffs, p.s.file, err.Error())
			}
		}
	default:
		ignore = false
	}
	p.next()
	p.ignore = ignore
	return
}

func (p *implParser) tryParseAssignStatement(varn string, block ast.BlockStatement) bool {
	switch p.cTok {
	case token.ASSIGN:
		if p.nTok == token.AT || p.nTok == token.HASH {
			op := p.cTok
			p.next()
			if _, _, expr := p.parseInvocationExpr(false); expr != nil {
				block.AddStatement(ast.NewAssignStatement(varn, op, expr))
			} else if !p.ignore {
				return false
			}
			return true
		}
		fallthrough
	case token.ADD_ASSIGN, token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN, token.REM_ASSIGN:
		op := p.cTok
		if expr := p.parseBinaryExpr(token.LowestPrec + 1); expr != nil {
			if !p.ignore {
				block.AddStatement(ast.NewAssignStatement(varn, op, expr))
			}
			p.expect(token.LF)
			return true
		} else if p.ignore {
			return true
		}
	}
	return false
}

func (p *implParser) parseInvocationStatement(block ast.BlockStatement) {
	redirectTo, isAppend, x := p.parseInvocationExpr(true)
	if !p.ignore {
		if len(redirectTo) > 0 {
			block.AddStatement(ast.NewRedirectToStatement(isAppend, x, redirectTo))
		} else {
			block.AddStatement(ast.NewWrapExprStatement(x))
		}
	}
}

func (p *implParser) parseInvocationExpr(canHasAssignTo bool) (redirectTo []ast.Expr, isAppend bool, x ast.Expr) {
	offs := p.cOffs
	kind := p.cTok
	p.next()
	if p.cTok != token.IDENT {
		p.errorHandler(p.cOffs, p.s.file, "expect identifier")
		return
	}
	name := p.cLit
	readFromOffs := -1
	redirectToOffs := -1
	p.next()
	var args []ast.Expr
	if !p.ignore {
		args = make([]ast.Expr, 0)
		redirectTo = make([]ast.Expr, 0)
	}
	for p.cTok != token.LF {
		switch p.cTok {
		case token.STRING, token.IDENT:
			if !p.ignore {
				x := p.cExpr
				if x == nil {
					x = ast.NewBasicLiteral(p.cOffs, p.s.file, token.STRING, p.cLit)
				}
				if readFromOffs != -1 {
					args = append(args, ast.NewReadFromExpr(readFromOffs, p.s.file, x))
					readFromOffs = -1
				} else if redirectToOffs != -1 {
					redirectTo = append(redirectTo, x)
					if p.nTok != token.LF && p.nTok != token.STRING && p.nTok != token.IDENT {
						p.next()
						p.errorHandler(p.cOffs, p.s.file, fmt.Sprintf("invalid token %s", p.cTok))
						return nil, false, nil
					}
				} else {
					args = append(args, x)
				}
			}
		case token.APPEND_TO:
			isAppend = true
			redirectToOffs = p.cOffs
		case token.ASSIGN_TO:
			redirectToOffs = p.cOffs
		case token.READ_FROM:
			readFromOffs = p.cOffs
		default:
			p.errorHandler(p.cOffs, p.s.file, "calling target or function required string or identifier token")
			return nil, false, nil
		}
		p.next()
	}
	if p.expect(token.LF) != -1 {
		if !canHasAssignTo && len(redirectTo) > 0 {
			p.errorHandler(offs, p.s.file, "assign to or redirect syntax is not allow here")
		} else if !p.ignore {
			return redirectTo, isAppend, ast.NewCallExpr(offs, p.s.file, name, kind, args)
		}
	}
	return nil, false, nil
}

func (p *implParser) parseForStatment(block ast.BlockStatement) {
	p.next()
	label := ""
	if p.cTok == token.AT {
		// special case to reset track of token
		p.s.prevTokens[0] = token.ILLEGAL
		p.s.prevTokens[1] = token.ILLEGAL
		p.next()
		label = p.cLit
		if p.expect(token.IDENT) == -1 {
			return
		}
	}
	offs, lit := p.cOffs, p.cLit
	var ranges []ast.Expr
	var fob ast.BlockStatement
	switch p.cTok {
	case token.IDENT:
		p.next()
		switch p.cTok {
		case token.COMMA:
			// for k, v on list or map
			p.next()
			vOffs, vLit := p.cOffs, p.cLit
			if p.expect(token.IDENT) == -1 || p.expect(token.IN) == -1 {
				return
			}
			var expr ast.Expr
			switch p.cTok {
			case token.LBRACE:
				// literal map
				expr = p.parseOperand()
			case token.LBRACK:
				// literal array
				expr = p.parseOperand()
			case token.IDENT:
				expr = ast.NewIdentExpr(p.cOffs, p.s.file, p.cLit)
				p.next()
			}
			if p.expect(token.LBRACE) != -1 {
				i1 := ast.NewIdentExpr(offs, p.s.file, lit)
				i2 := ast.NewIdentExpr(vOffs, p.s.file, vLit)
				fob = ast.NewForLMStatement(label, i1, i2, expr)
			}
		case token.IN:
			p.next()
		readAgain:
			if p.cTok == token.INTEGER {
				ranges = append(ranges, ast.NewBasicLiteral(p.cOffs, p.s.file, p.cTok, p.cLit))
			} else if p.cTok == token.IDENT {
				ranges = append(ranges, ast.NewIdentExpr(p.cOffs, p.s.file, p.cLit))
			} else {
				p.errorHandler(p.cOffs, p.s.file, "expect identifier or integer")
				return
			}
			p.next()
			if len(ranges) == 2 {
				if p.expect(token.LBRACE) != -1 {
					i1 := ast.NewIdentExpr(offs, p.s.file, lit)
					fob = ast.NewForRangeStatement(label, i1, ranges)
				}
			} else if p.expect(token.RANGE) != -1 {
				goto readAgain
			}
		default:
			goto invalidToken
		}
	case token.LBRACE:
		p.next()
		fob = ast.NewForStatement(label)
	default:
		goto invalidToken
	}
	// parse for body
	for p.cTok != token.RBRACE {
		if !p.parseSimpleStatement(fob) {
			break
		}
	}
	if p.expect(token.RBRACE) != -1 && p.expect(token.LF) != -1 {
		block.AddStatement(fob)
	}
	return

invalidToken:
	p.errorHandler(p.cOffs, p.s.file, fmt.Sprintf("unexpected token %s", p.cTok))
}

func (p *implParser) parseIfStatement(block ast.BlockStatement) {
	expr := p.parseBinaryExpr(token.LowestPrec + 1)
	if p.expect(token.LBRACE) != -1 {
		ifStmt := ast.NewIfStatement(expr)
		for p.cTok != token.RBRACE {
			if !p.parseSimpleStatement(ifStmt) {
				break
			}
		}
		if p.expect(token.RBRACE) != -1 {
			block.AddStatement(ifStmt)
		}

		if p.cTok == token.ELSE {
			p.next()
			elStmt := ast.NewElseStatement()
			if p.cTok == token.IF {
				p.parseIfStatement(elStmt)
			} else if p.cTok == token.LBRACE {
				p.next()
				for p.cTok != token.RBRACE {
					if !p.parseSimpleStatement(elStmt) {
						break
					}
				}
				p.expect(token.RBRACE)
				p.expect(token.LF)
			} else {
				p.errorHandler(p.cOffs, p.s.file, "expected {")
				return
			}
			ifStmt.Else = elStmt
		} else {
			p.expect(token.LF)
		}
	}
}

func (p *implParser) parseBinaryExpr(priority int) (x ast.Expr) {
	p.next()
	x = p.parseUnaryExpr()

	for {
		op, oprec := p.cTok, p.cTok.Precedence()
		if oprec < priority {
			return
		}
		// special case for is, ternary and fallback expression
		if op == token.QES {
			// ternary case or short if
			x = p.parseTernaryExpr(x)
		} else if op == token.DQES {
			// fallback expression
			x = p.parseFallbackExpr(x)
		} else if op == token.IS {
			x = p.parseIsExpr(x)
		} else {
			offs := p.cOffs
			y := p.parseBinaryExpr(oprec + 1)
			if !p.ignore {
				x = ast.NewBinaryExpr(offs, p.s.file, op, x, y)
			}
		}
	}
}

func (p *implParser) parseTernaryExpr(cond ast.Expr) (x ast.Expr) {
	offs := p.cOffs
	t := p.parseBinaryExpr(token.LowestPrec + 1)
	if p.cTok == token.COLON {
		f := p.parseBinaryExpr(token.LowestPrec + 1)
		x = ast.NewTernaryExpr(offs, p.s.file, cond, t, f)
	} else {
		p.errorHandler(p.cOffs, p.s.file, fmt.Sprintf("expect : but got %s", p.cLit))
		p.next()
	}
	return
}

func (p *implParser) parseFallbackExpr(primary ast.Expr) (x ast.Expr) {
	offs := p.cOffs
	fx := p.parseBinaryExpr(token.LowestPrec + 1)
	return ast.NewFallbackExpr(offs, p.s.file, primary, fx)
}

func (p *implParser) parseIsExpr(t ast.Expr) (x ast.Expr) {
	offs := p.cOffs
	p.next()
	ttok := make([]token.Token, 0)
	for {
		if p.cTok.Type() > 0 {
			ttok = append(ttok, p.cTok)
		} else if p.cTok != token.OR {
			break
		}
		p.next()
	}
	if len(ttok) == 0 {
		p.errorHandler(offs, p.s.file, "invalid type check expression")
	} else {
		x = ast.NewIsTypeExpr(offs, p.s.file, t, ttok...)
	}
	return
}

func (p *implParser) parseUnaryExpr() (x ast.Expr) {
	switch p.cTok {
	case token.ADD, token.SUB, token.NOT, token.XOR:
		offs, op := p.cOffs, p.cTok
		p.next()
		y := p.parseOperand()
		if !p.ignore {
			x = ast.NewUnaryExpr(offs, p.s.file, op, y)
		}
	case token.SIZEOF:
		offs := p.cOffs
		p.next()
		y := p.parseOperand()
		if !p.ignore {
			x = ast.NewSizeOfExpr(offs, p.s.file, y)
		}
	case token.VAR:
		offs := p.cOffs
		p.next()
		lit := p.cLit
		if p.expect(token.INTEGER) != -1 {
			x = ast.NewIdentExpr(offs, p.s.file, lit)
		}
	case token.TINTEGER:
		x = p.parseTypeCaseExpr(reflect.Int64)
	case token.TFLOAT:
		x = p.parseTypeCaseExpr(reflect.Float64)
	case token.TSTRING:
		x = p.parseTypeCaseExpr(reflect.String)
	case token.TBOOLEAN:
		x = p.parseTypeCaseExpr(reflect.Bool)
	default:
		x = p.parseOperand()
	}
	return
}

func (p *implParser) parseTypeCaseExpr(to reflect.Kind) (x ast.Expr) {
	offs := p.cOffs
	p.next()
	if p.expect(token.LPAREN) != -1 {
		y := p.parseOperand()
		if p.expect(token.RPAREN) != -1 && !p.ignore {
			x = ast.NewTypeCastExpr(offs, p.s.file, y, to)
		}
	}
	return
}

func (p *implParser) parseOperand() (x ast.Expr) {
	// ensure that simpleOperand reset after done parsing
	defer func() { p.simpleOperand = false }()

	switch p.cTok {
	case token.IDENT:
		offs := p.cOffs
		if !p.ignore {
			x = ast.NewIdentExpr(offs, p.s.file, p.cLit)
		}
		p.next()
		if p.cTok == token.LBRACK {
			// indexing expr
			p.simpleOperand = true
			index := p.parseBinaryExpr(token.LowestPrec + 1)
			if index == nil {
				return
			}
			if p.expect(token.RBRACK) != -1 && !p.ignore {
				x = ast.NewIndexExpr(offs, p.s.file, index, x)
			}
		}
		return

	case token.INTEGER, token.FLOAT, token.STRING, token.BOOLEAN:
		if !p.ignore {
			if p.cExpr != nil && p.cTok == token.STRING {
				x = p.cExpr
			} else {
				x = ast.NewBasicLiteral(p.cOffs, p.s.file, p.cTok, p.cLit)
			}
		}
		p.next()
		return

	case token.LPAREN:
		lparen := p.cOffs
		inx := p.parseBinaryExpr(token.LowestPrec + 1) // types may be parenthesized: (some type)
		if p.expect(token.RPAREN) != -1 {
			if !p.ignore {
				x = ast.NewParanExpr(lparen, p.s.file, inx)
			}
		}
		return
	default:
		switch {
		case !p.simpleOperand && p.cTok == token.LBRACE:
			return p.parseMapLiteral()
		case !p.simpleOperand && p.cTok == token.LBRACK:
			return p.parseListLiteral()
		default:
			p.errorHandler(p.cOffs, p.s.file, fmt.Sprintf("invalid token %s", p.cTok))
		}
	}
	return nil
}

func (p *implParser) parseListLiteral() (x ast.Expr) {
	offs := p.s.offset
	p.next()
	var values []ast.Expr
	if !p.ignore {
		if p.cTok != token.RBRACK {
			values = append(values, p.parseOperand())
		} else {
			values = make([]ast.Expr, 0)
		}
	} else if p.cTok != token.RBRACK {
		p.parseOperand()
	}
loop:
	for {
		switch p.cTok {
		case token.RBRACK, token.EOF:
			break loop
		case token.COMMA:
			p.next()
			if p.cTok == token.RBRACK {
				break loop
			}
			y := p.parseOperand()
			if !p.ignore {
				values = append(values, y)
			}
		}
	}
	if p.expect(token.RBRACK) != -1 && !p.ignore {
		x = ast.NewListLiteral(offs, p.s.file, values)
	}
	return
}

func (p *implParser) parseMapLiteral() (x ast.Expr) {
	offs := p.s.offset
	p.next()
	var keys []ast.Expr
	var values []ast.Expr
	if !p.ignore {
		keys = make([]ast.Expr, 0)
		values = make([]ast.Expr, 0)
	}
	if p.cTok != token.RBRACE {
	loop:
		for {
			k := p.parseOperand()
			if !p.ignore {
				keys = append(keys, k)
			}
			if p.expect(token.COLON) != -1 {
				v := p.parseOperand()
				if !p.ignore {
					values = append(values, v)
				}
			} else {
				return
			}
			switch p.cTok {
			case token.RBRACE, token.EOF:
				break loop
			case token.COMMA:
				p.next()
				if p.cTok == token.RBRACE {
					break loop
				}
			}
		}
	}
	if p.expect(token.RBRACE) != -1 && !p.ignore {
		x = ast.NewMapLiteral(offs, p.s.file, keys, values)
	}
	return
}
