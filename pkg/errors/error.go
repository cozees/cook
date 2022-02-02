package errors

import (
	"io/fs"
	"strconv"
	"strings"
)

// CookError implement Go error interface type which present an error in plain text
// in the sense of Cook context. It strip out any Go error info.
type CookError []error

func (ce *CookError) StackError(err error) { *ce = append(*ce, err) }

func (ce *CookError) Error() string {
	b := &strings.Builder{}
	b.WriteString("CookError:\n")
	for _, err := range *ce {
		b.WriteString("   ")
		switch v := err.(type) {
		case *strconv.NumError:
			b.WriteString("parsing " + strconv.Quote(v.Num) + ": " + v.Err.Error())
		case *fs.PathError:
			b.WriteString("path " + strconv.Quote(v.Path) + " " + v.Err.Error())
		default:
			msg := err.Error()
			// replace any potential error message that contain go package
			msg = strings.Replace(msg, "archive/tar: ", "", 1)
			msg = strings.Replace(msg, "archive/zip: ", "", 1)
			b.WriteString(msg)
		}
		b.WriteByte('\n')
	}
	return b.String()
}
