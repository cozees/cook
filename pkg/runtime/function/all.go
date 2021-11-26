package function

import (
	"bytes"
	"flag"
	"fmt"
	"sync"
)

type Option interface {
	Parse(fs *flag.FlagSet, args []string) (interface{}, error)
	reset()
	copy() interface{}
}

type options struct {
	opts      Option
	args      []string
	argNeeded int
	lock      sync.RWMutex
}

func (o *options) Parse(fs *flag.FlagSet, args []string) (i interface{}, err error) {
	o.lock.Lock()
	defer func() {
		o.opts.reset()
		o.lock.Unlock()
	}()
	if err = fs.Parse(args); err != nil {
		return
	}
	o.args = fs.Args()
	return o.opts.copy(), nil
}

func flagBool(fs *flag.FlagSet, v *bool, name string, def bool, desc string, alias ...string) {
	fs.BoolVar(v, name, def, desc)
	for _, a := range alias {
		fs.BoolVar(v, a, def, desc)
	}
}

func flagString(fs *flag.FlagSet, v *string, name, def, desc string, alias ...string) {
	fs.StringVar(v, name, def, desc)
	for _, a := range alias {
		fs.StringVar(v, a, def, desc)
	}
}

type FuncHandler func(*GeneralFunction, interface{}) (interface{}, error)
type FlagInitHandler func(*flag.FlagSet) Option

type Function interface {
	Apply([]string) (interface{}, error)
	Name() string
	Alias() []string
	Help(name string) string
}

// store function reference by name
var funcStore map[string]Function = make(map[string]Function)

func IsExist(name string) bool         { return funcStore[name] != nil }
func GetFunction(name string) Function { return funcStore[name] }

func registerFunction(f Function) {
	funcStore[f.Name()] = f
	if alias := f.Alias(); alias != nil {
		for _, a := range alias {
			funcStore[a] = f
		}
	}
}

type GeneralFunction struct {
	name    string
	alias   []string
	handler FuncHandler

	flagSet  *flag.FlagSet
	flagInit FlagInitHandler
	option   Option
	flagOnce sync.Once
}

func (gf *GeneralFunction) Name() string    { return gf.name }
func (gf *GeneralFunction) Alias() []string { return gf.alias }

func (gf *GeneralFunction) ensureFlag() {
	gf.flagOnce.Do(func() {
		gf.flagSet = flag.NewFlagSet(gf.name, flag.ExitOnError)
		gf.option = gf.flagInit(gf.flagSet)
	})
}

func (gf *GeneralFunction) Help(name string) string {
	gf.ensureFlag()
	buffer := bytes.NewBufferString(fmt.Sprintf("Usage of %s:\n", name))
	gf.flagSet.SetOutput(buffer)
	gf.flagSet.PrintDefaults()
	return buffer.String()
}

func (gf *GeneralFunction) Apply(args []string) (interface{}, error) {
	gf.ensureFlag()
	i, err := gf.option.Parse(gf.flagSet, args)
	if err != nil {
		return nil, err
	}
	return gf.handler(gf, i)
}
