package ast

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/cozees/cook/pkg/cook/token"
)

//
type CookProgram interface {
	BlockStatement
	AddTarget(name string) (Target, error)
	Execute(args map[string]interface{})
	ExecuteWithTarget(args map[string]interface{}, names ...string)
}

type implCookProgram struct {
	*implBlock
	*implContext
}

func NewCookProgram() CookProgram {
	return &implCookProgram{
		implBlock: &implBlock{},
		implContext: &implContext{
			gvar:          make(map[string]interface{}),
			targetsByName: make(map[string]int),
			restrictVar:   make(map[string]reflect.Kind),
		},
	}
}

func (icp *implCookProgram) AddTarget(name string) (Target, error) {
	if _, ok := icp.targetsByName[name]; !ok {
		target := NewTarget(name)
		icp.targetsByName[name] = len(icp.targets)
		icp.targets = append(icp.targets, target)
		return target, nil
	}
	return nil, fmt.Errorf("target %s is already existed", name)
}

func (icp *implCookProgram) Execute(args map[string]interface{}) {
	if icp.getTarget("all") == nil {
		icp.onError(fmt.Errorf("missing target or all target is not defined"))
		return
	}
	icp.ExecuteWithTarget(args, "all")
}

func (icp *implCookProgram) ExecuteWithTarget(args map[string]interface{}, names ...string) {
	// add argument to global variable with suffix g
	for name, value := range args {
		icp.gvar[name] = value
	}
	// execute any instruction before execute target
	// this instruction is defined before any target
	if icp.evaluate(icp.implContext); icp.hasCanceled() {
		return
	}
	// execute initialize target if defined
	if tg := icp.getTarget("initialize"); tg != nil {
		tg.run(icp.implContext, nil)
	}
	// execute finalize target if defined
	if tg := icp.getTarget("finalize"); tg != nil {
		defer func() {
			recover()
			tg.run(icp.implContext, nil)
		}()
	}
	// execute target by name
	for _, tgName := range names {
		if tgName == "initialize" || tgName == "finalize" {
			continue
		}
		if tg := icp.getTarget(tgName); tg != nil {
			if tg.run(icp.implContext, nil); icp.hasCanceled() {
				return
			}
			if tgName == "all" {
				return
			}
		}
	}
}

func (icp *implCookProgram) String(indent int) string {
	buffer := bytes.NewBufferString("")
	for _, ins := range icp.statements {
		buffer.WriteString(ins.String(indent))
	}
	// 3 specials target first
	if tg := icp.getTarget("initialize"); tg != nil {
		buffer.WriteString(tg.String(indent))
	}
	if tg := icp.getTarget("finalize"); tg != nil {
		buffer.WriteString(tg.String(indent))
	}
	if tg := icp.getTarget("all"); tg != nil {
		buffer.WriteString(tg.String(indent))
	}
	// the rest last
	for _, tg := range icp.getTargets() {
		tgName := tg.Name()
		if tgName == "initialize" || tgName == "finalize" || tgName == "all" {
			continue
		}
		buffer.WriteString(tg.String(indent))
	}
	return buffer.String()
}

//
type BlockStatement interface {
	CookStatement
	AddStatement(ins CookStatement)
}

type implBlock struct {
	statements []CookStatement
}

func (ib *implBlock) AddStatement(ins CookStatement) {
	ib.statements = append(ib.statements, ins)
}

func (ib *implBlock) evaluate(ctx cookContext) {
	for _, ins := range ib.statements {
		if ctx.hasCanceled() {
			return
		}
		ins.evaluate(ctx)
		if fcc, ok := ctx.(*implForContext); ok {
			if fcc.shouldBreak(fcc.currentLoop()) {
				break
			}
			fcc.reset()
		}
	}
}

func (ib *implBlock) String(indent int) string {
	buffer := bytes.NewBufferString(" {\n")
	for _, ins := range ib.statements {
		buffer.WriteString(ins.String(indent))
	}
	buffer.WriteString(strings.Repeat(" ", indent-4))
	buffer.WriteRune('}')
	return buffer.String()
}

//
type CookStatement interface {
	evaluate(ctx cookContext)
	String(indent int) string
}

type exprWrapperStatement struct {
	expr Expr
}

