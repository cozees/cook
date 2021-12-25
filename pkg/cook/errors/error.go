package errors

import (
	"strings"
)

type CookError []error

func (ce *CookError) StackError(err error) { *ce = append(*ce, err) }

func (ce *CookError) Error() string {
	b := &strings.Builder{}
	b.WriteString("CookError:\n")
	for _, err := range *ce {
		b.WriteString("   ")
		b.WriteString(err.Error())
		b.WriteByte('\n')
	}
	return b.String()
}
