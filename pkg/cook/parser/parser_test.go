package parser

import (
	"testing"

	"github.com/cozees/cook/pkg/cook/ast"
	"github.com/cozees/cook/pkg/cook/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestParser(t *testing.T, src string) *parser {
	f := token.NewFile("sample", len(src))
	p := NewParser().(*parser)
	require.NoError(t, p.init(f, []byte(src)))
	return p
}

type parserInputCase struct {
	in  string
	out string
}

var simpleCases = []*parserInputCase{
	/* case 1 */ {in: "VAR = 123", out: "VAR = 123\n"},
	/* case 2 */ {in: "VAR = 123 * 383", out: "VAR = 123 * 383\n"},
	/* case 3 */ {in: "VAR = 123 > 383 > a", out: "VAR = (123 > 383) && (383 > a)\n"},
	/* case 4 */ {in: "VAR = 'text'", out: "VAR = 'text'\n"},
	/* case 5 */ {in: "VAR[1] = 32", out: "VAR[1] = 32\n"},
	/* case 6 */ {in: "VAR['key'] = 34 * 73 / a2", out: "VAR['key'] = (34 * 73) / a2\n"},
	/* case 7 */ {in: "VAR = 34 * a1[1] / a2", out: "VAR = (34 * a1[1]) / a2\n"},
	/* case 8 */ {in: "if a == 1 { A += 2 }", out: "if a == 1 {\nA += 2\n}\n"},
	/* case 9 */ {in: "if a == 1 { A += 2 } else { A = 3 }", out: "if a == 1 {\nA += 2\n} else {\nA = 3\n}\n"},
	/* case 10 */ {in: "if a == 1 { A += 2 } else if cc { A = 3 }", out: "if a == 1 {\nA += 2\n} else if cc {\nA = 3\n}\n"},
	/* case 11 */ {in: "for i in [1..4] { A += i }", out: "for i in [1..4] {\nA += i\n}\n"},
	/* case 12 */ {in: "for i in ]1..4] { A += i }", out: "for i in (1..4] {\nA += i\n}\n"},
	/* case 13 */ {in: "for i in (1..4[ { A += i }", out: "for i in (1..4) {\nA += i\n}\n"},
	/* case 14 */ {in: "for i in (1..4) { A += i }", out: "for i in (1..4) {\nA += i\n}\n"},
	/* case 15 */ {in: "for i,v in [1,2,3] { A += i }", out: "for i, v in [1, 2, 3] {\nA += i\n}\n"},
	/* case 16 */ {in: "for key, val in {a:b, 1:'text'} { A += i }", out: "for key, val in {a: b, 1: 'text'} {\nA += i\n}\n"},
	/* case 17 */ {in: "for i in [1..4] { if a > 2 { break } else { continue } }", out: "for i in [1..4] {\nif a > 2 {\nbreak\n} else {\ncontinue\n}\n}\n"},
	/* case 18 */ {in: "for i in [1..4] { if a > 2 { break:label }\n }", out: "for i in [1..4] {\nif a > 2 {\nbreak:label\n}\n}\n"},
	/* case 19 */ {in: "for:label i in [1..4] { A += i }", out: "for:label i in [1..4] {\nA += i\n}\n"},
	/* case 20 */ {in: "for { A += i }", out: "for {\nA += i\n}\n"},
	/* case 21 */ {in: "for:label { A += i }", out: "for:label {\nA += i\n}\n"},
	/* case 22 */ {in: "target:", out: "target:\n"},
	/* case 23 */ {in: "exit a & 1 & u", out: "exit (a & 1) & u\n"},
	/* case 24 */ {in: "exit A", out: "exit A\n"},
	/* case 25 */ {in: "exit A[1]", out: "exit A[1]\n"},
	/* case 26 */ {in: "@cmd test 'text' 123", out: "@cmd test 'text' 123\n"},
	/* case 27 */ {in: "@cmd test 'text' \\\n 123 sample \\\n 4.2 '932'", out: "@cmd test 'text' 123 sample 4.2 '932'\n"},
	/* case 28 */ {in: "@cmd test 'text' \\\n 123 sample \\\n 4.2 '932'\nA += 2", out: "@cmd test 'text' 123 sample 4.2 '932'\nA += 2\n"},
	/* case 29 */ {in: "if a { @cmd a b c 1\n }", out: "if a {\n@cmd a b c 1\n}\n"},
	/* case 30 */ {in: "A = B(i, v) => i + 1 * v", out: "A = B(i, v) => i + (1 * v)\n"},
	/* case 31 */ {in: "A = 'string interpolation $a text also allow an ${89 + 3} expression'", out: "A = 'string interpolation ${a} text also allow an ${89 + 3} expression'\n"},
	/* case 32 */ {in: "A = '$a is a way to format \\${a * 2} value ${a * 2} output'", out: "A = '${a} is a way to format \\${a * 2} value ${a * 2} output'\n"},
	/* case 33 */ {in: "@print '-e' '-n' a >> file", out: "@print '-e' '-n' a >> file\n"},
	/* case 34 */ {in: "A = @print '-e' '-n' a >> file", out: "A = @print '-e' '-n' a >> file\n"},
	/* case 35 */ {in: "if a is integer | float {}", out: "if a is integer | float {\n}\n"},
	/* case 36 */ {in: "a++", out: "a++\n"},
	/* case 37 */ {in: "if b { a++\n }", out: "if b {\na++\n}\n"},
	/* case 38 */ {in: "A = B ?? 12", out: "A = B ?? 12\n"},
	/* case 39 */ {in: "A = B ? 12 : C", out: "A = B ? 12 : C\n"},
	/* case 40 */ {in: "if a { A[i] = 2 * 3 }", out: "if a {\nA[i] = 2 * 3\n}\n"},
	/* case 41 */ {
		in:  "#pub 'run' file '-o' \"${COVERAGE}/${base}.lcov.info\" \\\n \"--packages=.packages\" \"--report-on=lib\"",
		out: "#pub 'run' file '-o' \"${COVERAGE}/${base}.lcov.info\" \"--packages=.packages\" \"--report-on=lib\"\n",
	},
	/* case 42 */ {in: "A = ['*.go']", out: "A = ['parser.go', 'parser_test.go', 'scanner.go', 'scanner_test.go']\n"},
	/* case 43 */ {in: "A = [123, '*.go']", out: "A = [123, 'parser.go', 'parser_test.go', 'scanner.go', 'scanner_test.go']\n"},
	/* case 44 */ {in: "if sizeof ~'file' == -1 { @print 123 \n }", out: "if sizeof ~'file' == -1 {\n@print 123\n}\n"},
	/* case 45 */ {in: "@print 45 | @print 'text'", out: "@print 45 | @print 'text'\n"},
	/* case 46 */ {in: "@print 45 | @print 'text' >> FILE", out: "@print 45 | @print 'text' >> FILE\n"},
	/* case 47 */ {in: "if on linux {}", out: "if on linux {\n}\n"},
	/* case 48 */ {in: "if on darwin {}", out: "if on darwin {\n}\n"},
	/* case 49 */ {in: "if on windows {}", out: "if on windows {\n}\n"},
	/* case 50 */ {in: "if ~'file' exists {}", out: "if ~'file' exists {\n}\n"},
	/* case 51 */ {in: "if A exists {}", out: "if A exists {\n}\n"},
	/* case 52 */ {in: "if A[1] exists {}", out: "if A[1] exists {\n}\n"},
	/* case 53 */ {in: "if @print exists {}", out: "if @print exists {\n}\n"},
	/* case 54 */ {in: "if #rmdir exists {}", out: "if #rmdir exists {\n}\n"},
	/* case 55 */ {in: "if #rmdir exists && on windows {}", out: "if #rmdir exists && on windows {\n}\n"},
}

