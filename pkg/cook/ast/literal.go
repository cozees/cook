package ast

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/cozees/cook/pkg/cook/token"
)

//
type BasicLiteral struct {
	*baseExpr
	lit  string
	kind token.Token
}

func NewBasicLiteral(offs int, f *token.File, kind token.Token, lit string) Expr {
	return &BasicLiteral{
		baseExpr: BaseExpr(offs, f),
		lit:      lit,
		kind:     kind,
	}
}

func (bl *BasicLiteral) evaluate(ctx cookContext) (v interface{}, vk reflect.Kind) {
	ctx.position(bl.baseExpr)
	if !ctx.hasCanceled() {
		var err error
		switch bl.kind {
		case token.INTEGER:
			if v, err = strconv.ParseInt(bl.lit, 10, 64); err != nil {
				panic(err)
			}
			vk = reflect.Int64
		case token.BOOLEAN:
			if v, err = strconv.ParseBool(bl.lit); err != nil {
				panic(err)
			}
			vk = reflect.Bool
		case token.FLOAT:
			if v, err = strconv.ParseFloat(bl.lit, 64); err != nil {
				panic(err)
			}
			vk = reflect.Float64
		case token.STRING:
			v, vk = bl.lit, reflect.String
		default:
			panic("illegal state parser. Parser must ensure create the right ast node.")
		}
	}
	return
}

func (bl *BasicLiteral) String() string {
	if bl.kind == token.STRING {
		return "\"" + strings.ReplaceAll(bl.lit, "\"", "\\\"") + "\""
	} else {
		return bl.lit
	}
}
