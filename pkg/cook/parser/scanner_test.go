package parser

import (
	"path/filepath"
	"testing"

	"github.com/cozees/cook/pkg/cook/ast"
	"github.com/cozees/cook/pkg/cook/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type textSegment struct {
	text   string
	offset int
}

type out struct {
	offs    int
	tok     token.Token
	lit     string
	ci      ast.Expr
	buildCi [][2]textSegment
}

func buildStringBinding(offs int, f *token.File, input [][2]textSegment) ast.Expr {
	builder := ast.NewStringBindingBuilder("")
	for _, isn := range input {
		builder.WriteString(isn[0].text)
		if len(isn[1].text) > 0 {
			builder.AddExpression(ast.NewIdentExpr(isn[1].offset, f, isn[1].text))
		}
	}
	return builder.Build(offs, f)
}

func buildOutput(f *token.File) []*out {
	return []*out{
		{offs: 0, tok: token.INCLUDE, lit: "include"},
		{offs: 8, tok: token.STRING, lit: "Cookfile.debug"},
		{offs: 24, tok: token.LF, lit: "\n"},
		{offs: 26, tok: token.IDENT, lit: "var"},
		{offs: 30, tok: token.ASSIGN},
		{offs: 32, tok: token.INTEGER, lit: "123"},
		{offs: 35, tok: token.LF, lit: "\n"},
		{offs: 36, tok: token.IDENT, lit: "VAR"},
		{offs: 40, tok: token.ASSIGN},
		{offs: 42, tok: token.INTEGER, lit: "123"},
		{offs: 45, tok: token.LF, lit: "\n"},
		{offs: 46, tok: token.IDENT, lit: "VAR"},
		{offs: 50, tok: token.ASSIGN},
		{offs: 52, tok: token.INTEGER, lit: "123"},
		{offs: 56, tok: token.ADD},
		{offs: 64, tok: token.INTEGER, lit: "124"},
		{offs: 68, tok: token.SUB},
		{offs: 76, tok: token.INTEGER, lit: "383"},
		{offs: 79, tok: token.LF, lit: "\n"},
		{offs: 80, tok: token.IDENT, lit: "V"},
		{offs: 82, tok: token.ASSIGN},
		{offs: 84, tok: token.IDENT, lit: "A"},
		{offs: 86, tok: token.QUO},
		{offs: 88, tok: token.IDENT, lit: "B"},
		{offs: 89, tok: token.LF, lit: "\n"},
		{offs: 90, tok: token.IDENT, lit: "V"},
		{offs: 92, tok: token.ASSIGN},
		{offs: 94, tok: token.IDENT, lit: "A"},
		{offs: 95, tok: token.QUO},
		{offs: 96, tok: token.IDENT, lit: "B"},
		{offs: 97, tok: token.LF, lit: "\n"},
		{offs: 98, tok: token.IDENT, lit: "V"},
		{offs: 100, tok: token.ASSIGN},
		{
			offs: 102,
			tok:  token.STRING,
			lit:  "String1 $V ${V} \\v String2 String3",
			buildCi: [][2]textSegment{
				{{text: "String1 ", offset: 102}, textSegment{text: "V", offset: 111}},
				{{text: " ", offset: 113}, textSegment{text: "V", offset: 114}},
				{{text: " \\v String2 String3", offset: 118}, textSegment{text: "", offset: 0}},
			},
		},
		{offs: 138, tok: token.LF, lit: "\n"},
		{offs: 139, tok: token.IDENT, lit: "V"},
		{offs: 141, tok: token.ASSIGN},
		{offs: 143, tok: token.STRING, lit: "String1 ${VAR} String2 String3"},
		{offs: 175, tok: token.LF, lit: "\n"},
		{offs: 176, tok: token.IDENT, lit: "VAR1"},
		{offs: 181, tok: token.ASSIGN},
		{offs: 183, tok: token.FLOAT, lit: "1.32"},
		{offs: 188, tok: token.ADD},
		{offs: 190, tok: token.IDENT, lit: "var"},
		{offs: 193, tok: token.LF, lit: "\n"},
		{offs: 194, tok: token.IDENT, lit: "V1"},
		{offs: 197, tok: token.ASSIGN},
		{offs: 199, tok: token.IDENT, lit: "V"},
		{offs: 200, tok: token.LF, lit: "\n"},
		{offs: 201, tok: token.IDENT, lit: "FILE"},
		{offs: 206, tok: token.ASSIGN},
		{offs: 208, tok: token.STRING, lit: "test/sample/txt"},
		{offs: 225, tok: token.LF, lit: "\n"},
		{offs: 226, tok: token.IDENT, lit: "ARRAY1"},
		{offs: 233, tok: token.ASSIGN},
		{offs: 235, tok: token.LBRACK},
		{offs: 236, tok: token.IDENT, lit: "V1"},
		{offs: 238, tok: token.COMMA},
		{offs: 240, tok: token.IDENT, lit: "V2"},
		{offs: 242, tok: token.COMMA},
		{offs: 244, tok: token.IDENT, lit: "V3"},
		{offs: 246, tok: token.RBRACK},
		{offs: 247, tok: token.LF, lit: "\n"},
		{offs: 248, tok: token.IDENT, lit: "MAP"},
		{offs: 252, tok: token.ASSIGN},
		{offs: 254, tok: token.LBRACE},
		{offs: 255, tok: token.INTEGER, lit: "123"},
		{offs: 258, tok: token.COLON},
		{offs: 260, tok: token.IDENT, lit: "V"},
		{offs: 261, tok: token.COMMA},
		{offs: 263, tok: token.IDENT, lit: "V"},
		{offs: 264, tok: token.COLON},
		{offs: 266, tok: token.IDENT, lit: "var"},
		{offs: 269, tok: token.RBRACE},
		{offs: 270, tok: token.LF, lit: "\n"},
		{offs: 272, tok: token.IDENT, lit: "initialize"},
		{offs: 282, tok: token.COLON},
		{offs: 285, tok: token.IDENT, lit: "initialize"},
		{offs: 295, tok: token.AT},
		{offs: 296, tok: token.IDENT, lit: "window"},
		{offs: 302, tok: token.COLON},
		{offs: 308, tok: token.IDENT, lit: "FILE"},
		{offs: 313, tok: token.ASSIGN},
		{offs: 315, tok: token.STRING, lit: "C:\\\\test\\\\sample\\\\112.txt"},
		{offs: 342, tok: token.LF, lit: "\n"},
		{offs: 344, tok: token.IDENT, lit: "finalize"},
		{offs: 352, tok: token.COLON},
		{offs: 355, tok: token.IDENT, lit: "all"},
		{offs: 358, tok: token.COLON},
		{offs: 360, tok: token.MUL},
		{offs: 363, tok: token.IDENT, lit: "TARGET1"},
		{offs: 370, tok: token.COLON},
		{offs: 372, tok: token.AT},
		{offs: 373, tok: token.IDENT, lit: "workin"},
		{offs: 380, tok: token.STRING, lit: "test/sample"},
		{offs: 391, tok: token.LF, lit: "\n"},
		{offs: 418, tok: token.AT},
		{offs: 419, tok: token.IDENT, lit: "TARGET"},
		{offs: 425, tok: token.LF, lit: "\n"},
		{offs: 430, tok: token.FOR, lit: "for"},
		{offs: 434, tok: token.IDENT, lit: "index"},
		{offs: 440, tok: token.IN, lit: "in"},
		{offs: 443, tok: token.INTEGER, lit: "1"},
		{offs: 444, tok: token.RANGE},
		{offs: 446, tok: token.IDENT, lit: "VAR"},
		{offs: 450, tok: token.LBRACE},
		{offs: 451, tok: token.RBRACE},
		{offs: 452, tok: token.LF, lit: "\n"},
		{offs: 457, tok: token.FOR, lit: "for"},
		{offs: 461, tok: token.IDENT, lit: "index"},
		{offs: 467, tok: token.IN, lit: "in"},
		{offs: 470, tok: token.INTEGER, lit: "1"},
		{offs: 471, tok: token.RANGE},
		{offs: 473, tok: token.INTEGER, lit: "40"},
		{offs: 476, tok: token.LBRACE},
		{offs: 477, tok: token.RBRACE},
		{offs: 478, tok: token.LF, lit: "\n"},
		{offs: 483, tok: token.FOR, lit: "for"},
		{offs: 487, tok: token.IDENT, lit: "index"},
		{offs: 492, tok: token.COMMA},
		{offs: 494, tok: token.IDENT, lit: "val"},
		{offs: 498, tok: token.IN, lit: "in"},
		{offs: 501, tok: token.IDENT, lit: "ARRAY"},
		{offs: 507, tok: token.LBRACE},
		{offs: 508, tok: token.RBRACE},
		{offs: 509, tok: token.LF, lit: "\n"},
		{offs: 514, tok: token.FOR, lit: "for"},
		{offs: 518, tok: token.IDENT, lit: "key"},
		{offs: 521, tok: token.COMMA},
		{offs: 523, tok: token.IDENT, lit: "val"},
		{offs: 527, tok: token.IN, lit: "in"},
		{offs: 530, tok: token.IDENT, lit: "MAP"},
		{offs: 534, tok: token.LBRACE},
		{offs: 535, tok: token.RBRACE},
		{offs: 536, tok: token.LF, lit: "\n"},
		{offs: 538, tok: token.IDENT, lit: "v"},
		{offs: 540, tok: token.ASSIGN},
		{offs: 542, tok: token.INTEGER, lit: "1"},
		{offs: 543, tok: token.LF, lit: "\n"},
		{offs: 548, tok: token.FOR, lit: "for"},
		{offs: 552, tok: token.LBRACE},
		{offs: 556, tok: token.IF, lit: "if"},
		{offs: 559, tok: token.IDENT, lit: "v"},
		{offs: 561, tok: token.GTR},
		{offs: 563, tok: token.INTEGER, lit: "3"},
		{offs: 565, tok: token.LBRACE},
		{offs: 576, tok: token.BREAK, lit: "break"},
		{offs: 581, tok: token.LF, lit: "\n"},
		{offs: 584, tok: token.RBRACE},
		{offs: 586, tok: token.ELSE, lit: "else"},
		{offs: 591, tok: token.IF, lit: "if"},
		{offs: 594, tok: token.IDENT, lit: "v"},
		{offs: 596, tok: token.EQL},
		{offs: 599, tok: token.INTEGER, lit: "2"},
		{offs: 601, tok: token.LBRACE},
		{offs: 612, tok: token.CONTINUE, lit: "continue"},
		{offs: 621, tok: token.AT},
		{offs: 622, tok: token.IDENT, lit: "TARGET"},
		{offs: 628, tok: token.LF, lit: "\n"},
		{offs: 631, tok: token.RBRACE},
		{offs: 632, tok: token.LF, lit: "\n"},
		{offs: 635, tok: token.IDENT, lit: "v"},
		{offs: 636, tok: token.INC},
		{offs: 638, tok: token.LF, lit: "\n"},
		{offs: 643, tok: token.RBRACE},
		{offs: 644, tok: token.LF, lit: "\n"},
		{offs: 649, tok: token.IF, lit: "if"},
		{offs: 652, tok: token.IDENT, lit: "VAR"},
		{offs: 656, tok: token.EQL},
		{offs: 659, tok: token.INTEGER, lit: "123"},
		{offs: 663, tok: token.LBRACE},
		{offs: 673, tok: token.AT},
		{offs: 674, tok: token.IDENT, lit: "TARGET"},
		{offs: 681, tok: token.STRING, lit: "-p"},
		{offs: 683, tok: token.LF, lit: "\n"},
		{offs: 688, tok: token.RBRACE},
		{offs: 690, tok: token.ELSE, lit: "else"},
		{offs: 695, tok: token.IF, lit: "if"},
		{offs: 698, tok: token.IDENT, lit: "VAR"},
		{offs: 702, tok: token.EQL},
		{offs: 705, tok: token.IDENT, lit: "abc"},
		{offs: 709, tok: token.LBRACE},
		{offs: 719, tok: token.AT},
		{offs: 720, tok: token.IDENT, lit: "TARGET1"},
		{offs: 727, tok: token.LF, lit: "\n"},
		{offs: 732, tok: token.RBRACE},
		{offs: 734, tok: token.ELSE, lit: "else"},
		{offs: 739, tok: token.LBRACE},
		{offs: 746, tok: token.RBRACE},
		{offs: 747, tok: token.LF, lit: "\n"},
		{offs: 749, tok: token.IDENT, lit: "TARGET"},
		{offs: 755, tok: token.AT},
		{offs: 756, tok: token.IDENT, lit: "window"},
		{offs: 762, tok: token.COLON},
		{offs: 764, tok: token.IDENT, lit: "TARGET"},
		{offs: 770, tok: token.AT},
		{offs: 771, tok: token.IDENT, lit: "macos"},
		{offs: 776, tok: token.COLON},
		{offs: 778, tok: token.IDENT, lit: "TARGET"},
		{offs: 784, tok: token.AT},
		{offs: 785, tok: token.IDENT, lit: "linux"},
		{offs: 790, tok: token.COLON},
		{offs: 792, tok: token.IDENT, lit: "TARGET"},
		{offs: 798, tok: token.COLON},
		{offs: 863, tok: token.AT},
		{offs: 864, tok: token.IDENT, lit: "GET"},
		{
			offs: 868,
			tok:  token.STRING,
			lit:  "http://www.example.com/$VAR",
			buildCi: [][2]textSegment{
				{{text: "http://www.example.com/", offset: 868}, {text: "VAR", offset: 892}},
			},
		},
		{offs: 896, tok: token.ASSIGN_TO, lit: ">"},
		{
			offs: 898,
			tok:  token.STRING,
			lit:  "$FILE",
			buildCi: [][2]textSegment{
				{{text: "", offset: 898}, {text: "FILE", offset: 898}},
			},
		},
		{offs: 903, tok: token.LF, lit: "\n"},
		{offs: 908, tok: token.HASH},
		{offs: 909, tok: token.IDENT, lit: "ls"},
		{offs: 911, tok: token.LF, lit: "\n"},
		{offs: 916, tok: token.IF, lit: "if"},
		{offs: 919, tok: token.IDENT, lit: "VAR"},
		{offs: 923, tok: token.EQL},
		{offs: 926, tok: token.INTEGER, lit: "234"},
		{offs: 930, tok: token.LBRACE},
		{offs: 945, tok: token.RBRACE},
		{offs: 947, tok: token.ELSE, lit: "else"},
		{offs: 952, tok: token.LBRACE},
		{offs: 959, tok: token.RBRACE},
		{offs: 960, tok: token.LF, lit: "\n"},
	}
}

