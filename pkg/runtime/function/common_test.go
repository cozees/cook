package function

import (
	"reflect"

	"github.com/cozees/cook/pkg/runtime/args"
)

func convertToFunctionArgs(sargs []string) (fnArgs []*args.FunctionArg) {
	for _, arg := range sargs {
		fnArgs = append(fnArgs, &args.FunctionArg{
			Val:  arg,
			Kind: reflect.String,
		})
	}
	return
}
