package ast

import (
	"fmt"
	"os"
	"reflect"

	"github.com/cozees/cook/pkg/runtime/function"
)

type cookContext interface {
	// add local scope for the tail of scope list.
	addScope() map[string]interface{}
	// remove last scope from the list
	popScope()

	getTargets() []Target
	getTarget(name string) Target
	getFunction(name string) function.Function

	getVariable(name string) (value interface{}, kind reflect.Kind, fromEnv bool)
	setVariable(name string, val interface{}) error

	position(be *baseExpr)
	onError(err error)
	hasCanceled() bool

	// tell context to record the error instead of printing out
	// and ignore it so no cancel is occurred
	recordFailure()
	hasFailure(reset bool) bool

	restrictVariable(name string, kind reflect.Kind)
}

type implContext struct {
	locals        []map[string]interface{}
	gvar          map[string]interface{}
	targets       []Target
	targetsByName map[string]int
	isCanceled    bool
	curBaseExpr   *baseExpr

	restrictVar map[string]reflect.Kind

	// record last error
	lastError   error
	recordError bool
}

func (icp *implContext) restrictVariable(name string, kind reflect.Kind) {
	if kind == reflect.Invalid {
		delete(icp.restrictVar, name)
	} else {
		icp.restrictVar[name] = kind
	}
}

func (icp *implContext) recordFailure() { icp.recordError = true }
func (icp *implContext) hasFailure(reset bool) bool {
	if reset {
		defer func() {
			icp.recordError = false
			icp.lastError = nil
		}()
	}
	return icp.lastError != nil
}

func (icp *implContext) addScope() map[string]interface{} {
	store := make(map[string]interface{})
	icp.locals = append(icp.locals, store)
	return store
}

func (icp *implContext) popScope()            { icp.locals = icp.locals[0 : len(icp.locals)-1] }
func (icp *implContext) getTargets() []Target { return icp.targets }

func (icp *implContext) getTarget(name string) Target {
	if index, ok := icp.targetsByName[name]; !ok {
		return nil
	} else {
		return icp.targets[index]
	}
}

func (icp *implContext) getFunction(name string) function.Function {
	return function.GetFunction(name)
}

func (icp *implContext) getVariable(name string) (value interface{}, kind reflect.Kind, fromEnv bool) {
	for i := len(icp.locals) - 1; i >= 0; i-- {
		if v, ok := icp.locals[i][name]; ok {
			value, kind, fromEnv = v, reflect.ValueOf(v).Kind(), false
			return
		}
	}
	if v, ok := icp.gvar[name]; ok {
		value, kind, fromEnv = v, reflect.ValueOf(v).Kind(), false
	} else if v := os.Getenv(name); v != "" {
		value, kind, fromEnv = v, reflect.String, true
	}
	return
}

func (icp *implContext) setVariable(name string, val interface{}) error {
	if sk, ok := icp.restrictVar[name]; ok {
		if sk != reflect.ValueOf(val).Kind() {
			return fmt.Errorf("variable \"%s\" must keep it origin type %s", name, sk)
		}
	}

	last := len(icp.locals) - 1
	for i := last; i >= 0; i-- {
		if _, ok := icp.locals[i][name]; ok {
			icp.locals[i][name] = val
			return nil
		}
	}
	// if exist in global replace the value otherwise create in local scope only.
	if _, ok := icp.gvar[name]; ok || last < 0 {
		icp.gvar[name] = val
	} else {
		icp.locals[last][name] = val
	}
	return nil
}

func (icp *implContext) hasCanceled() bool     { return icp.isCanceled }
func (icp *implContext) position(be *baseExpr) { icp.curBaseExpr = be }

func (icp *implContext) onError(err error) {
	if icp.recordError {
		icp.lastError = err
	} else {
		fmt.Fprintf(os.Stderr, "%s %s\n", icp.curBaseExpr.PosInfo(), err.Error())
		icp.isCanceled = true
	}
}

type forCookContext interface {
	cookContext
	reset()
	register(id string) (index int)
	unregister(index int)
	breakWith(id string) error
	continueWith(id string) error
	shouldBreak(ind int) bool
	currentLoop() int
	enterLoop(index int)
	exitLoop(index int)
}

type implForContext struct {
	*implContext
	loops         []string
	breakIndex    int
	continueIndex int
	loopStack     []int
}

func (it *implForContext) reset() { it.breakIndex, it.continueIndex = -1, -1 }

func (it *implForContext) currentLoop() int {
	if len(it.loopStack) > 0 {
		return it.loopStack[len(it.loopStack)-1]
	} else {
		return -1
	}
}

func (it *implForContext) enterLoop(index int) {
	it.loopStack = append(it.loopStack, index)
}

func (it *implForContext) exitLoop(index int) {
	if len(it.loopStack) > 0 {
		it.loopStack = it.loopStack[:len(it.loopStack)-1]
	}
}

func (it *implForContext) shouldBreak(index int) bool {
	return (it.breakIndex != -1 && it.breakIndex <= index) || (it.continueIndex != -1 && it.continueIndex < index)
}

func (it *implForContext) breakWith(id string) error {
	if id == "" {
		it.breakIndex = len(it.loops) - 1
	} else if i, err := it.isLoopExist(id); err != nil {
		return err
	} else {
		it.breakIndex = i
	}
	return nil
}

func (it *implForContext) continueWith(id string) error {
	if id == "" {
		it.continueIndex = len(it.loops) - 1
	} else if i, err := it.isLoopExist(id); err != nil {
		return err
	} else {
		it.continueIndex = i
	}
	return nil
}

func (it *implForContext) isLoopExist(id string) (int, error) {
	for i, n := range it.loops {
		if n == id {
			return i, nil
		}
	}
	return -1, fmt.Errorf("for name %s not found", id)
}

func (it *implForContext) register(id string) (index int) {
	it.loops = append(it.loops, id)
	return len(it.loops) - 1
}

func (it *implForContext) unregister(index int) {
	if 0 <= index && index < len(it.loops) {
		it.loops = append(it.loops[:index], it.loops[index+1:]...)
	} else {
		panic(fmt.Sprintf("loop index out of range %d, should be between 0, %d", index, len(it.loops)))
	}
}