func NewWrapExprStatement(expr Expr) CookStatement { return &exprWrapperStatement{expr: expr} }

func (cs *exprWrapperStatement) evaluate(ctx cookContext) { cs.expr.evaluate(ctx) }

func (cs *exprWrapperStatement) String(indent int) string {
	return strings.Repeat(" ", indent) + cs.expr.String() + "\n"
}

//
type assingStatement struct {
	variable string
	op       token.Token
	value    Expr
}

func NewAssignStatement(variable string, op token.Token, value Expr) CookStatement {
	return &assingStatement{variable: variable, op: op, value: value}
}

func (ai *assingStatement) evaluate(ctx cookContext) {
	if !ctx.hasCanceled() {
		if ce, ok := ai.value.(*CallExpr); ok {
			ce.OutputAsResult = true
		}
		i, ik := ai.value.evaluate(ctx)
		switch {
		case ai.op == token.ASSIGN:
			if err := ctx.setVariable(ai.variable, i); err != nil {
				ctx.onError(err)
			}
		default:
			v, vk, _ := ctx.getVariable(ai.variable)
			if v == nil {
				ctx.onError(fmt.Errorf("variable %s has not been modified", ai.variable))
			} else {
				switch ai.op {
				case token.ADD_ASSIGN:
					sum, _ := addOperator(ctx, v, i, vk, ik)
					if err := ctx.setVariable(ai.variable, sum); err != nil {
						ctx.onError(err)
					}
				case token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN, token.REM_ASSIGN:
					r, _ := numOperator(ctx, ai.op-5, v, i, vk, ik)
					if err := ctx.setVariable(ai.variable, r); err != nil {
						ctx.onError(err)
					}
				default:
					panic("illegal state parser. Parser should verify the permitted operator already")
				}
			}
		}
	}
}

func (ai *assingStatement) String(indent int) string {
	buffer := bytes.NewBufferString(strings.Repeat(" ", indent))
	buffer.WriteString(ai.variable)
	buffer.WriteRune(' ')
	buffer.WriteString(ai.op.String())
	buffer.WriteRune(' ')
	buffer.WriteString(ai.value.String())
	buffer.WriteRune('\n')
	return buffer.String()
}

type redirectToStatement struct {
	Files    []Expr
	Call     Expr
	IsAppend bool
}

func NewRedirectToStatement(isAppend bool, call Expr, files []Expr) CookStatement {
	return &redirectToStatement{Files: files, Call: call, IsAppend: isAppend}
}

func (ats *redirectToStatement) evaluate(ctx cookContext) {
	ats.Call.(*CallExpr).OutputAsResult = true
	v, vk := ats.Call.evaluate(ctx)
	if vk != reflect.Invalid {
		if r := convertToReadCloser(ctx, v, vk); r != nil {
			defer r.Close()
			var writers []io.Writer
			addWriter := func(fp interface{}, k reflect.Kind) io.WriteCloser {
				if k != reflect.String {
					ctx.onError(fmt.Errorf("value %v type %s is not a file path", v, vk))
					return nil
				} else {
					fpath := fp.(string)
					flag := os.O_CREATE | os.O_WRONLY
					if ats.IsAppend {
						flag |= os.O_APPEND
					} else {
						flag |= os.O_TRUNC
					}
					f, err := os.OpenFile(fpath, flag, 0700)
					if err != nil {
						ctx.onError(err)
						return nil
					}
					return f
				}
			}
			// create writer
			for _, f := range ats.Files {
				fp, k := f.evaluate(ctx)
				if k == reflect.Array {
					for _, ifv := range fp.([]interface{}) {
						w := addWriter(ifv, reflect.ValueOf(ifv).Kind())
						if w == nil {
							return
						}
						defer w.Close()
						writers = append(writers, w)
					}
				} else if w := addWriter(fp, k); w == nil {
					return
				} else {
					defer w.Close()
					writers = append(writers, w)
				}
			}
			//
			if _, err := io.Copy(io.MultiWriter(writers...), r); err != nil {
				ctx.onError(err)
			}
		}
	} else {
		fmt.Println("Warning: \"AssignTo\" is used on a function that might not return any value.")
	}
}

