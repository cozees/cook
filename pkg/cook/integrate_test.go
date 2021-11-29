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

const nestedResult = "nasted-result.txt"

// 0: LIST, 1: MAP, 2: result
var nestLoopTestCase = [][]interface{}{
	{
		[]interface{}{int64(1), int64(2), int64(3)},
		[]interface{}{int64(1), int64(2), 1.3},
		"392 32",
	},
	{
		[]interface{}{int64(11), int64(4), int64(9)},
		[]interface{}{22.4, int64(2), 1.3},
		"108 33",
	},
	{
		[]interface{}{int64(11), int64(5), int64(12)},
		[]interface{}{true, false, 1.3},
		"348 34",
	},
	{
		[]interface{}{int64(9), int64(15), int64(122)},
		[]interface{}{1.3, int64(40), false},
		"192 34",
	},
	{
		[]interface{}{int64(5), int64(7), int64(32)},
		[]interface{}{24.8, int64(4), int64(40)},
		"85 28",
	},
}

func cleanup() {
	for _, tc := range cases {
		os.Remove(tc.name)
	}
	os.Remove(nestedResult)
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
	args := make(map[string]interface{})
	args["TEST_NEST_LOOP"] = false
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

	// test nested loop
	args["TEST_NEST_LOOP"] = true
	for i, tc := range nestLoopTestCase {
		t.Logf("TestCookProgram Nested Loop case #%d", i+1)
		args["LIST"] = tc[0]
		args["LISTA"] = tc[1]
		args["FILE1"] = nestedResult
		cook.ExecuteWithTarget(args, "sampleNestLoop")
		bo, err := ioutil.ReadFile(nestedResult)
		assert.NoError(t, err)
		assert.Equal(t, tc[2], string(bo))
	}
}