func TestParseSimpleStatement(t *testing.T) {
	for i, tc := range simpleCases {
		t.Logf("TestParseSimple case #%d", i+1)
		p := NewParser()
		c, err := p.ParseSrc(token.NewFile("sample", len(tc.in)), []byte(tc.in))
		if tc.out == "" {
			assert.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.Equal(t, tc.out, c.String())
		}
	}
}

var unaryTestCases = []*parserInputCase{
	/* case 1 */ {in: "!a", out: "!a"},
	/* case 2 */ {in: "+a", out: "+a"},
	/* case 3 */ {in: "-a", out: "-a"},
	/* case 4 */ {in: "^a", out: "^a"},
	/* case 5 */ {in: "$1", out: "1"},
	/* case 6 */ {in: "$a", out: ""},
	/* case 7 */ {in: "sizeof a", out: "sizeof a"},
	/* case 8 */ {in: "integer(1.2)", out: "integer(1.2)"},
	/* case 9 */ {in: "float(12)", out: "float(12)"},
	/* case 10 */ {in: "boolean('true')", out: "boolean('true')"},
	/* case 11 */ {in: "string(12)", out: "string(12)"},
	/* case 12 */ {in: "~a", out: "~a"},
	/* case 12 */ {in: "~'file/text.txt'", out: "~'file/text.txt'"},
	/* case 12 */ {in: "sizeof ~'file/text.txt'", out: "sizeof ~'file/text.txt'"},
}

