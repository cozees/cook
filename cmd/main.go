package main

import (
	"fmt"
	"os"

	"github.com/cozees/cook/pkg/cook/parser"
	"github.com/cozees/cook/pkg/runtime/args"
)

func main() {
	opts, err := args.ParseMainArgument(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	p := parser.NewParser()
	cook, err := p.Parse(opts.Cookfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	if len(opts.Targets) > 0 {
		cook.ExecuteWithTarget(opts.Args, opts.Targets...)
	} else {
		cook.Execute(opts.Args)
	}
}
