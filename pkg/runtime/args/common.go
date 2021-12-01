package args

import (
	"reflect"
	"strings"
)

// integer, float, string, boolean, map
const typeCharacters = "ifsba"

var typeMatch = []reflect.Kind{
	reflect.Int64,
	reflect.Float64,
	reflect.String,
	reflect.Bool,
	reflect.Interface,
}

func checkFormat(a string, ignore bool) (k reflect.Kind, err error) {
	if len(a) == 0 {
		if !ignore {
			err = ErrFlagSyntax
		}
	} else if i := strings.IndexByte(typeCharacters, a[0]); i == -1 {
		err = ErrAllowFormat
	} else {
		k = typeMatch[i]
	}
	return
}

func parseFlagFormat(fs string) (o string, p, s reflect.Kind, err error) {
	if len(fs) == 1 && fs == ":" {
		err = ErrFlagSyntax
		return
	}
	i := strings.IndexByte(fs, ':')
	if i > 0 {
		ps, ss := "", ""
		ps, o = fs[i+1:], fs[:i]
		if i = strings.IndexByte(ps, ':'); i == 0 {
			err = ErrFlagSyntax
			return
		} else if i > 0 {
			ss, ps = ps[i+1:], ps[:i]
		}
		if p, err = checkFormat(ps, false); err == nil {
			s, err = checkFormat(ss, true)
		}
	} else {
		o = fs
		p = reflect.String
	}
	return
}

func lower(ch byte) byte { return ('a' - 'A') | ch }