package function

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type caseInOut struct {
	args   []string
	output interface{}
}

var splitCase = []*caseInOut{
	{
		args:   []string{"-by", ".", "53.24.0"},
		output: []interface{}{"53", "24", "0"},
	},
	{
		args: []string{"-S", "-L", "A yo-yo is a toy consisting of an axle connected to two disks"},
		output: []interface{}{
			[]interface{}{"A", "yo-yo", "is", "a", "toy", "consisting", "of", "an", "axle", "connected", "to", "two", "disks"},
		},
	},
	{
		args: []string{"-S", "-L", "A yo-yo is a toy consisting of\nan axle connected to two disks"},
		output: []interface{}{
			[]interface{}{"A", "yo-yo", "is", "a", "toy", "consisting", "of"},
			[]interface{}{"an", "axle", "connected", "to", "two", "disks"},
		},
	},
	{
		args: []string{"-by", ",*", "-L", "A yo-yo is,* a toy consisting of\nan axle connected,* to two disks"},
		output: []interface{}{
			[]interface{}{"A yo-yo is", " a toy consisting of"},
			[]interface{}{"an axle connected", " to two disks"},
		},
	},
	{
		args:   []string{"-by", ",*", "-L", "-rc", "1:0", "A yo-yo is,* a toy consisting of\nan axle connected,* to two disks"},
		output: "an axle connected",
	},
	{
		args:   []string{"-S", "-L", "-rc", "1:0", "A yo-yo is a toy consisting of\nan axle connected to two disks"},
		output: "an",
	},
	{
		args:   []string{"-S", "-L", "-rc", "1", "A yo-yo is a toy consisting of\nan axle connected to two disks"},
		output: []interface{}{"an", "axle", "connected", "to", "two", "disks"},
	},
	{
		args:   []string{"-S", "-L", "-rc", ":1", "A yo-yo is a toy consisting of\nan axle connected to two disks"},
		output: []interface{}{"yo-yo", "axle"},
	},
	{
		args:   []string{"-reg", "\\s", "A yo-yo is a toy consisting of\nan axle connected to two disks"},
		output: []string{"A", "yo-yo", "is", "a", "toy", "consisting", "of", "an", "axle", "connected", "to", "two", "disks"},
	},
}

func TestSplit(t *testing.T) {
	fn := GetFunction("ssplit")
	for i, tc := range splitCase {
		t.Logf("TestSplit case #%d", i+1)
		result, err := fn.Apply(tc.args)
		if tc.output == nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.output, result)
		}
	}
}

var paddingCase = []*caseInOut{
	{
		args:   []string{"-l", "2", "-by", "0", "ax"},
		output: "00ax",
	},
	{
		args:   []string{"-l", "2", "-by", "0", "ax", "ko"},
		output: []string{"00ax", "00ko"},
	},
	{
		args:   []string{"-l", "1", "-r", "3", "-by", "0", "ax"},
		output: "0ax000",
	},
	{
		args:   []string{"-l", "1", "-r", "3", "-by", "0", "ax", "kl", "long text"},
		output: []string{"0ax000", "0kl000", "0long text000"},
	},
	{
		args:   []string{"-l", "1", "-r", "3", "-by", "0", "-m", "10", "ax"},
		output: "0ax000",
	},
	{
		args:   []string{"-l", "1", "-r", "3", "-by", "0", "-m", "6", "ax"},
		output: "0ax000",
	},
	{
		args:   []string{"-l", "1", "-r", "3", "-by", "0", "-m", "5", "ax"},
		output: "0ax00",
	},
	{
		args:   []string{"-l", "1", "-by", "0", "-m", "2", "ax"},
		output: "ax",
	},
	{
		args:   []string{"-l", "1", "-by", "0", "-m", "2", "a"},
		output: "0a",
	},
	{
		args:   []string{"-l", "1", "-by", "0", "-m", "3", "ax"},
		output: "0ax",
	},
}

func TestPadding(t *testing.T) {
	fn := GetFunction("spad")
	for i, tc := range paddingCase {
		t.Logf("TestPadding case #%d", i+1)
		result, err := fn.Apply(tc.args)
		if tc.output == nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.output, result)
		}
	}
}

const content = `At vero eos et accusamus et iusto odio dignissimos ducimus qui blanditiis praesentium voluptatum 
deleniti atque corrupti quos dolores et quas molestias excepturi sint occaecati cupiditate non provident, 
similique sunt in culpa qui officia deserunt mollitia animi, id est laborum et dolorum fuga. Et harum quidem 
rerum facilis est et expedita distinctio. Nam libero tempore, cum soluta nobis est eligendi optio cumque nihil 
impedit quo minus id quod maxime placeat facere possimus, omnis voluptas assumenda est, omnis dolor repellendus. 
Temporibus autem quibusdam et aut officiis debitis aut rerum necessitatibus saepe eveniet ut et voluptates 
repudiandae sint et molestiae non recusandae. Itaque earum rerum hic tenetur a sapiente delectus, ut aut 
reiciendis voluptatibus maiores alias consequatur aut perferendis doloribus asperiores repellat.`

