package ast

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
)

type Target interface {
	BlockStatement
	Name() string
	run(ctx cookContext, args []string)
}

type implTarget struct {
	*implBlock
	*implForContext
	name string
}

func NewTarget(name string) Target {
	return &implTarget{name: name, implBlock: &implBlock{}}
}

func (it *implTarget) Name() string { return it.name }

func (it *implTarget) run(ctx cookContext, args []string) {
	it.implForContext = &implForContext{implContext: ctx.(*implContext), breakIndex: -1, continueIndex: -1}
	dir, err := os.Getwd()
	if err != nil {
		ctx.onError(err)
		return
	}
	// set directory back to where it is after target executed
	defer func() {
		os.Chdir(dir)
	}()
	localCtx := ctx.addScope()
	defer ctx.popScope()
	if it.name == "all" && len(it.statements) == 0 {
		// execute all other target
		for _, t := range it.getTargets() {
			switch t.Name() {
			case "all", "initialize", "finalize":
				continue
			}
			if !ctx.hasCanceled() {
				t.run(ctx, args)
			}
		}
	} else {
		localCtx[strconv.Itoa(0)] = int64(len(args))
		for i, v := range args {
			localCtx[strconv.Itoa(i+1)] = v
		}
		it.evaluate(it.implForContext)
	}
}

func (it *implTarget) String(indent int) string {
	buffer := bytes.NewBufferString(it.name)
	buffer.WriteString(":\n")
	for _, ins := range it.statements {
		buffer.WriteString(ins.String(indent + 4))
	}
	return buffer.String()
}

func isTargetContext(ctx cookContext) (forCookContext, error) {
	fcc, ok := ctx.(*implForContext)
	if !ok {
		return nil, fmt.Errorf("allow inside target only")
	}
	return fcc, nil
}
