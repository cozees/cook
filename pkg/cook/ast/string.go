package ast

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/cozees/cook/pkg/cook/token"
)

type StringBinding struct {
	*baseExpr
	raw       string
	injectPos []int
	exprs     []Expr
}

func (sb *StringBinding) Raw() string { return sb.raw }

func (sb *StringBinding) evaluate(ctx cookContext) (interface{}, reflect.Kind) {
	ctx.position(sb.baseExpr)
	if !ctx.hasCanceled() {
		buffer := bytes.NewBufferString("")
		offs, i, pos := 0, -1, -1
		for i, pos = range sb.injectPos {
			buffer.WriteString(sb.raw[offs:pos])
			val, kn := sb.exprs[i].evaluate(ctx)
			if val == nil {
				ctx.onError(fmt.Errorf("variable %s not found", sb.exprs[i]))
				return nil, reflect.Invalid
			}
			ss := convertToString(ctx, val, kn)
			buffer.WriteString(ss)
			offs = pos
		}
		if pos < len(sb.raw)-1 {
			buffer.WriteString(sb.raw[pos:])
		}
		return buffer.String(), reflect.String
	}
	return nil, reflect.Invalid
}

func (sb *StringBinding) String() string {
	buffer := bytes.NewBufferString("")
	offs, i, pos := 0, -1, -1
	for i, pos = range sb.injectPos {
		buffer.WriteString(sb.raw[offs:pos])
		buffer.WriteString(sb.exprs[i].String())
	}
	return buffer.String()
}

//
type Builder interface {
	WriteString(s string) (int, error)
	WriteRune(r rune) (int, error)
	AddExpression(x Expr)
	Build(offs int, f *token.File) Expr
}

func NewStringBindingBuilder(s string) Builder {
	return &sbEncoder{buffer: bytes.NewBufferString(s), sb: &StringBinding{}}
}

type sbEncoder struct {
	buffer *bytes.Buffer
	sb     *StringBinding
}

func (sbe *sbEncoder) WriteRune(r rune) (int, error)     { return sbe.buffer.WriteRune(r) }
func (sbe *sbEncoder) WriteString(s string) (int, error) { return sbe.buffer.WriteString(s) }

func (sbe *sbEncoder) AddExpression(x Expr) {
	sbe.sb.injectPos = append(sbe.sb.injectPos, sbe.buffer.Len())
	sbe.sb.exprs = append(sbe.sb.exprs, x)
}

func (sbe *sbEncoder) Build(offs int, f *token.File) Expr {
	sbe.sb.raw = sbe.buffer.String()
	if len(sbe.sb.exprs) == 1 && sbe.sb.raw == "" {
		// no need to export string binding expr, a $VAR or ${VAR} only will be treat as
		// identifier as value itself evaluate later on when execute
		return sbe.sb.exprs[0]
	} else if len(sbe.sb.exprs) > 1 {
		sbe.sb.baseExpr = BaseExpr(offs, f)
		return sbe.sb
	} else {
		return nil
	}
}
