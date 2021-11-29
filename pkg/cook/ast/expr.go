package ast

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"

	"github.com/cozees/cook/pkg/cook/token"
)

//
type Expr interface {
	PosInfo() string
	Line() int
	Column() int
	Filename() string
	String() string
	evaluate(ctx cookContext) (interface{}, reflect.Kind)
}

type (
	baseExpr struct {
		line     int
		column   int
		filename string
	}

	Ident struct {
		*baseExpr
		Name string
	}

	Conditional struct {
		*baseExpr
		Cond  Expr
		True  Expr
		False Expr
	}

	Fallback struct {
		*baseExpr
		Primary   Expr
		Secondary Expr
	}

	SizeOf struct {
		*baseExpr
		X Expr
	}

	IsType struct {
		*baseExpr
		Var   Expr
		Bit   int
		Types []token.Token
	}

	TypeCast struct {
		*baseExpr
		X  Expr
		To reflect.Kind
	}

	ListLiteralExpr struct {
		*baseExpr
		Values []Expr
	}

	MapLiteralExpr struct {
		*baseExpr
		Keys   []Expr
		Values []Expr
	}

	IndexExpr struct {
		*baseExpr
		Index    Expr
		Variable Expr
	}

	CallExpr struct {
		*baseExpr
		Kind           token.Token
		Name           string
		Args           []Expr
		OutputAsResult bool // use for external command call with #
	}

	ReadFromExpr struct {
		*baseExpr
		X Expr
	}

	ParenExpr struct {
		*baseExpr
		Inner Expr
	}

	UnaryExpr struct {
		*baseExpr
		Op token.Token
		X  Expr
	}

	IncDecExpr struct {
		*baseExpr
		Op token.Token
		X  Expr
	}

	BinaryExpr struct {
		*baseExpr
		L  Expr
		Op token.Token
		R  Expr
	}
)

func (be *baseExpr) Line() int        { return be.line }
func (be *baseExpr) Column() int      { return be.column }
func (be *baseExpr) Filename() string { return be.filename }
func (be *baseExpr) PosInfo() string  { return fmt.Sprintf("%s:%d:%d", be.filename, be.line, be.column) }

func BaseExpr(offs int, file *token.File) *baseExpr {
	name, line, column := file.Position(offs)
	return &baseExpr{filename: name, line: line, column: column}
}

func NewIdentExpr(offs int, f *token.File, name string) Expr {
	return &Ident{
		baseExpr: BaseExpr(offs, f),
		Name:     name,
	}
}

func (i *Ident) evaluate(ctx cookContext) (v interface{}, vk reflect.Kind) {
	ctx.position(i.baseExpr)
	v, vk, _ = ctx.getVariable(i.Name)
	return
}

func (i *Ident) String() string { return i.Name }

func NewTernaryExpr(offs int, f *token.File, cond, T, F Expr) Expr {
	return &Conditional{
		baseExpr: BaseExpr(offs, f),
		Cond:     cond,
		True:     T,
		False:    F,
	}
}

func (c *Conditional) evaluate(ctx cookContext) (v interface{}, vk reflect.Kind) {
	ctx.position(c.baseExpr)
	a, ak := c.Cond.evaluate(ctx)
	if ak != reflect.Bool {
		ctx.onError(fmt.Errorf("%s does not produce boolean value", c.Cond.String()))
		return
	}
	if a.(bool) {
		return c.True.evaluate(ctx)
	} else {
		return c.False.evaluate(ctx)
	}
}

func (c *Conditional) String() string {
	return c.Cond.String() + " ? " + c.True.String() + " : " + c.False.String()
}

func NewFallbackExpr(offs int, f *token.File, primary, secondary Expr) Expr {
	return &Fallback{
		baseExpr:  BaseExpr(offs, f),
		Primary:   primary,
		Secondary: secondary,
	}
}

func (fn *Fallback) evaluate(ctx cookContext) (v interface{}, vk reflect.Kind) {
	ctx.position(fn.baseExpr)
	ctx.recordFailure()
	v, vk = fn.Primary.evaluate(ctx)
	// ensure that flag record failure is alway reset to false
	isFaile := ctx.hasFailure()
	if (vk == reflect.Invalid && isFaile) || (v == nil && !ctx.hasCanceled()) {
		v, vk = fn.Secondary.evaluate(ctx)
	}
	return
}

func (fn *Fallback) String() string {
	return fn.Primary.String() + " ?? " + fn.Secondary.String()
}

