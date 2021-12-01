package args

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type OptionsTest struct {
	Flaga string                 `flag:"flaga"`
	Flagb bool                   `flag:"flagb"`
	Flagc int64                  `flag:"flagc"`
	Flagd float64                `flag:"flagd"`
	Flage []int64                `flag:"flage"`
	Flagf []interface{}          `flag:"flagf"`
	Flagg map[string]interface{} `flag:"flagg"`
	Args  []interface{}
}

const dummyDescription = `Lorem Ipsum is simply dummy text of the printing and typesetting industry.
Lorem Ipsum has been the industry's standard dummy text ever since the 1500s,
when an unknown printer took a galley of type and scrambled it to make a type specimen book.`

var testFlags = &Flags{
	Flags: []*Flag{
		{Short: "a", Long: "flaga"},
		{Short: "b", Long: "flagb"},
		{Short: "c", Long: "flagc"},
		{Short: "d", Long: "flagd"},
		{Short: "e", Long: "flage"},
		{Short: "f", Long: "flagf"},
		{Short: "g", Long: "flagg"},
	},
	Result: reflect.TypeOf((*OptionsTest)(nil)).Elem(),
}

type argsCase struct {
	input   []string
	opts    interface{}
	failure bool
	err     error
}

var testCases = []*argsCase{
	{
		input: []string{"-b", "-a", "text"},
		opts: &OptionsTest{
			Flagb: true,
			Flaga: "text",
		},
	},
	{
		input: []string{"--flagb", "-a", "text", "non-flag-or-options"},
		opts: &OptionsTest{
			Flagb: true,
			Flaga: "text",
			Args:  []interface{}{"non-flag-or-options"},
		},
	},
	{
		input: []string{"--flaga", "discard text", "-a", "text", "non-flag-or-options"},
		opts: &OptionsTest{
			Flaga: "text",
			Args:  []interface{}{"non-flag-or-options"},
		},
	},
	{
		input: []string{"-c", "873", "-a", "text", "non-flag-or-options"},
		opts: &OptionsTest{
			Flaga: "text",
			Flagc: int64(873),
			Args:  []interface{}{"non-flag-or-options"},
		},
	},
	{
		input: []string{"-c", "12", "--flagc", "873", "-a", "text of text", "non-flag-or-options"},
		opts: &OptionsTest{
			Flaga: "text of text",
			Flagc: int64(873),
			Args:  []interface{}{"non-flag-or-options"},
		},
	},
	{
		input: []string{"-d", "1.2", "non1", "non2", "--flage", "22", "-e", "42", "non3"},
		opts: &OptionsTest{
			Flagd: 1.2,
			Flage: []int64{22, 42},
			Args:  []interface{}{"non1", "non2", "non3"},
		},
	},
	{
		input: []string{"--flagf", "99", "-f", "2.2", "-f", "text", "non3"},
		opts: &OptionsTest{
			Flagf: []interface{}{int64(99), 2.2, "text"},
			Args:  []interface{}{"non3"},
		},
	},
	{
		input: []string{"--flagg", "99:99", "-g", "2.2:abc", "-g", "text:2.3", "non3"},
		opts: &OptionsTest{
			Flagg: map[string]interface{}{"99": int64(99), "2.2": "abc", "text": 2.3},
			Args:  []interface{}{"non3"},
		},
	},
}

func TestFlag(t *testing.T) {
	for i, tc := range testCases {
		t.Logf("TestFlag case #%d", i+1)
		opts, err := testFlags.Parse(tc.input)
		require.NoError(t, err)
		assert.Equal(t, tc.opts, opts)
	}
}

type testFnFlag struct {
	input []*FunctionArg
	opts  interface{}
}