func (ats *redirectToStatement) String(indent int) string {
	buffer := bytes.NewBufferString(strings.Repeat(" ", indent))
	buffer.WriteString(ats.Call.String())
	if ats.IsAppend {
		buffer.WriteString(" >> ")
	} else {
		buffer.WriteString(" > ")
	}
	for _, f := range ats.Files {
		if _, ok := f.(*Ident); ok {
			buffer.WriteRune('$')
		}
		buffer.WriteString(f.String())
		buffer.WriteRune(' ')
	}
	buffer.Truncate(buffer.Len() - 1)
	buffer.WriteRune('\n')
	return buffer.String()
}

//
type ForStatement struct {
	*implBlock
	I1    Expr // index or key
	I2    Expr // value of array index or map
	Range []Expr
	Opr   Expr
	Label string
}

func NewForStatement(label string) BlockStatement {
	return &ForStatement{
		implBlock: &implBlock{},
		Label:     label,
	}
}

func NewForRangeStatement(label string, i1 Expr, range_ []Expr) BlockStatement {
	return &ForStatement{
		implBlock: &implBlock{},
		I1:        i1,
		Range:     range_,
		Label:     label,
	}
}

func NewForLMStatement(label string, i1, i2 Expr, opr Expr) BlockStatement {
	return &ForStatement{
		implBlock: &implBlock{},
		I1:        i1,
		I2:        i2,
		Opr:       opr,
		Label:     label,
	}
}

func (fs *ForStatement) blockEvaluate(index int, ctx forCookContext) {
	for _, ins := range fs.statements {
		if ctx.hasCanceled() {
			return
		}
		ins.evaluate(ctx)
		if ctx.shouldBreak(index) {
			return
		}
		ctx.reset()
	}
}

func (fs *ForStatement) evaluate(ctx cookContext) {
	fcc, err := isTargetContext(ctx)
	if err != nil {
		ctx.onError(fmt.Errorf("for loop is %w", err))
		return
	}
	// register loop in the context, usually use for break, if label is provided,
	// break or continue can be specify label as addition
	forInd := fcc.register(fs.Label)
	defer fcc.unregister(forInd)
	// add local scope for variable available in for index or index/value or key/value
	localCtx := ctx.addScope()
	defer ctx.popScope()
	// Opr is array or map
	eval := func(ctx forCookContext) bool {
		fs.blockEvaluate(forInd, ctx)
		return ctx.shouldBreak(forInd)
	}
	if fs.Opr != nil {
		i1 := fs.I1.(*Ident).Name
		i2 := ""
		if fs.I2 != nil {
			i2 = fs.I2.(*Ident).Name
		}
		v, vk := fs.Opr.evaluate(ctx)
		switch vk {
		case reflect.Array, reflect.Slice:
			fcc.enterLoop(forInd)
			defer fcc.exitLoop(forInd)
			for i, iv := range v.([]interface{}) {
				localCtx[i1] = int64(i)
				if i2 != "" {
					localCtx[i2] = iv
				}
				if eval(fcc) {
					break
				}
				fcc.reset()
			}
		case reflect.Map:
			fcc.enterLoop(forInd)
			defer fcc.exitLoop(forInd)
			for k, kv := range v.(map[interface{}]interface{}) {
				localCtx[i1] = k
				if i2 != "" {
					localCtx[i2] = kv
				}
				if eval(fcc) {
					break
				}
				fcc.reset()
			}
		default:
			ctx.onError(fmt.Errorf("cannot loop value %v only map or list is allowed", v))
		}
	} else if fs.Range != nil { // indexing
		l, lk := fs.Range[0].evaluate(ctx)
		g, gk := fs.Range[1].evaluate(ctx)
		if lk != reflect.Int64 || gk != reflect.Int64 {
			ctx.onError(fmt.Errorf("unsupport range value %v..%v, must be integer", l, g))
		}
		i := fs.I1.(*Ident).Name
		i1, i2 := l.(int64), g.(int64)
		ctx.restrictVariable(i, reflect.Int64)
		defer ctx.restrictVariable(i, reflect.Invalid)
		fnb := func(ind int64) (int64, bool) {
			localCtx[i] = ind
			if eval(fcc) {
				return ind, true
			}
			fcc.reset()
			if localCtx[i] != ind {
				if ii, ok := localCtx[i].(int64); ok {
					return ii, false
				} else {
					ctx.position(fs.I1.(*Ident).baseExpr)
					ctx.onError(fmt.Errorf("modify index variable \"%s\" must maintain it type integer", fs.I1.String()))
					return 0, true
				}
			}
			return ind, false
		}
		notOk := false
		fcc.enterLoop(forInd)
		defer fcc.exitLoop(forInd)
		if i1 > i2 {
			for ind := i1; ind >= i2; ind-- {
				if ind, notOk = fnb(ind); notOk {
					break
				}
			}
		} else {
			for ind := i1; ind <= i2; ind++ {
				if ind, notOk = fnb(ind); notOk {
					break
				}
			}
		}
	} else { // loop until break or continue target specified
		fcc.enterLoop(forInd)
		defer fcc.exitLoop(forInd)
		for !eval(fcc) {
		}
	}
}