func NewSizeOfExpr(offs int, f *token.File, x Expr) Expr {
	return &SizeOf{
		baseExpr: BaseExpr(offs, f),
		X:        x,
	}
}

func (so *SizeOf) evaluate(ctx cookContext) (v interface{}, vk reflect.Kind) {
	ctx.position(so.baseExpr)
	i, ik := so.X.evaluate(ctx)
	switch ik {
	case reflect.Array, reflect.Slice, reflect.Map:
		v, vk = int64(reflect.ValueOf(i).Len()), reflect.Int64
	case reflect.String:
		v, vk = int64(len(i.(string))), reflect.Int64
	case reflect.Int64, reflect.Float64:
		v, vk = int64(8), reflect.Int64
	case reflect.Bool:
		v, vk = int64(1), reflect.Int64
	}
	return
}

func (so *SizeOf) String() string { return "sizeof " + so.X.String() }

func NewIsTypeExpr(offs int, f *token.File, v Expr, ttok ...token.Token) Expr {
	// bit indicator
	// TINTEGER, TFLOAT, TSTRING, TBOOLEAN, TARRAY, TMAP, TOBJECT
	// 1,		 2,		 3,		  4,		5,		6,	  7
	// 000001  , 000010, 000011
	bit := 0
	for _, tok := range ttok {
		bit |= tok.Type()
	}
	return &IsType{
		baseExpr: BaseExpr(offs, f),
		Var:      v,
		Bit:      bit,
		Types:    ttok,
	}
}

func (it *IsType) evaluate(ctx cookContext) (v interface{}, vk reflect.Kind) {
	ctx.position(it.baseExpr)
	if bl, ok := it.Var.(*BasicLiteral); ok {
		bit := bl.kind.Type()
		v = it.Bit&bit == bit
		vk = reflect.Bool
	} else if _, ok = it.Var.(*Ident); ok {
		var bit int
		switch _, vk = it.Var.evaluate(ctx); vk {
		case reflect.Int64:
			bit = token.TINTEGER.Type()
		case reflect.Float64:
			bit = token.TFLOAT.Type()
		case reflect.String:
			bit = token.TSTRING.Type()
		case reflect.Bool:
			bit = token.TBOOLEAN.Type()
		case reflect.Array, reflect.Slice:
			bit = token.TARRAY.Type()
		case reflect.Map:
			bit = token.TMAP.Type()
		default:
			bit = token.TOBJECT.Type()
		}
		v = it.Bit&bit == bit
		vk = reflect.Bool
	} else {
		ctx.onError(fmt.Errorf("%s must be a literal value or variable", it.Var.String()))
	}
	return
}

func (it *IsType) String() string {
	buf := strings.Builder{}
	for _, t := range it.Types {
		buf.WriteString(" | ")
		buf.WriteString(t.String())
	}
	return it.Var.String() + " is " + buf.String()[3:]
}

func NewTypeCastExpr(offs int, f *token.File, x Expr, to reflect.Kind) Expr {
	return &TypeCast{
		baseExpr: BaseExpr(offs, f),
		X:        x,
		To:       to,
	}
}

func (ctv *TypeCast) evaluate(ctx cookContext) (v interface{}, vk reflect.Kind) {
	ctx.position(ctv.baseExpr)
	iv, ik := ctv.X.evaluate(ctx)
	if ctv.To != ik {
		var err error
		switch ctv.To {
		case reflect.Int64:
			if ik == reflect.Float64 {
				return int64(iv.(float64)), ctv.To
			} else if vk == reflect.String {
				if v, err = strconv.ParseInt(iv.(string), 10, 64); err == nil {
					vk = ctv.To
					return
				}
			}
		case reflect.Float64:
			if ik == reflect.Int64 {
				return float64(iv.(int64)), ctv.To
			} else if ik == reflect.String {
				if v, err = strconv.ParseFloat(iv.(string), 64); err == nil {
					vk = ctv.To
					return
				}
			}
		case reflect.Bool:
			if ik == reflect.String {
				if v, err = strconv.ParseBool(iv.(string)); err == nil {
					vk = ctv.To
					return
				}
			}
		case reflect.String:
			if ik == reflect.Int64 {
				return strconv.FormatInt(iv.(int64), 10), ctv.To
			} else if ik == reflect.Float64 {
				return strconv.FormatFloat(iv.(float64), 'g', -1, 64), ctv.To
			} else if ik == reflect.Bool {
				return strconv.FormatBool(iv.(bool)), ctv.To
			}
		}
		ctx.onError(fmt.Errorf("cannot cast %v to type %s", iv, ctv.To))
	} else {
		v, vk = iv, ik
	}
	return
}