var testFnCases = []*testFnFlag{
	{
		input: []*FunctionArg{
			{val: "-b", kind: reflect.String},
			{val: "-a", kind: reflect.String},
			{val: "text", kind: reflect.String},
		},
		opts: &OptionsTest{
			Flagb: true,
			Flaga: "text",
		},
	},
	{
		input: []*FunctionArg{
			{val: "--flagb", kind: reflect.String},
			{val: "-a", kind: reflect.String},
			{val: "text", kind: reflect.String},
			{val: "non-flag-or-options", kind: reflect.String},
		},
		opts: &OptionsTest{
			Flagb: true,
			Flaga: "text",
			Args:  []interface{}{"non-flag-or-options"},
		},
	},
	{
		input: []*FunctionArg{
			{val: "--flaga", kind: reflect.String},
			{val: "discard text", kind: reflect.String},
			{val: "-a", kind: reflect.String},
			{val: "text", kind: reflect.String},
			{val: "non-flag-or-options", kind: reflect.String},
		},
		opts: &OptionsTest{
			Flaga: "text",
			Args:  []interface{}{"non-flag-or-options"},
		},
	},
	{
		input: []*FunctionArg{
			{val: "-c", kind: reflect.String},
			{val: int64(873), kind: reflect.Int64},
			{val: "-a", kind: reflect.String},
			{val: "text", kind: reflect.String},
			{val: "non-flag-or-options", kind: reflect.String},
		},
		opts: &OptionsTest{
			Flaga: "text",
			Flagc: int64(873),
			Args:  []interface{}{"non-flag-or-options"},
		},
	},
	{
		input: []*FunctionArg{
			{val: "-c", kind: reflect.String},
			{val: int64(12), kind: reflect.Int64},
			{val: "--flagc", kind: reflect.String},
			{val: int64(873), kind: reflect.Int64},
			{val: "-a", kind: reflect.String},
			{val: "text of text", kind: reflect.String},
			{val: "non-flag-or-options", kind: reflect.String},
		},
		opts: &OptionsTest{
			Flaga: "text of text",
			Flagc: int64(873),
			Args:  []interface{}{"non-flag-or-options"},
		},
	},
	{
		input: []*FunctionArg{
			{val: "-d", kind: reflect.String},
			{val: 1.2, kind: reflect.Float64},
			{val: "non1", kind: reflect.String},
			{val: "non2", kind: reflect.String},
			{val: "--flage", kind: reflect.String},
			{val: int64(22), kind: reflect.Int64},
			{val: "-e", kind: reflect.String},
			{val: int64(42), kind: reflect.Int64},
			{val: "non3", kind: reflect.String},
		},
		opts: &OptionsTest{
			Flagd: 1.2,
			Flage: []int64{22, 42},
			Args:  []interface{}{"non1", "non2", "non3"},
		},
	},
	{
		input: []*FunctionArg{
			{val: "--flagf", kind: reflect.String},
			{val: int64(99), kind: reflect.Int64},
			{val: "-f", kind: reflect.String},
			{val: 2.2, kind: reflect.Float64},
			{val: "-f", kind: reflect.String},
			{val: "text", kind: reflect.String},
			{val: "non3", kind: reflect.String},
		},
		opts: &OptionsTest{
			Flagf: []interface{}{int64(99), 2.2, "text"},
			Args:  []interface{}{"non3"},
		},
	},
	{
		input: []*FunctionArg{
			{val: "--flagg", kind: reflect.String},
			{val: map[string]int64{"99": int64(99)}, kind: reflect.Map},
			{val: "-g", kind: reflect.String},
			{val: map[string]string{"2.2": "abc"}, kind: reflect.Map},
			{val: "-g", kind: reflect.String},
			{val: "text:2.3", kind: reflect.String},
			{val: "non3", kind: reflect.String},
		},
		opts: &OptionsTest{
			Flagg: map[string]interface{}{"99": int64(99), "2.2": "abc", "text": 2.3},
			Args:  []interface{}{"non3"},
		},
	},
}

func TestFuncFlagParsing(t *testing.T) {
	for i, tc := range testFnCases {
		t.Logf("TestFnParsing case #%d", i+1)
		opts, err := testFlags.ParseFunctionArgs(tc.input)
		require.NoError(t, err)
		assert.Equal(t, tc.opts, opts)
	}
}

var testMainFlagCases = []*argsCase{
	{
		input: []string{"--name", "nyu", "target"},
		opts: &MainOptions{
			Cookfile: defaultCookfile,
			Args: map[string]interface{}{
				"name": "nyu",
			},
			Targets: []string{"target"},
		},
	},
	{
		input: []string{"--age:i", "32", "--age:i", "34", "--height:f", "1.57"},
		opts: &MainOptions{
			Cookfile: defaultCookfile,
			Args: map[string]interface{}{
				"age":    []interface{}{int64(32), int64(34)},
				"height": 1.57,
			},
		},
	},
	{
		input: []string{"--age:s", "32", "--age:i", "34", "--info:a:i", "3.21:33", "--info:s:i", "8.23:33", "--info:s:a", "bb:32.1"},
		opts: &MainOptions{
			Cookfile: defaultCookfile,
			Args: map[string]interface{}{
				"age": []interface{}{"32", int64(34)},
				"info": map[interface{}]interface{}{
					3.21:   int64(33),
					"8.23": int64(33),
					"bb":   32.1,
				},
			},
		},
	},
	{
		input: []string{"sample1", "sample2", "-c", "Cooksample", "--name=test", "--info:i=123", "--lerp:a", "3.21"},
		opts: &MainOptions{
			Cookfile: "Cooksample",
			Args: map[string]interface{}{
				"name": "test",
				"info": int64(123),
				"lerp": 3.21,
			},
			Targets: []string{"sample1", "sample2"},
		},
	},
	// test error
	{
		input:   []string{"--dict:a", "22", "--dict:i:s", "11:aa"},
		failure: true,
	},
	{
		input:   []string{"--dict:a:a", "22"},
		failure: true,
	},
	{
		input:   []string{"--val:i", "22", "9038"},
		failure: true,
	},
	{
		input:   []string{"--dict:o", "22"},
		failure: true,
		err:     ErrAllowFormat,
	},
	{
		input:   []string{"--dict:", "22"},
		failure: true,
		err:     ErrFlagSyntax,
	},
}

