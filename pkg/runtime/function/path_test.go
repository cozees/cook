package function

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pathInOut struct {
	name   string
	args   []string
	output interface{}
}

var pathTestCase []*pathInOut

func init() {
	var cdir, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	allGoFile, err := filepath.Glob("./*.go")
	if err != nil {
		panic(err)
	}
	pathTestCase = []*pathInOut{
		{
			name:   "pabs",
			args:   []string{"test/abc/text.txt"},
			output: fmt.Sprintf("%s/test/abc/text.txt", cdir),
		},
		{
			name:   "pabs",
			args:   []string{"/usr/abc/text.txt"},
			output: "/usr/abc/text.txt",
		},
		{
			name:   "pbase",
			args:   []string{"test/aa/bb"},
			output: "bb",
		},
		{
			name:   "pbase",
			args:   []string{"test/aa/bb.txt"},
			output: "bb.txt",
		},
		{
			name:   "pext",
			args:   []string{"test/aa/bb"},
			output: "",
		},
		{
			name:   "pext",
			args:   []string{"test/aa/bb.txt"},
			output: ".txt",
		},
		{
			name:   "pdir",
			args:   []string{"test/aa/bb.txt"},
			output: "test/aa",
		},
		{
			name:   "pclean",
			args:   []string{"abc/two/../test/./aa/bb.txt"},
			output: "abc/test/aa/bb.txt",
		},
		{
			name:   "psplit",
			args:   []string{"abc/two/../test/./aa/bb.txt"},
			output: []string{"abc", "two", "..", "test", ".", "aa", "bb.txt"},
		},
		{
			name:   "prel",
			args:   []string{"/test/abc/test", "abc/two/../test/./aa/bb.txt"},
			output: nil,
		},
		{
			name:   "prel",
			args:   []string{"/test/abc/test", "/test/abc/two/../test/./aa/bb.txt"},
			output: "aa/bb.txt",
		},
		{
			name:   "pglob",
			args:   []string{"./*.go"},
			output: allGoFile,
		},
	}
}

func TestPathFunction(t *testing.T) {
	for i, tc := range pathTestCase {
		t.Logf("TestPath function %s case #%d", tc.name, i+1)
		fn := GetFunction(tc.name)
		result, err := fn.Apply(tc.args)
		if tc.output == nil {
			assert.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.Equal(t, tc.output, result)
		}
	}
}