func TestScanner(t *testing.T) {
	eh := func(offset int, f *token.File, msg string) bool {
		n, l, c := f.Position(offset)
		assert.Fail(t, "syntax error", "%s:%d:%d %s", n, l, c, msg)
		// false terminiate as soon as error occurred otherwise true
		return false
	}

	files := token.NewFiles()
	f, err := files.AddFile(filepath.Join("testdata", "Cookfile"))
	require.NoError(t, err)

	raw, err := f.ReadFile()
	require.NoError(t, err)

	output := buildOutput(f)
	s := NewScanner(f, raw, eh)
	index := 0
	for {
		offs, tok, lit, ci := s.Scan()
		if tok == token.EOF {
			break
		}
		o := output[index]
		pass := assert.Equal(t, o.offs, offs)
		pass = pass && assert.Equal(t, o.tok, tok)
		pass = pass && assert.Equal(t, o.lit, lit)
		if o.buildCi != nil {
			o.ci = buildStringBinding(o.offs, f, o.buildCi)
			pass = pass && assert.EqualValues(t, o.ci, ci)
		} else {
			pass = pass && assert.Nil(t, ci)
		}
		if !pass {
			n, l, c := f.Position(offs)
			require.Fail(t, "syntax error", "%s:%d:%d %s with token %s, %s", n, l, c, "Fail here.", tok, lit)
		}
		index++
	}
}