func TestMainFlag(t *testing.T) {
	for i, tc := range testMainFlagCases {
		t.Logf("TestMainFlag case #%d", i+1)
		opts, err := ParseMainArgument(tc.input)
		if tc.failure {
			if tc.err != nil {
				assert.ErrorIs(t, tc.err, err)
			} else {
				assert.Error(t, err)
			}
		} else {
			require.NoError(t, err)
			assert.Equal(t, tc.opts, opts)
		}
	}
}

type SampleOptions struct {
	Echo   bool   `flag:"echo"`
	Format bool   `flag:"format"`
	Extra  bool   `flag:"extra"`
	Output string `flag:"out"`
	Args   []string
}

var flags = &Flags{
	FuncName: "sample",
	Flags: []*Flag{
		{Short: "e", Long: "echo", Description: dummyDescription},
		{Short: "f", Long: "format", Description: dummyDescription},
		{Short: "x", Long: "extra", Description: dummyDescription},
		{Long: "out", Description: dummyDescription},
	},
	Result:      reflect.TypeOf((*SampleOptions)(nil)),
	ShortDesc:   "Lorem Ipsum is simply dummy",
	Usage:       "sample [-efx] [--out name] file [file ...]",
	Example:     "sample -e sample.json",
	Description: dummyDescription,
}

var flagMarkdown = `
## @sample

Usage:
` + "```" + `cook
sample [-efx] [--out name] file [file ...]
` + "```" + `

Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -e, --echo | false | Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. |
| -f, --format | false | Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. |
| -x, --extra | false | Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. |
| --out value | "" | Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. |

Example:

` + "```" + `cook
sample -e sample.json
` + "```" + `
`

func TestMarkdownDoc(t *testing.T) {
	result := flags.Help(true)
	assert.Equal(t, flagMarkdown[1:], result)
}

var flagsConsole = `
NAME
    sample -- Lorem Ipsum is simply dummy

USAGE
    sample [-efx] [--out name] file [file ...]

DESCRIPTION
    Lorem Ipsum is simply dummy text of the printing and typesetting industry.
    Lorem Ipsum has been the industry's standard dummy text ever since the
    1500s, when an unknown printer took a galley of type and scrambled it to
    make a type specimen book.

Available options or flag:

    -e, --echo        Lorem Ipsum is simply dummy text of the printing and
                      typesetting industry. Lorem Ipsum has been the
                      industry's standard dummy text ever since the 1500s,
                      when an unknown printer took a galley of type and
                      scrambled it to make a type specimen book.

    -f, --format      Lorem Ipsum is simply dummy text of the printing and
                      typesetting industry. Lorem Ipsum has been the
                      industry's standard dummy text ever since the 1500s,
                      when an unknown printer took a galley of type and
                      scrambled it to make a type specimen book.

    -x, --extra       Lorem Ipsum is simply dummy text of the printing and
                      typesetting industry. Lorem Ipsum has been the
                      industry's standard dummy text ever since the 1500s,
                      when an unknown printer took a galley of type and
                      scrambled it to make a type specimen book.

    --out             Lorem Ipsum is simply dummy text of the printing and
                      typesetting industry. Lorem Ipsum has been the
                      industry's standard dummy text ever since the 1500s,
                      when an unknown printer took a galley of type and
                      scrambled it to make a type specimen book.

EXAMPLE
    sample -e sample.json
`

func TestConsoleDoc(t *testing.T) {
	result := flags.Help(false)
	assert.Equal(t, flagsConsole[1:], result)
}
