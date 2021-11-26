package ast

import (
	"reflect"
	"testing"

	"github.com/cozees/cook/pkg/cook/token"
	"github.com/stretchr/testify/assert"
)

func TestCookProgram(t *testing.T) {
	cook := NewCookProgram()
	// dummy base expr
	be := &baseExpr{line: 0, column: 0, filename: "sample.txt"}
	cook.AddStatement(NewAssignStatement("v1", token.ASSIGN, &BasicLiteral{
		baseExpr: be,
		kind:     token.INTEGER,
		lit:      "123",
	}))
	cook.AddStatement(NewAssignStatement("v2", token.ASSIGN, &BinaryExpr{
		baseExpr: be,
		L:        &Ident{baseExpr: be, Name: "v1"},
		Op:       token.ADD,
		R: &BasicLiteral{
			baseExpr: be,
			kind:     token.FLOAT,
			lit:      "283.234",
		},
	}))
	cook.AddStatement(NewAssignStatement("v3", token.ASSIGN, &BinaryExpr{
		baseExpr: be,
		L:        &Ident{baseExpr: be, Name: "v2"},
		Op:       token.MUL,
		R:        &Ident{baseExpr: be, Name: "v1"},
	}))
	target, err := cook.AddTarget("sample")
	assert.NoError(t, err)
	target.AddStatement(NewAssignStatement("v4", token.ASSIGN, &BinaryExpr{
		baseExpr: be,
		L:        &Ident{baseExpr: be, Name: "v2"},
		Op:       token.MUL,
		R:        &Ident{baseExpr: be, Name: "v1"},
	}))
	cook.ExecuteWithTarget([]string{"sample"}, nil)
	icp := cook.(*implCookProgram)
	v, vk, env := icp.getVariable("v1")
	assert.Equal(t, false, env)
	assert.Equal(t, reflect.Int64, vk)
	assert.Equal(t, int64(123), v)
	v, vk, env = icp.getVariable("v2")
	assert.Equal(t, false, env)
	assert.Equal(t, reflect.Float64, vk)
	assert.Equal(t, float64(123+283.234), v)
	v, vk, env = icp.getVariable("v3")
	assert.Equal(t, false, env)
	assert.Equal(t, reflect.Float64, vk)
	assert.Equal(t, float64((123+283.234)*123), v)
	// any variable there initial target, for loop or if/else statement is allocate for local scope only
	// thus when exist target, for loop, or if/else statement the allocation is discard
	v, vk, env = icp.getVariable("v4")
	assert.Equal(t, false, env)
	assert.Equal(t, reflect.Invalid, vk)
	assert.Equal(t, nil, v)
}