var replaceCase = []*caseInOut{
	{
		args:   []string{"abc", "xyz", "At vero eos et accusamus et iusto odio"},
		output: "At vero eos et accusamus et iusto odio",
	},
	{
		args:   []string{"eos", "xyz", "At vero eos et accusamus et iusto odio"},
		output: "At vero xyz et accusamus et iusto odio",
	},
	{
		args:   []string{".t", "xyz", "At vero eos et accusamus et iusto odio"},
		output: "At vero eos et accusamus et iusto odio",
	},
	{
		args:   []string{"-e", ".t", "xyz", "At vero eos et accusamus et iusto odio"},
		output: "xyz vero eos xyz accusamus xyz iuxyzo odio",
	},
	{
		args:   []string{"-e", ".(t)", "x${1}yz", "At vero eos et accusamus et iusto odio"},
		output: "xtyz vero eos xtyz accusamus xtyz iuxtyzo odio",
	},
}

func getlines(lines string) []int {
	if lines == "" {
		return nil
	}
	var repline []int
	for _, lstr := range strings.Split(lines, ",") {
		i, _ := strconv.ParseInt(lstr, 10, 32)
		repline = append(repline, int(i))
	}
	sort.Ints(repline)
	return repline
}

func replaceAllAt(src, search, replace, lines string) string {
	if lines == "" {
		return strings.ReplaceAll(src, search, replace)
	} else {
		repline := getlines(lines)
		var result []byte
		line := 1
		buf := bytes.NewBufferString(src)
		for ls, err := buf.ReadString('\n'); true; ls, err = buf.ReadString('\n') {
			if ls != "" {
				if len(repline) > 0 && repline[0] == line {
					result = append(result, strings.ReplaceAll(ls, search, replace)...)
					repline = repline[1:]
				} else {
					result = append(result, ls...)
				}
				line++
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
		}
		return string(result)
	}
}

func replaceRegExpAllAt(buf *bufio.Reader, search, replace, lines string) string {
	repline := getlines(lines)
	reg := regexp.MustCompile(search)
	result := ""
	line := 1
	for {
		bline, err := buf.ReadString('\n')
		if bline != "" {
			switch {
			case len(repline) > 0 && repline[0] == line:
				repline = repline[1:]
				fallthrough
			case repline == nil:
				result += reg.ReplaceAllString(bline, replace)
			default:
				result += bline
			}
			line++
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
	}
	return result
}

func TestReplace(t *testing.T) {
	fn := GetFunction("sreplace")
	for i, tc := range replaceCase {
		t.Logf("TestReplace case #%d", i+1)
		out, err := fn.Apply(tc.args)
		if tc.output == nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.output, out)
		}
	}
	// test file
	f := "string-sample.txt"
	defer os.Remove(f)
	as := strings.Split(content, " ")
	num := len(as)
	count := 0
	buf := bytes.NewBufferString("")
	for count < 5242880 {
		word := as[rand.Intn(num)]
		count += len(word)
		_, err := buf.WriteString(word)
		assert.NoError(t, err)
	}
	// test non regular expression
	ofile := "result-" + f
	testFileArgs := [][]string{
		{"accusamus", "aac", "@" + f},
		{"-L", "12,35", "accusamus", "aac", "@" + f},
		{"-e", "et.?", "*${1}*", "@" + f, "@" + ofile},
		{"-e", "-L", "22,13,32,14", "et.?", "*${1}*", "@" + f, "@" + ofile},
	}
	defer os.Remove(ofile)
	for _, args := range testFileArgs {
		require.NoError(t, ioutil.WriteFile(f, buf.Bytes(), 0700))
		out, err := fn.Apply(args)
		assert.NoError(t, err)
		assert.Nil(t, out)
		bout, err := ioutil.ReadFile(args[len(args)-1][1:])
		assert.NoError(t, err)
		var testResult = string(bout)
		var expectStr string
		if args[0] != "-e" {
			if args[0] == "-L" {
				expectStr = replaceAllAt(buf.String(), args[2], args[3], args[1])
			} else {
				expectStr = replaceAllAt(buf.String(), args[0], args[1], "")
			}
		} else {
			if args[1] == "-L" {
				expectStr = replaceRegExpAllAt(bufio.NewReader(bytes.NewReader(buf.Bytes())), args[3], args[4], args[2])
			} else {
				expectStr = replaceRegExpAllAt(bufio.NewReader(bytes.NewReader(buf.Bytes())), args[1], args[2], "")
			}
		}
		assert.Equal(t, len(expectStr), len(testResult), "args: %s", strings.Join(args, " "))
		for offs, i := 0, 1048; true; offs, i = i, i+1048 {
			if i < len(bout) {
				require.Equal(t, expectStr[offs:i], testResult[offs:i], "segments %d:%d", offs, i)
			} else {
				require.Equal(t, expectStr[offs:], testResult[offs:])
				break
			}
		}
	}
}
