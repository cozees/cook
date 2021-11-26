package ast

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/cozees/cook/pkg/cook/token"
)

func convertToNum(s string) (interface{}, reflect.Kind) {
	if iv, err := strconv.ParseInt(s, 10, 64); err == nil {
		return iv, reflect.Int64
	} else if fv, err := strconv.ParseFloat(s, 64); err == nil {
		return fv, reflect.Float64
	}
	return nil, reflect.Invalid
}

func convertToFloat(ctx cookContext, val interface{}, kind reflect.Kind) float64 {
	switch kind {
	case reflect.Int64:
		return float64(val.(int64))
	case reflect.Float64:
		return val.(float64)
	case reflect.String:
		f, err := strconv.ParseFloat(val.(string), 64)
		if err != nil {
			ctx.onError(err)
			return 0
		} else {
			return f
		}
	default:
		ctx.onError(fmt.Errorf("value %v cannot cast/convert to float", val))
		return 0
	}
}

func convertToInt(ctx cookContext, val interface{}, kind reflect.Kind) int64 {
	switch kind {
	case reflect.Int64:
		return val.(int64)
	case reflect.Float64:
		ctx.onError(fmt.Errorf("value %v will be cut when cast to integer", val))
		return 0
	case reflect.String:
		v, err := strconv.ParseInt(val.(string), 10, 64)
		if err != nil {
			ctx.onError(err)
		}
		return v
	default:
		ctx.onError(fmt.Errorf("value %v cannot cast to string", val))
		return 0
	}
}

