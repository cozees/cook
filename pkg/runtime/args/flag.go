package args

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrFlagSyntax  = fmt.Errorf("invalid flag format. E.g. --name:s, --name:i:f")
	ErrAllowFormat = fmt.Errorf("only i, f, s, b, a format is allowed")
	ErrVarSyntax   = fmt.Errorf("variable flag must start with --")
)

const (
	defaultCookfile = "Cookfile"
)

type MainOptions struct {
	Cookfile string
	Targets  []string
	Args     map[string]interface{}
}

func ParseMainArgument(args []string) (*MainOptions, error) {
	mo := &MainOptions{Cookfile: defaultCookfile}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--"):
			val := ""
			ieql := strings.IndexByte(arg, '=')
			if ieql == -1 {
				ieql = len(arg)
				if n := i + 1; n < len(args) {
					i = n
					val = args[i]
				}
			} else {
				val = arg[ieql+1:]
			}
			vname, p, s, err := parseFlagFormat(arg[2:ieql])
			if err != nil {
				return nil, err
			}
			// create args if first encouter
			if mo.Args == nil {
				mo.Args = make(map[string]interface{})
			}
			// special case for map
			var pv, pk interface{}
			if s != reflect.Invalid {
				icolon := strings.IndexByte(val, ':')
				if icolon < 1 {
					return nil, fmt.Errorf("invalid flag value map %s, must be format of key:value", val)
				}
				if pk, err = parseFlagValue(p, val[:icolon]); err != nil {
					return nil, err
				} else if pv, err = parseFlagValue(s, val[icolon+1:]); err != nil {
					return nil, err
				} else if mp, ok := mo.Args[vname]; !ok {
					mo.Args[vname] = map[interface{}]interface{}{pk: pv}
				} else if mpv, ok := mp.(map[interface{}]interface{}); ok {
					mpv[pk] = pv
				} else {
					return nil, fmt.Errorf("variable %s value %v (%s) is not a map", vname, mp, reflect.ValueOf(mp).Kind())
				}
			} else if pv, err = parseFlagValue(p, val); err != nil {
				return nil, err
			} else if exist, ok := mo.Args[vname]; ok {
				vex := reflect.ValueOf(exist)
				if vex.Kind() == reflect.Slice {
					mo.Args[vname] = reflect.Append(vex, reflect.ValueOf(pv)).Interface()
				} else {
					mo.Args[vname] = []interface{}{exist, pv}
				}
			} else {
				mo.Args[vname] = pv
			}
		case strings.HasPrefix(arg, "-"):
			if arg == "-c" {
				if n := i + 1; n < len(args) {
					i = n
					mo.Cookfile = args[i]
				}
				break
			}
			return nil, ErrVarSyntax
		default:
			if len(arg) == 0 || !('a' <= lower(arg[0]) && lower(arg[0]) <= 'z' || arg[0] == '_') {
				return nil, fmt.Errorf("invalid target name %s", arg)
			}
			if mo.Targets == nil {
				mo.Targets = make([]string, 0, 2)
			}
			mo.Targets = append(mo.Targets, arg)
		}
	}
	return mo, nil
}

type Flag struct {
	Short       string // single character, e.g. -e, -e
	Long        string // more 2 character, e.g. --name or -name
	Description string
}

func (flag *Flag) Set(field reflect.Value, val string, i int, args []string) (consumed bool, err error) {
	// get field type
	t := field.Kind()
	// convert value if needed
	n := i + 1
	switch {
	case t == reflect.Array:
		return false, fmt.Errorf("flag type must be a slice not an array")
	case t == reflect.Bool:
		if val != "" {
			return false, fmt.Errorf("boolean flag should not have extral value")
		}
		field.Set(reflect.ValueOf(true))
	case val == "":
		if n >= len(args) {
			return false, fmt.Errorf("not enough argument, missing value for flag %s", flag.Long)
		}
		val = args[n]
		i = n
		fallthrough
	default:
		// special case for map and array
		te := field.Type()
		switch t {
		case reflect.Slice:
			t = te.Elem().Kind()
			if fval, err := parseFlagValue(t, val); err != nil {
				return false, err
			} else {
				field.Set(reflect.Append(field, reflect.ValueOf(fval)))
			}
		case reflect.Map:
			icolon := strings.IndexByte(val, ':')
			if icolon < 1 {
				return false, fmt.Errorf("invalid map entry %s for flag %s", val, flag.Long)
			}
			var kval, vval interface{}
			if kval, err = parseFlagValue(te.Key().Kind(), val[:icolon]); err != nil {
				return false, err
			}
			if vval, err = parseFlagValue(te.Elem().Kind(), val[icolon+1:]); err != nil {
				return false, err
			}
			if field.IsNil() {
				field.Set(reflect.MakeMap(field.Type()))
			}
			field.SetMapIndex(reflect.ValueOf(kval), reflect.ValueOf(vval))
		default:
			var vval interface{}
			if vval, err = parseFlagValue(t, val); err != nil {
				return false, err
			}
			field.Set(reflect.ValueOf(vval))
		}
	}
	return n == i, nil
}

