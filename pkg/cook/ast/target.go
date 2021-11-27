package ast

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/cozees/cook/pkg/runtime/function"
)

type Target interface {
	BlockStatement
	Name() string
	run(ctx cookContext, args []string)
}

type implTarget struct {
	*implBlock
	ctx  cookContext
	name string

	// for name
	loops         []string
	breakIndex    int
	continueIndex int
}

func NewTarget(name string) Target {
	return &implTarget{name: name, implBlock: &implBlock{}, breakIndex: -1, continueIndex: -1}
}

// cookContext
func (it *implTarget) recordFailure()                            { it.ctx.recordFailure() }
func (it *implTarget) hasFailure() bool                          { return it.ctx.hasFailure() }
func (it *implTarget) addScope() map[string]interface{}          { return it.ctx.addScope() }
func (it *implTarget) popScope()                                 { it.ctx.popScope() }
func (it *implTarget) getTarget(name string) Target              { return it.ctx.getTarget(name) }
func (it *implTarget) getTargets() []Target                      { return it.ctx.getTargets() }
func (it *implTarget) getFunction(name string) function.Function { return it.ctx.getFunction(name) }
func (it *implTarget) setVariable(name string, val interface{})  { it.ctx.setVariable(name, val) }
func (it *implTarget) onError(err error)                         { it.ctx.onError(err) }
func (it *implTarget) hasCanceled() bool                         { return it.ctx.hasCanceled() }
func (it *implTarget) position(be *baseExpr)                     { it.ctx.position(be) }

func (it *implTarget) getVariable(name string) (value interface{}, kind reflect.Kind, fromEnv bool) {
	return it.ctx.getVariable(name)
}

func (it *implTarget) isLoopExist(id string) (int, error) {
	for i, n := range it.loops {
		if n == id {
			return i, nil
		}
	}
	return -1, fmt.Errorf("for name %s not found", id)
}

func (it *implTarget) breakAt() int    { return it.breakIndex }
func (it *implTarget) continueAt() int { return it.continueIndex }

func (it *implTarget) shouldBreakingOf(index int) bool {
	return it.breakIndex != -1 && (it.breakIndex <= index || it.continueIndex < index)
}

func (it *implTarget) shouldContinueOf(index int) bool {
	return it.continueIndex != -1 && it.continueIndex == index
}

// forCookContext
func (it *implTarget) breakWith(id string) error {
	if id == "" {
		it.breakIndex = len(it.loops)
	} else if i, err := it.isLoopExist(id); err != nil {
		return err
	} else {
		it.breakIndex = i
	}
	return nil
}

func (it *implTarget) continueWith(id string) error {
	if id == "" {
		it.continueIndex = len(it.loops)
	} else if i, err := it.isLoopExist(id); err != nil {
		return err
	} else {
		it.continueIndex = i
	}
	return nil
}

func (it *implTarget) register(id string) (index int) {
	it.loops = append(it.loops, id)
	return len(it.loops) - 1
}

func (it *implTarget) unregister(index int) {
	if 0 <= index && index < len(it.loops) {
		it.loops = append(it.loops[:index], it.loops[index+1:]...)
	} else {
		panic(fmt.Sprintf("loop index out of range %d, should be between 0, %d", index, len(it.loops)))
	}
}

func (it *implTarget) Name() string { return it.name }

func (it *implTarget) run(ctx cookContext, args []string) {
	it.ctx = ctx
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
		it.evaluate(it)
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
	fcc, ok := ctx.(*implTarget)
	if !ok {
		return nil, fmt.Errorf("allow inside target only")
	}
	return fcc, nil
}