func (fs *ForStatement) String(indent int) string {
	buffer := bytes.NewBufferString(strings.Repeat(" ", indent))
	buffer.WriteString("for")
	if fs.Label != "" {
		buffer.WriteRune('@')
		buffer.WriteString(fs.Label)
	}
	buffer.WriteRune(' ')
	buffer.WriteString(fs.I1.String())
	if fs.I2 != nil {
		buffer.WriteString(", ")
		buffer.WriteString(fs.I2.String())
	}
	buffer.WriteString(" in ")
	if len(fs.Range) == 2 {
		buffer.WriteString(fs.Range[0].String())
		buffer.WriteString("..")
		buffer.WriteString(fs.Range[1].String())
	} else {
		buffer.WriteString(fs.Opr.String())
	}
	buffer.WriteString(fs.implBlock.String(indent + 4))
	buffer.WriteRune('\n')
	return buffer.String()
}

//
type IfStatement struct {
	*implBlock
	Cond Expr
	Else CookStatement
}

type ElseStatement struct {
	*implBlock
}

func NewIfStatement(cond Expr) *IfStatement {
	return &IfStatement{
		implBlock: &implBlock{},
		Cond:      cond,
	}
}

func NewElseStatement() BlockStatement { return &ElseStatement{implBlock: &implBlock{}} }

func (is *IfStatement) evaluate(ctx cookContext) {
	v, vk := is.Cond.evaluate(ctx)
	if vk != reflect.Bool {
		ctx.onError(fmt.Errorf("is not a boolean expression"))
	} else if v.(bool) {
		is.implBlock.evaluate(ctx)
	} else if is.Else != nil {
		is.Else.evaluate(ctx)
	}
}

func (is *IfStatement) String(indent int) string {
	buffer := bytes.NewBufferString(strings.Repeat(" ", indent))
	buffer.WriteString("if ")
	buffer.WriteString(is.Cond.String())
	buffer.WriteString(is.implBlock.String(indent + 4))
	if is.Else != nil {
		buffer.WriteString(" else ")
		buffer.WriteString(is.Else.String(indent + 4))
	}
	buffer.WriteRune('\n')
	return buffer.String()
}

//
type bcLoopStatement struct {
	Kind  token.Token
	Label string
}

func NewBreakContinueStatement(kind token.Token, label string) CookStatement {
	return &bcLoopStatement{
		Kind:  kind,
		Label: label,
	}
}

func (bcls *bcLoopStatement) evaluate(ctx cookContext) {
	fcc, err := isTargetContext(ctx)
	if err != nil {
		ctx.onError(fmt.Errorf("for loop is %w", err))
		return
	}
	switch bcls.Kind {
	case token.BREAK:
		if err = fcc.breakWith(bcls.Label); err != nil {
			ctx.onError(err)
		}
	case token.CONTINUE:
		if err = fcc.continueWith(bcls.Label); err != nil {
			ctx.onError(err)
		}
	default:
		panic("illegal state parser, parser create this statement only with break or continue")
	}
}

func (bcls *bcLoopStatement) String(indent int) string {
	buffer := bytes.NewBufferString(strings.Repeat(" ", indent))
	buffer.WriteString(bcls.Kind.String())
	if bcls.Label != "" {
		buffer.WriteRune(' ')
		buffer.WriteString(bcls.Label)
		buffer.WriteRune('\n')
	}
	return buffer.String()
}
