package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cozees/cook/pkg/cook/parser"
)

const (
	defaultCookfile = "Cookfile"
)

func lower(ch byte) byte { return ('a' - 'A') | ch }

func parseArgs() (cookfile string, targets []string, args map[string]interface{}) {
	cookfile = defaultCookfile
	iargs := os.Args[1:]
	if len(iargs) > 0 {
		args = make(map[string]interface{})
		offset, in := 0, 0
		for i := 0; i < len(iargs); i++ {
			in = i + 1
			switch {
			case strings.HasPrefix(iargs[i], "--"):
				offset = 2
			case strings.HasPrefix(iargs[i], "-"):
				offset = 1
			default:
				if len(iargs[i]) == 0 || !('a' <= lower(iargs[i][0]) && lower(iargs[i][0]) <= 'z' || iargs[i][0] == '_') {
					fmt.Fprintln(os.Stderr, "invalid target name", iargs[i])
					os.Exit(1)
				}
				targets = append(targets, iargs[i])
				continue
			}
			name, val := "", interface{}(nil)
			if equali := strings.IndexRune(iargs[i], '='); equali != -1 {
				name, val = iargs[i][offset:equali], iargs[i][equali+1:]
			} else if in < len(iargs) {
				name, val = iargs[i][offset:], iargs[in]
				i = in
			} else {
				name, val = iargs[i][offset:], true
			}
			if v, ok := args[name]; ok {
				if vs, ok := v.([]interface{}); ok {
					args[name] = append(vs, val)
				} else {
					args[name] = []interface{}{v, val}
				}
			} else {
				args[name] = val
			}
		}
	}
	return
}

func main() {
	cfile, targets, args := parseArgs()
	p := parser.NewParser()
	cook, err := p.Parse(cfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	if len(targets) > 0 {
		cook.ExecuteWithTarget(targets, args)
	} else {
		cook.Execute(args)
	}
}