type Flags struct {
	Flags       []*Flag
	Result      reflect.Type
	Description string
}

func (flags *Flags) Validate() error {
	m := make(map[string]bool)
	for _, flag := range flags.Flags {
		if flag.Short != "" {
			if len(flag.Short) != 1 {
				return fmt.Errorf("short flag %s must have only 1 characters", flag.Short)
			}
			if _, ok := m[flag.Short]; ok {
				return fmt.Errorf("short flag %s already registered", flag.Short)
			}
			m[flag.Short] = true
		}
		if len(flag.Long) < 2 {
			return fmt.Errorf("long flag is required")
		}
		if _, ok := m[flag.Long]; ok {
			return fmt.Errorf("long flag %s already registered", flag.Long)
		}
		m[flag.Long] = true
	}
	return nil
}

func (flags *Flags) Parse(args []string) (v interface{}, err error) {
	if flags.Result.Kind() != reflect.Struct {
		return nil, fmt.Errorf("option result must be a pointer or struct")
	}
	var flag *Flag
	var remaining []string
	var fval string
	var val = reflect.New(flags.Result).Elem()
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--"):
			if flag, fval, err = flags.findFlag(arg[2:], false); err != nil {
				return
			}
		case strings.HasPrefix(arg, "-"):
			if len(arg) > 2 {
				err = fmt.Errorf("long flag %s required (--)", arg)
				return
			}
			if flag, fval, err = flags.findFlag(arg[1:], true); err != nil {
				return
			}
		default:
			remaining = append(remaining, arg)
			continue
		}
		// find field in struct with tag that belong to the flag
		field := findField(flags.Result, val, flag.Long)
		if !field.CanSet() {
			err = fmt.Errorf("field %s was not found or not exported", flag.Long)
			return
		}
		if inc, err := flag.Set(field, fval, i, args); err != nil {
			return nil, err
		} else if inc {
			i++
		}
	}
	if len(remaining) > 0 {
		field := val.FieldByName("Args")
		if !field.CanSet() {
			err = fmt.Errorf("options %s must defined field Args []string", flags.Result.Name())
			return
		}
		field.Set(reflect.ValueOf(remaining))
	}
	v = val.Addr().Interface()
	return
}

func (flags *Flags) findFlag(s string, short bool) (*Flag, string, error) {
	var val string
	if i := strings.IndexByte(s, '='); i == 0 {
		return nil, "", fmt.Errorf("flag must not start with =")
	} else if i > 0 {
		s, val = s[:i], s[i+1:]
	}
	for _, flag := range flags.Flags {
		if (short && flag.Short == s) || (!short && flag.Long == s) {
			return flag, val, nil
		}
	}
	return nil, "", fmt.Errorf("unrecognize flag %s", s)
}

func parseFlagValue(kind reflect.Kind, v string) (interface{}, error) {
	switch kind {
	case reflect.Int64:
		return strconv.ParseInt(v, 10, 64)
	case reflect.Float64:
		return strconv.ParseFloat(v, 64)
	case reflect.Bool:
		return strconv.ParseBool(v)
	case reflect.String: // s or default treat as string
		return v, nil
	case reflect.Interface:
		if i, err := parseFlagValue(reflect.Int64, v); err == nil {
			return i, nil
		} else if i, err = parseFlagValue(reflect.Float64, v); err == nil {
			return i, nil
		} else if i, err = parseFlagValue(reflect.Bool, v); err == nil {
			return i, nil
		} else {
			return v, nil
		}
	default:
		return nil, fmt.Errorf("unsupported type %s, only integer, float, boolean and string is allowed", kind)
	}
}

func findField(t reflect.Type, v reflect.Value, name string) reflect.Value {
	numField := t.NumField()
	for i := 0; i < numField; i++ {
		field := t.Field(i)
		if field.Name == name || field.Tag.Get("flag") == name {
			return v.Field(i)
		}
	}
	return reflect.Value{}
}