func (ctv *TypeCast) String() string {
	to := ""
	switch ctv.To {
	case reflect.Int64:
		to = token.TINTEGER.String()
	case reflect.Float64:
		to = token.TFLOAT.String()
	case reflect.String:
		to = token.TSTRING.String()
	case reflect.Bool:
		to = token.TBOOLEAN.String()
	default:
		panic("illegal state parser")
	}
	return to + "(" + ctv.X.String() + ")"
}

func NewListLiteral(offs int, f *token.File, exprs []Expr) Expr {
	return &ListLiteralExpr{
		baseExpr: BaseExpr(offs, f),
		Values:   exprs,
	}
}

func (ll *ListLiteralExpr) evaluate(ctx cookContext) (v interface{}, vk reflect.Kind) {
	ctx.position(ll.baseExpr)
	vals := make([]interface{}, len(ll.Values))
	for i, val := range ll.Values {
		if vals[i], _ = val.evaluate(ctx); ctx.hasCanceled() {
			return nil, reflect.Invalid
		}
	}
	return vals, reflect.Array
}

func (ll *ListLiteralExpr) String() string {
	buffer := bytes.NewBufferString("[")
	for _, val := range ll.Values {
		buffer.WriteString(val.String())
		buffer.WriteString(", ")
	}
	buffer.Truncate(buffer.Len() - 2)
	buffer.WriteRune(']')
	return buffer.String()
}

func NewMapLiteral(offs int, f *token.File, keys, values []Expr) Expr {
	return &MapLiteralExpr{
		baseExpr: BaseExpr(offs, f),
		Keys:     keys,
		Values:   values,
	}
}

func (ml *MapLiteralExpr) evaluate(ctx cookContext) (v interface{}, vk reflect.Kind) {
	ctx.position(ml.baseExpr)
	vals := make(map[interface{}]interface{})
	for i, key := range ml.Keys {
		if vals[key], _ = ml.Values[i].evaluate(ctx); ctx.hasCanceled() {
			return nil, reflect.Invalid
		}
	}
	return vals, reflect.Map
}

func (ml *MapLiteralExpr) String() string {
	buffer := bytes.NewBufferString("{")
	for i, key := range ml.Keys {
		val := ml.Values[i]
		buffer.WriteString(key.String())
		buffer.WriteString(": ")
		buffer.WriteString(val.String())
		buffer.WriteString(", ")
	}
	buffer.Truncate(buffer.Len() - 2)
	buffer.WriteRune('}')
	return buffer.String()
}

func NewIndexExpr(offs int, f *token.File, index Expr, variable Expr) Expr {
	return &IndexExpr{
		baseExpr: BaseExpr(offs, f),
		Index:    index,
		Variable: variable,
	}
}

func (i *IndexExpr) evaluate(ctx cookContext) (v interface{}, vk reflect.Kind) {
	ctx.position(i.baseExpr)
	if v, vk = i.Variable.evaluate(ctx); ctx.hasCanceled() {
		return
	}
	if vk != reflect.Array && vk != reflect.Slice {
		ctx.onError(fmt.Errorf("variable %s is not array", i.Variable))
		return
	}
	ind, vk := i.Index.evaluate(ctx)
	if ctx.hasCanceled() {
		return
	}
	var arr = v.([]interface{})
	var ii, fi = 0, 0.0
	switch vk {
	case reflect.Int64:
		ii = int(ind.(int64))
		goto exitWithResult
	case reflect.Float64:
		fi = ind.(float64)
		goto floatNum
	case reflect.Float32:
		fi = float64(ind.(float32))
		goto floatNum
	}
	// no suitable conversion
	goto exitWithError

floatNum:
	if math.Trunc(fi) != fi {
		goto exitWithError
	}
	ii = int(fi)

exitWithResult:
	if ii < 0 || ii >= len(arr) {
		ctx.onError(fmt.Errorf("%s index out of range 0, %d", i.PosInfo(), len(arr)-1))
		return
	}
	v = arr[ii]
	vk = reflect.ValueOf(v).Kind()
	return

exitWithError:
	ctx.onError(fmt.Errorf("index expression %s is not integer", i.Index))
	return nil, reflect.Invalid
}

func (i *IndexExpr) String() string {
	return i.Variable.String() + "[" + i.Index.String() + "]"
}

