package parser

import (
	"testing"

	"github.com/cozees/cook/pkg/cook/ast"
	"github.com/cozees/cook/pkg/cook/token"
	"github.com/stretchr/testify/assert"
)

type inout struct {
	in string
	so string
}

var binaryExprTestCase = []*inout{
	{in: "2 + 6 * 2", so: "(2+(6*2))"},
	{in: "5 * 3 + 1", so: "((5*3)+1)"},
	{in: "V + 2 / 5", so: "(V+(2/5))"},
	{in: "V * A + 1 / 2", so: "((V*A)+(1/2))"},
	{in: "+V - 2 + 3 / 4 * 5", so: "((+V-2)+((3/4)*5))"},
	{in: "V > 1 + 3 / 5", so: "(V>(1+(3/5)))"},
	{in: "1 + 2 * V > 9 + A / 5", so: "((1+(2*V))>(9+(A/5)))"},
	{in: "1 << 2 > 2 * 1 * 1 - A", so: "((1<<2)>(((2*1)*1)-A))"},
	{in: "'abc text' + 1 * 2", so: "(\"abc text\"+(1*2))"},
	{in: "!V == false", so: "(!V==false)"},
	{in: "[1, 2, V, 2.3, 'text'] + 1", so: "([1, 2, V, 2.3, \"text\"]+1)"},
	// parsed but failed to evaluate
	{in: "!V != 2 == 2", so: "((!V!=2)==2)"},
	{in: "V > 2 > 3 + 1", so: "((V>2)>(3+1))"}, // boolean compare with integer failed
	{in: "V * A > 1 + V % 2 < 2 * 6", so: "(((V*A)>(1+(V%2)))<(2*6))"},
}

func TestBinaryExprParser(t *testing.T) {
	p := NewParser().(*implParser)
	files := token.NewFiles()
	for _, vio := range binaryExprTestCase {
		src := []byte(vio.in)
		f, err := files.AddFileInMemory("test/Cookfile", src, len(src))
		assert.NoError(t, err)
		p.s = NewScanner(f, src, p.errorHandler)
		p.init()
		expr := p.parseBinaryExpr(token.LowestPrec + 1)
		assert.Equal(t, vio.so, expr.String())
	}
}

var operandTestCase = []*inout{
	{in: "123", so: "123"},
	{in: "V", so: "V"},
	{in: "12.34", so: "12.34"},
	{in: "'text with single\\' quote'", so: "\"text with single\\' quote\""},
	{in: "false", so: "false"},
	{in: "true", so: "true"},
	{in: "[1, 2, 2.3, V]", so: "[1, 2, 2.3, V]"},
	{in: "[1, {1:2, 3.2:1}, V]", so: "[1, {1: 2, 3.2: 1}, V]"},
	{in: "{1:2, V:V1, 3.32:'text', 'abc':V2}", so: "{1: 2, V: V1, 3.32: \"text\", \"abc\": V2}"},
	{in: "V[1]", so: "V[1]"},
	{in: "V[V]", so: "V[V]"},
	{in: "V[V + 1]", so: "V[(V+1)]"},
}

func TestOperandParser(t *testing.T) {
	p := NewParser().(*implParser)
	files := token.NewFiles()
	for _, vio := range operandTestCase {
		src := []byte(vio.in)
		f, err := files.AddFileInMemory("test/Cookfile", src, len(src))
		assert.NoError(t, err)
		p.s = NewScanner(f, src, p.errorHandler)
		p.init()
		p.next()
		expr := p.parseOperand()
		assert.Equal(t, vio.so, expr.String())
	}
}

var stmtTestCase = []*inout{
	{in: "exit 1", so: "\nexit 1\n"},
	{in: "exit 201", so: "\nexit 201\n"},
	{in: "V = integer(A)", so: "\nV = integer(A)\n"},
	{in: "V = float(\"124.2\")", so: "\nV = float(\"124.2\")\n"},
	{in: "V = A is integer", so: "\nV = A is integer\n"},
	{in: "V = A is integer | float | string", so: "\nV = A is integer | float | string\n"},
	{in: "V = A ?? B", so: "\nV = A ?? B\n"},
	{in: "V = A + 1 * C ?? 1 + 2 * B", so: "\nV = (A+(1*C)) ?? (1+(2*B))\n"},
	{in: "V = A ?? B ?? C ?? D", so: "\nV = A ?? B ?? C ?? D\n"},
	{in: "V = A ?? false", so: "\nV = A ?? false\n"},
	{in: "C = A * 1.2 ?? false", so: "\nC = (A*1.2) ?? false\n"},
	{in: "V = A ? 123 : B", so: "\nV = A ? 123 : B\n"},
	{in: "V = A ? 123 : B ? 321 : C", so: "\nV = A ? 123 : B ? 321 : C\n"},
	{in: "V = A ? 123 : B + 123 * 2", so: "\nV = A ? 123 : (B+(123*2))\n"},
	{in: "V = A ? 123 + C * 2 : B + 123 * 2", so: "\nV = A ? (123+(C*2)) : (B+(123*2))\n"},
	{in: "V = (A ? 123 : B) + 1", so: "\nV = ((A ? 123 : B)+1)\n"},
	{in: "VAR = \"Sample ${A[1]} text\"", so: "\nVAR = \"Sample ${A[1]} text\"\n"},
	{in: "VAR = sizeof A", so: "\nVAR = sizeof A\n"},
	{in: "VAR = 123", so: "\nVAR = 123\n"},
	{in: "VAR = '12a53'", so: "\nVAR = \"12a53\"\n"},
	{in: "VAR = {1:123, V:VAR, V1:'12a'}", so: "\nVAR = {1: 123, V: VAR, V1: \"12a\"}\n"},
	{in: "VAR = [1, 2, 'a']", so: "\nVAR = [1, 2, \"a\"]\n"},
	{in: "VAR = [1, 2, \n'a']", so: "\nVAR = [1, 2, \"a\"]\n"},
	{in: "VAR = [[1, 2, 'a'],\n[1, 'b'],\n]", so: "\nVAR = [[1, 2, \"a\"], [1, \"b\"]]\n"},
	{
		in: `
VAR = 123
target:
    @GET -h $HEADER -k $V[1] www.example.com > $VAR test/sample.txt
	@ECHO $VAR >> sample.txt
`,
		so: `
VAR = 123
target:
    @GET "-h" HEADER "-k" V[1] "www.example.com" > $VAR "test/sample.txt"
    @ECHO VAR >> "sample.txt"
`,
	},
	{
		in: `
target:
    VAR = @GET www.example.com
	VAR = #ls
	for i in 1..100 {
		X = 123
		if K {
			@FUNC "arg1" "arg2"
		}
	}
`,
		so: `
target:
    VAR = @GET "www.example.com"
    VAR = #ls
    for i in 1..100 {
        X = 123
        if K {
            @FUNC "arg1" "arg2"
        }
    }
`,
	},
}

func TestStatementParser(t *testing.T) {
	p := NewParser().(*implParser)
	files := token.NewFiles()
	for i, vio := range stmtTestCase {
		t.Logf("TestStatementParser case #%d", i+1)
		src := []byte(vio.in)
		f, err := files.AddFileInMemory("test/Cookfile", src, len(src))
		assert.NoError(t, err)
		p.s = NewScanner(f, src, p.errorHandler)
		cp := ast.NewCookProgram()
		p.parseProgram(f, cp)
		assert.Equal(t, vio.so[1:], cp.String(0))
	}
}
