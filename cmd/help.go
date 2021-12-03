package main

import (
	"io"
	"os"

	"github.com/cozees/cook/pkg/runtime/args"
	"github.com/cozees/cook/pkg/runtime/function"
)

var mainFlags = &args.Flags{
	FuncName: "cook",
	Usage: `cook --VAR VALUE [TARGET ...]
			cook help [@FUNCTION]`,
	ShortDesc: `Cook interpreter to execute cookfile.`,
	Example: `cook --INPUT 1.32 sample_target
	          cook sample_target
			  cook help
			  cook`,
	Description: `Cook interpreter design to execute simple task defined in the a Cookfile.
				  Each task can be define as target which can be contain mutiple statement.
				  The goal of cook file is provide simple functionality for cross-platform
				  with simple syntax. You can use it to describe any task where  some  files
				  must be updated automatically from others whenever the others change.`,
}

const (
	helpDesc = `Print cook help to standard console if no function given otherwise print function help out instead.`
	varDesc  = `Define dynamic global variable via argument. By default, a dynamic global variable can be provided via
				environment variable however its a read-only variable. Variable define via argument is allowed to be
				change during execution.`
)

func PrintHelp(f *args.FunctionMeta) {
	if f != nil {
		rd := function.GetFunction(f.Name).Flags().HelpAsReader(false, "")
		io.Copy(os.Stdout, rd)
	} else {
		io.Copy(os.Stdout, mainFlags.HelpFlagVisitor(false, "", func(fw args.FlagWriter) {
			fw(12, "", "help", "", helpDesc)
			fw(12, "", "[VARIABLE]", "", varDesc)
		}))
	}
}