func NewCallExpr(offs int, f *token.File, name string, kind token.Token, args []Expr) Expr {
	return &CallExpr{
		baseExpr: BaseExpr(offs, f),
		Kind:     kind,
		Name:     name,
		Args:     args,
	}
}

func (c *CallExpr) evaluate(ctx cookContext) (v interface{}, vk reflect.Kind) {
	ctx.position(c.baseExpr)
	var err error
	switch c.Kind {
	case token.HASH:
		if args := c.args(ctx); !ctx.hasCanceled() {
			cmd := exec.Command(c.Name, args...)
			dir, err := os.Getwd()
			if err != nil {
				ctx.onError(err)
				return
			}
			cmd.Dir = dir
			cmd.Stdin = os.Stdin
			if !c.OutputAsResult {
				cmd.Stdout = os.Stdout
				if err = cmd.Run(); err != nil {
					ctx.onError(err)
				} else {
					return "", reflect.String
				}
			} else {
				result, err := cmd.Output()
				if err != nil {
					ctx.onError(err)
				} else {
					return string(result), reflect.String
				}
			}
		}
	case token.AT:
		t := ctx.getTarget(c.Name)
		if t == nil {
			f := ctx.getFunction(c.Name)
			if f == nil {
				ctx.onError(fmt.Errorf("target %s is not exist", v))
			} else if args := c.args(ctx); !ctx.hasCanceled() {
				if v, err = f.Apply(args); err != nil {
					ctx.onError(err)
				} else {
					return v, reflect.ValueOf(v).Kind()
				}
			}
		} else if args := c.args(ctx); !ctx.hasCanceled() {
			t.run(ctx, args)
		}
	default:
		panic("call expression kind must either @ or #")
	}
	return nil, reflect.Invalid
}

func (c *CallExpr) args(ctx cookContext) []string {
	args := make([]string, 0, len(c.Args))
	for _, arg := range c.Args {
		if v, vk := arg.evaluate(ctx); !ctx.hasCanceled() {
			switch vk {
			case reflect.Array, reflect.Slice:
				args = expandArrayTo(ctx, reflect.ValueOf(v), args)
			default:
				args = append(args, convertToString(ctx, v, vk))
			}
		}
	}
	return args
}

func (c *CallExpr) String() string {
	buffer := bytes.NewBufferString("")
	buffer.WriteString(c.Kind.String())
	buffer.WriteString(c.Name)
	buffer.WriteRune(' ')
	for _, arg := range c.Args {
		buffer.WriteString(arg.String())
		buffer.WriteRune(' ')
	}
	buffer.Truncate(buffer.Len() - 1)
	return buffer.String()
}

func NewReadFromExpr(offs int, f *token.File, x Expr) Expr {
	return &ReadFromExpr{
		baseExpr: BaseExpr(offs, f),
		X:        x,
	}
}

func (rf *ReadFromExpr) evaluate(ctx cookContext) (v interface{}, vk reflect.Kind) {
	ctx.position(rf.baseExpr)
	sv, svk := rf.X.evaluate(ctx)
	switch svk {
	case reflect.String:
		if b, err := ioutil.ReadFile(sv.(string)); err != nil {
			ctx.onError(err)
		} else {
			v, vk = string(b), reflect.String
		}
	default:
		ctx.onError(fmt.Errorf("value %v type %s is not a file path", sv, svk))
	}
	return
}

func (rf *ReadFromExpr) String() string {
	return "< " + rf.X.String()
}

func NewParanExpr(offs int, f *token.File, expr Expr) Expr {
	return &ParenExpr{
		baseExpr: BaseExpr(offs, f),
		Inner:    expr,
	}
}

func (p *ParenExpr) evaluate(ctx cookContext) (interface{}, reflect.Kind) {
	ctx.position(p.baseExpr)
	return p.Inner.evaluate(ctx)
}

func (p *ParenExpr) String() string {
	return "(" + p.Inner.String() + ")"
}

func NewUnaryExpr(offs int, f *token.File, op token.Token, expr Expr) Expr {
	return &UnaryExpr{
		baseExpr: BaseExpr(offs, f),
		Op:       op,
		X:        expr,
	}
}

