package ast

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"reflect"

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

	SizeOf struct {
		*baseExpr
		X Expr
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
