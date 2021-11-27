package ast

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/cozees/cook/pkg/cook/token"
	"github.com/cozees/cook/pkg/runtime/function"
)

type cookContext interface {
	// add local scope for the tail of scope list.
	addScope() map[string]interface{}
	// remove last scope from the list
	popScope()

	getTargets() []Target
	getTarget(name string) Target
	getFunction(name string) function.Function

	getVariable(name string) (value interface{}, kind reflect.Kind, fromEnv bool)
	setVariable(name string, val interface{})

	position(be *baseExpr)
	onError(err error)
	hasCanceled() bool

	// tell context to record the error instead of printing out
	// and ignore it so no cancel is occurred
	recordFailure()
	hasFailure() bool
}

type forCookContext interface {
	cookContext
	breakWith(id string) error
	continueWith(id string) error
	breakAt() int
	continueAt() int
	shouldBreakingOf(index int) bool
	shouldContinueOf(index int) bool
	register(id string) (index int)
	unregister(index int)
}

//
type CookProgram interface {
	BlockStatement
	AddTarget(name string) (Target, error)
	Execute(args map[string]interface{})
	ExecuteWithTarget(name []string, args map[string]interface{})
}

type implCookProgram struct {
	*implBlock

	locals        []map[string]interface{}
	gvar          map[string]interface{}
	targets       []Target
	targetsByName map[string]int
	isCanceled    bool
	curBaseExpr   *baseExpr

	// record last error
	lastError   error
	recordError bool
}

func NewCookProgram() CookProgram {
	return &implCookProgram{
		implBlock:     &implBlock{},
		gvar:          make(map[string]interface{}),
		targetsByName: make(map[string]int),
	}
}

func (icp *implCookProgram) recordFailure() { icp.recordError = true }
func (icp *implCookProgram) hasFailure() bool {
	defer func() {
		icp.recordError = false
		icp.lastError = nil
	}()
	return icp.lastError != nil
}

func (icp *implCookProgram) addScope() map[string]interface{} {
	store := make(map[string]interface{})
	icp.locals = append(icp.locals, store)
	return store
}

func (icp *implCookProgram) popScope()            { icp.locals = icp.locals[0 : len(icp.locals)-1] }
func (icp *implCookProgram) getTargets() []Target { return icp.targets }

func (icp *implCookProgram) getTarget(name string) Target {
	if index, ok := icp.targetsByName[name]; !ok {
		return nil
	} else {
		return icp.targets[index]
	}
}

func (icp *implCookProgram) getFunction(name string) function.Function {
	return function.GetFunction(name)
}

func (icp *implCookProgram) getVariable(name string) (value interface{}, kind reflect.Kind, fromEnv bool) {
	for i := len(icp.locals) - 1; i >= 0; i-- {
		if v, ok := icp.locals[i][name]; ok {
			value, kind, fromEnv = v, reflect.ValueOf(v).Kind(), false
			return
		}
	}
	if v, ok := icp.gvar[name]; ok {
		value, kind, fromEnv = v, reflect.ValueOf(v).Kind(), false
	} else if v := os.Getenv(name); v != "" {
		value, kind, fromEnv = v, reflect.String, true
	}
	return
}

func (icp *implCookProgram) setVariable(name string, val interface{}) {
	last := len(icp.locals) - 1
	for i := last; i >= 0; i-- {
		if _, ok := icp.locals[i][name]; ok {
			icp.locals[i][name] = val
			return
		}
	}
	// if exist in global replace the value otherwise create in local scope only.
	if _, ok := icp.gvar[name]; ok || last < 0 {
		icp.gvar[name] = val
	} else {
		icp.locals[last][name] = val
	}
}

func (icp *implCookProgram) hasCanceled() bool     { return icp.isCanceled }
func (icp *implCookProgram) position(be *baseExpr) { icp.curBaseExpr = be }

func (icp *implCookProgram) onError(err error) {
	if icp.recordError {
		icp.lastError = err
	} else {
		fmt.Fprintf(os.Stderr, "%s %s\n", icp.curBaseExpr.PosInfo(), err.Error())
		icp.isCanceled = true
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
	icp.ExecuteWithTarget([]string{"all"}, args)
}

func (icp *implCookProgram) ExecuteWithTarget(names []string, args map[string]interface{}) {
	// add argument to global variable with suffix g
	for name, value := range args {
		icp.gvar[name] = value
	}
	// execute any instruction before execute target
	// this instruction is defined before any target
	if icp.evaluate(icp); icp.hasCanceled() {
		return
	}
	// execute initialize target if defined
	if tg := icp.getTarget("initialize"); tg != nil {
		tg.run(icp, nil)
	}
	// execute finalize target if defined
	if tg := icp.getTarget("finalize"); tg != nil {
		defer func() {
			recover()
			tg.run(icp, nil)
		}()
	}
	// execute target by name
	for _, tgName := range names {
		if tgName == "initialize" || tgName == "finalize" {
			continue
		}
		if tg := icp.getTarget(tgName); tg != nil {
			if tg.run(icp, nil); icp.hasCanceled() {
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
			ctx.setVariable(ai.variable, i)
		default:
			v, vk, _ := ctx.getVariable(ai.variable)
			if v == nil {
				ctx.onError(fmt.Errorf("variable %s has not been modified", ai.variable))
			} else {
				switch ai.op {
				case token.ADD_ASSIGN:
					sum, _ := addOperator(ctx, v, i, vk, ik)
					ctx.setVariable(ai.variable, sum)
				case token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN, token.REM_ASSIGN:
					r, _ := numOperator(ctx, ai.op-5, v, i, vk, ik)
					ctx.setVariable(ai.variable, r)
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
					}
					f, err := os.OpenFile(fpath, flag, 0700)
					if err != nil {
						ctx.onError(err)
						return nil
					}
					return f
				}
			}
			// create reader
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

func (ib *implBlock) blockEvaluate(index int, ctx forCookContext) {
	for _, ins := range ib.statements {
		if ctx.hasCanceled() {
			return
		}
		ins.evaluate(ctx)
		if ctx.shouldBreakingOf(index) || ctx.shouldContinueOf(index) {
			return
		}
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
		return ctx.shouldBreakingOf(forInd)
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
			for i, iv := range v.([]interface{}) {
				localCtx[i1] = int64(i)
				if i2 != "" {
					localCtx[i2] = iv
				}
				if eval(fcc) {
					break
				}
			}
		case reflect.Map:
			for k, kv := range v.(map[interface{}]interface{}) {
				localCtx[i1] = k
				if i2 != "" {
					localCtx[i2] = kv
				}
				if eval(fcc) {
					break
				}
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
		if i1 > i2 {
			for ind := i1; ind >= i2; ind-- {
				localCtx[i] = ind
				if eval(fcc) {
					break
				}
			}
		} else {
			for ind := i1; ind <= i2; ind++ {
				localCtx[i] = ind
				if eval(fcc) {
					break
				}
			}
		}
	} else { // loop until break or continue target specified
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
