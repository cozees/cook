package args

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
| --out | "" | Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. |

Example:

` + "```" + `cook
sample -e sample.json
` + "```" + `
`

func TestMarkdownDoc(t *testing.T) {
	result := flags.Help(true, "")
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
	result := flags.Help(false, "")
	assert.Equal(t, flagsConsole[1:], result)
}
