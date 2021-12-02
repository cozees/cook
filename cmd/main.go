package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/cozees/cook/pkg/cook/parser"
	"github.com/cozees/cook/pkg/runtime/args"
	"github.com/cozees/cook/pkg/runtime/function"
)

func main() {
	opts, err := args.ParseMainArgument(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	} else if opts.FuncMeta != nil {
		executeFunction(opts)
		os.Exit(0)
	}
	p := parser.NewParser()
	cook, err := p.Parse(opts.Cookfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	} else if len(opts.Targets) > 0 {
		cook.ExecuteWithTarget(opts.Args, opts.Targets...)
	} else {
		cook.Execute(opts.Args)
	}
}

func executeFunction(opts *args.MainOptions) {
	fn := function.GetFunction(opts.FuncMeta.Name)
	result, err := fn.Apply(opts.FuncMeta.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while execute function @%s: %s\n", opts.FuncMeta.Name, err.Error())
		os.Exit(1)
	} else if result == nil {
		fmt.Fprintf(os.Stdout, "nil")
		os.Exit(0)
	}
	// build output
	var output func(w *bufio.Writer, v reflect.Value, vk reflect.Kind) error
	output = func(w *bufio.Writer, v reflect.Value, vk reflect.Kind) (err error) {
		switch {
		case vk <= reflect.Complex128 || vk == reflect.String:
			if _, err = w.WriteString(fmt.Sprintf("%v", v.Interface())); err != nil {
				return err
			}
		case vk == reflect.Array || vk == reflect.Slice:
			if err = w.WriteByte('['); err != nil {
				return err
			}
			size := v.Len()
			for i := 0; i < size; i++ {
				sv := v.Index(i)
				if err = output(w, sv, sv.Kind()); err != nil {
					return err
				}
			}
			if err = w.WriteByte(']'); err != nil {
				return err
			}
		case vk == reflect.Map:
			if err = w.WriteByte('{'); err != nil {
				return err
			}
			keys := v.MapRange()
			for keys.Next() {
				kv := keys.Key()
				if err = output(w, kv, kv.Kind()); err != nil {
					return err
				} else if _, err = w.WriteString(": "); err != nil {
					return err
				}
				sv := v.MapIndex(kv)
				if err = output(w, sv, sv.Kind()); err != nil {
					return err
				}
			}
			if err = w.WriteByte('}'); err != nil {
				return err
			}
		default:
			if v.CanInterface() {
				iv := v.Interface()
				var reader io.Reader
				if r, ok := iv.(io.ReadCloser); ok {
					defer r.Close()
					reader = r
				} else if r := iv.(io.Reader); ok {
					reader = r
				}
				if reader != nil {
					if _, err = io.Copy(w, reader); err != nil {
						return err
					}
				}
			} else {
				return fmt.Errorf("function return unsupported kind %s", vk)
			}
		}
		return nil
	}
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	if err = output(w, reflect.ValueOf(result), reflect.TypeOf(result).Kind()); err != nil {
		fmt.Fprintf(os.Stderr, "error while writing function @%s output: %s\n", opts.FuncMeta.Name, err.Error())
		os.Exit(1)
	}
	w.WriteByte('\n')
}