func TestParseUnary(t *testing.T) {
	for i, tc := range unaryTestCases {
		t.Logf("TestParseUnary case #%d", i+1)
		p := createTestParser(t, tc.in)
		p.next()
		x := p.parseUnaryExpr()
		if tc.out == "" {
			assert.Nil(t, x)
		} else {
			require.NotNil(t, x)
			assert.Equal(t, tc.out, x.String())
		}
	}
}

var binaryTestCases = []*parserInputCase{
	/* case 1 */ {in: "a * b + 1", out: "(a * b) + 1"},
	/* case 2 */ {in: "a + b * 1", out: "a + (b * 1)"},
	/* case 3 */ {in: "a / b * 1", out: "(a / b) * 1"},
	/* case 4 */ {in: "(a / b) * 1", out: "(a / b) * 1"},
	/* case 5 */ {in: "(a / b) > 1", out: "(a / b) > 1"},
	/* case 6 */ {in: "a && b && c", out: "(a && b) && c"},
	/* case 7 */ {in: "1 | 2 > 2 * a", out: "(1 | 2) > (2 * a)"},
	/* case 8 */ {in: "2 & 3 | 3 > 3 * h", out: "((2 & 3) | 3) > (3 * h)"},
	/* case 9 */ {in: "a & 2 && 3 | b > 98", out: "(a & 2) && ((3 | b) > 98)"},
	/* case 10 */ {in: "a && b > c", out: "a && (b > c)"},
	/* case 11 */ {in: "12 > b > c > 3", out: "((12 > b) && (b > c)) && (c > 3)"},
	/* case 12 */ {in: "12 > b > c + 3", out: "(12 > b) && (b > (c + 3))"},
	/* case 13 */ {in: "1 + a > b + 4 != c + 3", out: "((1 + a) > (b + 4)) && ((b + 4) != (c + 3))"},
	/* case 14 */ {in: "a | b << 2 * u & 1", out: "a | (((b << 2) * u) & 1)"},
}

func TestParseBinary(t *testing.T) {
	for i, tc := range binaryTestCases {
		t.Logf("TestParseBinary case #%d", i+1)
		p := createTestParser(t, tc.in)
		x := p.parseBinaryExpr(false, token.LowestPrec+1)
		require.NotNil(t, x)
		assert.Equal(t, tc.out, x.String())
	}
}

const src = `
A = 12
B = A * 2

initialize:
   A += 8.2

finalize:
   @print 123

all: *

target:
   if A < 20 {
      for i in [A..39] {
         @print i B
      }
   } else {
      @print "nothing to execute"
   }
`

func TestParseCode(t *testing.T) {
	p := NewParser()
	c, err := p.ParseSrc(token.NewFile("sample", len(src)), []byte(src))
	require.NoError(t, err)
	cb := ast.NewCodeBuilder("   ", true, 120)
	c.Visit(cb)
	assert.Equal(t, src[1:], cb.String())
}