func convertToString(ctx cookContext, val interface{}, kind reflect.Kind) string {
	switch kind {
	case reflect.Int64:
		return strconv.FormatInt(val.(int64), 10)
	case reflect.Float64:
		return strconv.FormatFloat(val.(float64), 'g', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(val.(bool))
	case reflect.String:
		return val.(string)
	default:
		if b, ok := val.([]byte); ok {
			return string(b)
		} else {
			ctx.onError(fmt.Errorf("value %v cannot cast to string", val))
			return ""
		}
	}
}

func convertToReadCloser(ctx cookContext, val interface{}, kind reflect.Kind) io.ReadCloser {
	switch kind {
	case reflect.Int64:
		return io.NopCloser(bytes.NewBuffer([]byte(strconv.FormatInt(val.(int64), 10))))
	case reflect.Float64:
		return io.NopCloser(bytes.NewBuffer([]byte(strconv.FormatFloat(val.(float64), 'g', -1, 64))))
	case reflect.Bool:
		return io.NopCloser(bytes.NewBuffer([]byte(strconv.FormatBool(val.(bool)))))
	case reflect.String:
		return io.NopCloser(bytes.NewBuffer([]byte(val.(string))))
	default:
		if b, ok := val.([]byte); ok {
			return io.NopCloser(bytes.NewBuffer(b))
		} else if reader, ok := val.(io.ReadCloser); ok {
			return reader
		} else {
			ctx.onError(fmt.Errorf("value %v cannot cast to string", val))
			return nil
		}
	}
}

func addOperator(ctx cookContext, vl, vr interface{}, vkl, vkr reflect.Kind) (interface{}, reflect.Kind) {
	// array operation
	// 0: 1 + ["a", 2, 3.5] => [1, "a", 2, 3.5]
	// 1: ["b", 123, true] + 2.1 => ["b", 123, true, 2.1]
	// 2: [1, 2] + ["a", true] => [1, 2, "a", true]
	head := 0
	switch {
	case vkl == reflect.Array || vkl == reflect.Slice:
		if vkr == reflect.Array || vkr == reflect.Slice {
			head = 2
		} else {
			head = 1
		}
		goto opOnArray
	case vkr == reflect.Array || vkr == reflect.Slice:
		goto opOnArray
	case vkl == reflect.String || vkr == reflect.String:
		return convertToString(ctx, vl, vkl) + convertToString(ctx, vr, vkr), reflect.String
	case vkl == reflect.Float64 || vkr == reflect.Float64:
		return convertToFloat(ctx, vl, vkl) + convertToFloat(ctx, vr, vkr), reflect.Float64
	case vkl == reflect.Int64 && vkr == reflect.Int64:
		return vl.(int64) + vr.(int64), reflect.Int64
	}
	// value is not suitable to operate with + operator
	ctx.onError(fmt.Errorf("operator + is not supported for value %v and %v", vl, vr))
	return nil, reflect.Invalid

opOnArray:
	switch head {
	case 0:
		return append([]interface{}{vl}, vr.([]interface{})...), reflect.Array
	case 1:
		return append(vl.([]interface{}), vr), reflect.Array
	case 2:
		return append(vl.([]interface{}), vr.([]interface{})...), reflect.Array
	default:
		panic("illegal state for array operation")
	}
}

func numOperator(ctx cookContext, op token.Token, vl, vr interface{}, vkl, vkr reflect.Kind) (interface{}, reflect.Kind) {
	var fa, fb float64
	var ia, ib int64
	switch {
	case vkl == reflect.Float64 || vkr == reflect.Float64:
		if token.ADD < op && op < token.REM {
			goto numFloat
		}
	case vkl == reflect.Int64 || vkr == reflect.Int64:
		if token.ADD < op && op < token.LAND {
			goto numInt
		}
	}
	// value is not suitable
	ctx.onError(fmt.Errorf("operator %s is not supported for value %v and %v", op, vl, vr))
	return nil, reflect.Invalid

numFloat:
	fa, fb = convertToFloat(ctx, vl, vkl), convertToFloat(ctx, vr, vkr)
	switch op {
	case token.SUB:
		return fa - fb, reflect.Float64
	case token.MUL:
		return fa * fb, reflect.Float64
	case token.QUO:
		return fa / fb, reflect.Float64
	default:
		panic("illegal state operation")
	}
numInt:
	ia, ib = convertToInt(ctx, vl, vkl), convertToInt(ctx, vr, vkr)
	switch op {
	case token.SUB:
		return ia - ib, reflect.Int64
	case token.MUL:
		return ia * ib, reflect.Int64
	case token.QUO:
		return ia / ib, reflect.Int64
	case token.REM:
		return ia % ib, reflect.Int64
	case token.AND:
		return ia & ib, reflect.Int64
	case token.OR:
		return ia | ib, reflect.Int64
	case token.XOR:
		return ia ^ ib, reflect.Int64
	case token.SHL:
		return ia << ib, reflect.Int64
	case token.SHR:
		return ia >> ib, reflect.Int64
	default:
		panic("illegal state operation")
	}
}

func logicOperator(ctx cookContext, op token.Token, vl, vr interface{}, vkl, vkr reflect.Kind) (interface{}, reflect.Kind) {
	if (vkl == reflect.Float64 || vkl == reflect.Int64) &&
		(vkr == reflect.Float64 || vkr == reflect.Int64) {
		fl, fr := convertToFloat(ctx, vl, vkl), convertToFloat(ctx, vr, vkr)
		switch op {
		case token.EQL:
			return fl == fr, reflect.Bool
		case token.LSS:
			return fl < fr, reflect.Bool
		case token.GTR:
			return fl > fr, reflect.Bool
		case token.NEQ:
			return fl != fr, reflect.Bool
		case token.LEQ:
			return fl <= fr, reflect.Bool
		case token.GEQ:
			return fl >= fr, reflect.Bool
		default:
			panic("illegal state operator")
		}
	} else if vkl == vkr {
		// any type other than integer or float must has the same type
		switch vkl {
		case reflect.Bool:
			if op == token.EQL {
				return vl.(bool) == vr.(bool), reflect.Bool
			} else if op == token.NEQ {
				return vl.(bool) != vr.(bool), reflect.Bool
			}
		case reflect.String:
			r := strings.Compare(vl.(string), vr.(string))
			switch op {
			case token.EQL:
				return r == 0, reflect.Bool
			case token.LSS:
				return r < 0, reflect.Bool
			case token.GTR:
				return r > 0, reflect.Bool
			case token.NEQ:
				return r != 0, reflect.Bool
			case token.LEQ:
				return r <= 0, reflect.Bool
			case token.GEQ:
				return r >= 0, reflect.Bool
			}
		default:
			if op == token.EQL {
				return vl == vr, reflect.Bool
			} else if op == token.NEQ {
				return vl != vr, reflect.Bool
			}
		}
	}
	// value is not suitable
	ctx.onError(fmt.Errorf("operator %s is not supported for value %v and %v", op, vl, vr))
	return nil, reflect.Invalid
}

func expandArrayTo(ctx cookContext, rv reflect.Value, array []string) []string {
	for i := 0; i < rv.Len(); i++ {
		v := rv.Index(i)
		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			array = expandArrayTo(ctx, v, array)
		default:
			i := v.Interface()
			array = append(array, convertToString(ctx, i, reflect.ValueOf(i).Kind()))
		}
	}
	return array
}
