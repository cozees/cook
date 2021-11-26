package function

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type pathOptions struct {
	args      []string
	argNeeded int
	*options
}

func (po *pathOptions) reset() {
	po.options.args = nil
}

func (po *pathOptions) copy() interface{} {
	return &pathOptions{
		args:      po.options.args,
		argNeeded: po.options.argNeeded,
	}
}

func (po *pathOptions) validate(gf *GeneralFunction, fn func(...string) (interface{}, error)) (interface{}, error) {
	if len(po.args) != po.argNeeded {
		return "", fmt.Errorf("%s require %d arugment file path", gf.name, po.argNeeded)
	}
	return fn(po.args...)
}

func newPathOptions(argNeeded int) Option {
	opts := &pathOptions{options: &options{argNeeded: argNeeded}}
	opts.options.opts = opts
	return opts
}

func dHandler(gf *GeneralFunction, i interface{}, fn func(string) (string, error)) (interface{}, error) {
	return i.(*pathOptions).validate(gf, func(s ...string) (interface{}, error) { return fn(s[0]) })
}

func sHandler(gf *GeneralFunction, i interface{}, fn func(string) string) (interface{}, error) {
	return i.(*pathOptions).validate(gf, func(s ...string) (interface{}, error) { return fn(s[0]), nil })
}

func init() {
	registerFunction(&GeneralFunction{
		name:     "pabs",
		flagInit: func(fs *flag.FlagSet) Option { return newPathOptions(1) },
		handler:  func(gf *GeneralFunction, i interface{}) (interface{}, error) { return dHandler(gf, i, filepath.Abs) },
	})

	registerFunction(&GeneralFunction{
		name:     "pbase",
		flagInit: func(fs *flag.FlagSet) Option { return newPathOptions(1) },
		handler:  func(gf *GeneralFunction, i interface{}) (interface{}, error) { return sHandler(gf, i, filepath.Base) },
	})

	registerFunction(&GeneralFunction{
		name:     "pext",
		flagInit: func(fs *flag.FlagSet) Option { return newPathOptions(1) },
		handler:  func(gf *GeneralFunction, i interface{}) (interface{}, error) { return sHandler(gf, i, filepath.Ext) },
	})

	registerFunction(&GeneralFunction{
		name:     "pdir",
		flagInit: func(fs *flag.FlagSet) Option { return newPathOptions(1) },
		handler:  func(gf *GeneralFunction, i interface{}) (interface{}, error) { return sHandler(gf, i, filepath.Dir) },
	})

	registerFunction(&GeneralFunction{
		name:     "pclean",
		flagInit: func(fs *flag.FlagSet) Option { return newPathOptions(1) },
		handler:  func(gf *GeneralFunction, i interface{}) (interface{}, error) { return sHandler(gf, i, filepath.Clean) },
	})

	registerFunction(&GeneralFunction{
		name:     "psplit",
		flagInit: func(fs *flag.FlagSet) Option { return newPathOptions(1) },
		handler: func(gf *GeneralFunction, i interface{}) (interface{}, error) {
			return i.(*pathOptions).validate(gf, func(s ...string) (interface{}, error) {
				return strings.Split(s[0], fmt.Sprintf("%c", os.PathSeparator)), nil
			})
		},
	})

	registerFunction(&GeneralFunction{
		name:     "pglob",
		flagInit: func(fs *flag.FlagSet) Option { return newPathOptions(1) },
		handler: func(gf *GeneralFunction, i interface{}) (interface{}, error) {
			return i.(*pathOptions).validate(gf, func(s ...string) (interface{}, error) {
				return filepath.Glob(s[0])
			})
		},
	})

	registerFunction(&GeneralFunction{
		name:     "prel",
		flagInit: func(fs *flag.FlagSet) Option { return newPathOptions(2) },
		handler: func(gf *GeneralFunction, i interface{}) (interface{}, error) {
			return i.(*pathOptions).validate(gf, func(s ...string) (interface{}, error) {
				return filepath.Rel(s[0], s[1])
			})
		},
	})
}
