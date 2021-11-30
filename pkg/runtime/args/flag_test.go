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
	Args  []string
}

const dummyDescription = `Lorem Ipsum is simply dummy text of the printing and typesetting industry.
Lorem Ipsum has been the industry's standard dummy text ever since the 1500s,
when an unknown printer took a galley of type and scrambled it to make a type specimen book.`

type argsCase struct {
	input   []string
	opts    interface{}
	failure bool
	err     error
}

type testFlags struct {
	*Flags
	cases []*argsCase
}

var testCases = []*testFlags{
	{
		Flags: &Flags{
			Flags: []*Flag{
				{Short: "a", Long: "flaga", Description: dummyDescription},
				{Short: "b", Long: "flagb", Description: dummyDescription},
				{Short: "c", Long: "flagc", Description: dummyDescription},
				{Short: "d", Long: "flagd", Description: dummyDescription},
				{Short: "e", Long: "flage", Description: dummyDescription},
				{Short: "f", Long: "flagf", Description: dummyDescription},
				{Short: "g", Long: "flagg", Description: dummyDescription},
			},
			Result:      reflect.TypeOf((*OptionsTest)(nil)).Elem(),
			Description: dummyDescription,
		},
		cases: []*argsCase{
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
					Args:  []string{"non-flag-or-options"},
				},
			},
			{
				input: []string{"--flaga", "discard text", "-a", "text", "non-flag-or-options"},
				opts: &OptionsTest{
					Flaga: "text",
					Args:  []string{"non-flag-or-options"},
				},
			},
			{
				input: []string{"-c", "873", "-a", "text", "non-flag-or-options"},
				opts: &OptionsTest{
					Flaga: "text",
					Flagc: int64(873),
					Args:  []string{"non-flag-or-options"},
				},
			},
			{
				input: []string{"-c", "12", "--flagc", "873", "-a", "text of text", "non-flag-or-options"},
				opts: &OptionsTest{
					Flaga: "text of text",
					Flagc: int64(873),
					Args:  []string{"non-flag-or-options"},
				},
			},
			{
				input: []string{"-d", "1.2", "non1", "non2", "--flage", "22", "-e", "42", "non3"},
				opts: &OptionsTest{
					Flagd: 1.2,
					Flage: []int64{22, 42},
					Args:  []string{"non1", "non2", "non3"},
				},
			},
			{
				input: []string{"--flagf", "99", "-f", "2.2", "-f", "text", "non3"},
				opts: &OptionsTest{
					Flagf: []interface{}{int64(99), 2.2, "text"},
					Args:  []string{"non3"},
				},
			},
			{
				input: []string{"--flagg", "99:99", "-g", "2.2:abc", "-g", "text:2.3", "non3"},
				opts: &OptionsTest{
					Flagg: map[string]interface{}{"99": int64(99), "2.2": "abc", "text": 2.3},
					Args:  []string{"non3"},
				},
			},
		},
	},
}

func TestFlag(t *testing.T) {
	for i, tc := range testCases {
		t.Logf("TestFlag case #%d", i+1)
		for si, stc := range tc.cases {
			t.Logf("\tTestFlag subcase #%d", si+1)
			opts, err := tc.Parse(stc.input)
			require.NoError(t, err)
			assert.Equal(t, stc.opts, opts)
		}
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
