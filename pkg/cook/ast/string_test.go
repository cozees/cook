package ast

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stringBindTestCase struct {
	store map[string]interface{}
	in    []interface{}
	out   string
}

var sbTestCase = []*stringBindTestCase{
	{
		store: map[string]interface{}{"V": int64(112), "V1": "string binding"},
		in: []interface{}{
			"simple ", &Ident{Name: "V"}, " number in a ", &Ident{Name: "V1"},
		},
		out: "simple 112 number in a string binding",
	},
	{
		store: map[string]interface{}{"V": int64(112), "V1": " number", "V2": "string binding"},
		in: []interface{}{
			"simple ", &Ident{Name: "V"}, &Ident{Name: "V1"}, " in a ", &Ident{Name: "V2"},
		},
		out: "simple 112 number in a string binding",
	},
}

func TestStringBinding(t *testing.T) {
	for i, tc := range sbTestCase {
		t.Logf("TestStringBinding case #%d", i+1)
		builder := NewStringBindingBuilder("").(*sbEncoder)
		cp := NewCookProgram().(*implCookProgram)
		for _, vi := range tc.in {
			switch v := vi.(type) {
			case string:
				builder.WriteString(v)
			case Expr:
				builder.AddExpression(v)
			default:
				panic("unknown type")
			}
		}
		// by pass build
		x := builder.sb
		x.raw = builder.buffer.String()
		cp.gvar = tc.store
		result, kind := x.evaluate(cp)
		assert.Equal(t, reflect.String, kind)
		assert.Equal(t, tc.out, result)
	}
}
