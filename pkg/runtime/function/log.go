package function

import (
	"bytes"
	"flag"
	"fmt"
	"strings"
)

type printOption struct {
	strip  bool // strip whitespace before print for each argument
	omitLF bool // add newline, default yes
	echo   bool
	args   []string
	*options
}

func (po *printOption) reset() {
	po.strip = false
	po.omitLF = false
	po.echo = false
	po.options.args = nil
}

func (po *printOption) copy() interface{} {
	return &printOption{
		strip:  po.strip,
		omitLF: po.omitLF,
		echo:   po.echo,
		args:   po.options.args,
	}
}

func newPrintOption(fs *flag.FlagSet) Option {
	opts := &printOption{options: &options{}}
	opts.options.opts = opts
	opts.reset()
	flagBool(fs, &opts.omitLF, "n", opts.omitLF, "don't add newline after print all argument")
	flagBool(fs, &opts.echo, "e", opts.echo, "echo simply return a string of a printed result instead of printing it.")
	flagBool(fs, &opts.strip, "s", opts.strip, "strip all whilespace from each argument before print it")
	return opts
}

func init() {
	registerFunction(&GeneralFunction{
		name:     "print",
		flagInit: func(fs *flag.FlagSet) Option { return newPrintOption(fs) },
		handler: func(gf *GeneralFunction, i interface{}) (interface{}, error) {
			opts := i.(*printOption)
			buf := bytes.NewBufferString("")
			for i := range opts.args {
				if opts.strip {
					buf.WriteByte(' ')
					buf.WriteString(strings.TrimSpace(opts.args[i]))
				} else {
					buf.WriteByte(' ')
					buf.WriteString(opts.args[i])
				}
			}
			txt := buf.String()[1:]
			if opts.omitLF {
				if opts.echo {
					return txt, nil
				}
				fmt.Print(txt)
			} else {
				if opts.echo {
					return txt + "\n", nil
				}
				fmt.Println(txt)
			}
			return nil, nil
		},
	})
}