func (ue *UnaryExpr) evaluate(ctx cookContext) (interface{}, reflect.Kind) {
	ctx.position(ue.baseExpr)
	v, vk := ue.X.evaluate(ctx)
	switch {
	case ue.Op == token.ADD:
		if vk == reflect.String {
			v, vk = convertToNum(v.(string))
		}
		if vk == reflect.Float64 || vk == reflect.Int64 {
			return v, vk
		}
	case ue.Op == token.SUB:
		if vk == reflect.String {
			v, vk = convertToNum(v.(string))
		}
		if vk == reflect.Float64 {
			return -v.(float64), vk
		} else if vk == reflect.Int64 {
			return -v.(int64), vk
		}
	case ue.Op == token.XOR && vk == reflect.Int64:
		return ^v.(int64), vk
	case ue.Op == token.NOT:
		switch vk {
		case reflect.Float64:
			return v.(float64) != 0.0, reflect.Bool
		case reflect.Int64:
			return v.(int64) != 0, reflect.Bool
		case reflect.Bool:
			return !v.(bool), vk
		case reflect.String:
			return v.(string) != "", reflect.Bool
		case reflect.Array:
			return len(v.([]interface{})) > 0, reflect.Bool
		case reflect.Map:
			return len(v.([]interface{})) > 0, reflect.Bool
		default:
			return v != nil, reflect.Bool
		}
	}
	ctx.onError(fmt.Errorf("unary operator %s is not supported on value %v`", ue.Op, v))
	return nil, reflect.Invalid
}

func (p *UnaryExpr) String() string {
	return p.Op.String() + p.X.String()
}

func NewIncDecExpr(offs int, f *token.File, op token.Token, x Expr) Expr {
	return &IncDecExpr{
		baseExpr: BaseExpr(offs, f),
		Op:       op,
		X:        x,
	}
}

func (ide *IncDecExpr) evaluate(ctx cookContext) (v interface{}, vk reflect.Kind) {
	ctx.position(ide.baseExpr)
	i, ok := ide.X.(*Ident)
	if !ok {
		panic("expression must be a variable")
	}
	v, vk, env := ctx.getVariable(i.Name)
	if env {
		ctx.onError(fmt.Errorf("variable %s is a read only environment variable", i.Name))
		return
	}
	// step value for increase or decrease
	step := int64(0)
	switch ide.Op {
	case token.INC:
		step = 1
	case token.DEC:
		step = -1
	default:
		panic("illegal state parser, operator must be increment or decrement")
	}
retryOnString:
	switch vk {
	case reflect.Float64:
		v = v.(float64) + float64(step)
		ctx.setVariable(i.Name, v)
		return
	case reflect.Int64:
		v = v.(int64) + step
		ctx.setVariable(i.Name, v)
		return
	case reflect.String:
		v, vk = convertToNum(v.(string))
		if vk == reflect.Int64 || vk == reflect.Float64 {
			goto retryOnString
		}
	}
	ctx.onError(fmt.Errorf("unsupported operator %s on variable %s of kind %s", ide.Op, i.Name, vk))
	return
}

func (ide *IncDecExpr) String() string {
	return ide.X.String() + ide.Op.String()
}

func NewBinaryExpr(offs int, f *token.File, op token.Token, L, R Expr) Expr {
	return &BinaryExpr{
		baseExpr: BaseExpr(offs, f),
		L:        L,
		Op:       op,
		R:        R,
	}
}

func (b *BinaryExpr) evaluate(ctx cookContext) (v interface{}, vk reflect.Kind) {
	ctx.position(b.baseExpr)
	vl, vkl := b.L.evaluate(ctx)
	vr, vkr := b.R.evaluate(ctx)
	if vkl == reflect.Invalid || vkr == reflect.Invalid {
		return nil, reflect.Invalid
	}
	switch {
	case b.Op == token.ADD:
		return addOperator(ctx, vl, vr, vkl, vkr)
	case token.ADD < b.Op && b.Op < token.LAND:
		return numOperator(ctx, b.Op, vl, vr, vkl, vkr)
	case token.EQL <= b.Op && b.Op <= token.GEQ:
		return logicOperator(ctx, b.Op, vl, vr, vkl, vkr)
	case vkl == vkr && vkl == reflect.Bool:
		if b.Op == token.LAND {
			return vl.(bool) && vr.(bool), reflect.Bool
		} else if b.Op == token.LOR {
			return vl.(bool) || vr.(bool), reflect.Bool
		}
	}
	ctx.onError(fmt.Errorf("unsupported operator %s on value %v and %v", b.Op, vl, vr))
	return
}

func (b *BinaryExpr) String() string {
	return "(" + b.L.String() + b.Op.String() + b.R.String() + ")"
}
