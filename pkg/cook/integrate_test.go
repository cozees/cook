package cook

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/cozees/cook/pkg/cook/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type integrateTestCase struct {
	name   string
	vname  string
	output string
}

var cases = []*integrateTestCase{
	{
		name:  "file1.txt",
		vname: "FILE1",
		output: `
0 2021 3.36.0
1 2020 3.32.0
2 2015 3.9.2
3 2012 3.1.2.1
data1 data2 data3
0 https://www.sqlite.org/2021/sqlite-amalgamation-3360000.zip
1 https://www.sqlite.org/2020/sqlite-amalgamation-3320000.zip
2 https://www.sqlite.org/2015/sqlite-amalgamation-3090200.zip
3 https://www.sqlite.org/2012/sqlite-amalgamation-3010201.zip
`,
	},
	{
		name:  "file2.txt",
		vname: "FILE2",
		output: `
123 abc
finalize executed
`,
	},
}

func cleanup() {
	for _, tc := range cases {
		os.Remove(tc.name)
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	cleanup()
	os.Exit(code)
}

func TestCookProgram(t *testing.T) {
	p := parser.NewParser()
	cook, err := p.Parse("testdata/Cookfile")
	require.NoError(t, err)
	args := make(map[string]string)
	for _, tc := range cases {
		args[tc.vname] = tc.name
	}
	cook.Execute(args)
	for i, tc := range cases {
		t.Logf("TestCookProgram case #%d", i+1)
		bo, err := ioutil.ReadFile(tc.name)
		assert.NoError(t, err)
		assert.Equal(t, tc.output, string(bo))
	}
}
