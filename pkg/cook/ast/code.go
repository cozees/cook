package ast

import (
	"io"
	"strings"

	"github.com/cozees/cook/pkg/cook/token"
)

func codeOf(c Code) string {
	sb := NewCodeBuilder("", false, 100)
	c.Visit(sb)
	return sb.String()
}

type Code interface {
	String() string
	Visit(cb CodeBuilder)
}

type StringTerminate uint8

const (
	StringTerminateNone  StringTerminate = 0
	StringTerminateBegin                 = 1
	StringTerminateEnd                   = 2
	StringTerminateBoth                  = StringTerminateBegin | StringTerminateEnd
)

type CodeBuilder interface {
	String() string
	io.Writer
	io.ByteWriter
	io.StringWriter
	WriteRune(r rune) (int, error)
	WriteIndent()
	WriteQuoteString(s string, quote byte, stm StringTerminate)
	GetIdent() int
	SetIdent(c int)
	IdentBy(c int)
	Len() int
}

type builder struct {
	*strings.Builder
	indent      string
	indentCount int
	maxLength   int
	formatter   bool
}

func NewCodeBuilder(indent string, formatter bool, maxLength int) CodeBuilder {
	return &builder{indent: indent, formatter: formatter, maxLength: maxLength, Builder: &strings.Builder{}}
}

func (b *builder) GetIdent() int  { return b.indentCount }
func (b *builder) SetIdent(c int) { b.indentCount = c }
func (b *builder) IdentBy(c int)  { b.indentCount += c }

func (b *builder) WriteIndent() {
	for i := 0; i < b.indentCount; i++ {
		b.WriteString(b.indent)
	}
}

func (b *builder) WriteQuoteString(s string, quote byte, stm StringTerminate) {
	if stm&StringTerminateBegin == StringTerminateBegin {
		b.WriteByte(quote)
	}
	for _, r := range s {
		if r == rune(quote) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	if stm&StringTerminateEnd == StringTerminateEnd {
		b.WriteByte(quote)
	}
}

func (bl *BasicLit) String() string            { return codeOf(bl) }
func (id *Ident) String() string               { return id.Name }
func (cd *Conditional) String() string         { return codeOf(cd) }
func (fb *Fallback) String() string            { return codeOf(fb) }
func (sf *SizeOf) String() string              { return codeOf(sf) }
func (it *IsType) String() string              { return codeOf(it) }
func (tc *TypeCast) String() string            { return codeOf(tc) }
func (e *Exit) String() string                 { return codeOf(e) }
func (al *ArrayLiteral) String() string        { return codeOf(al) }
func (ml *MapLiteral) String() string          { return codeOf(ml) }
func (mm *MergeMap) String() string            { return codeOf(mm) }
func (d *Delete) String() string               { return codeOf(d) }
func (ix *Index) String() string               { return codeOf(ix) }
func (sv *SubValue) String() string            { return codeOf(sv) }
func (r *Interval) String() string             { return codeOf(r) }
func (c *Call) String() string                 { return codeOf(c) }
func (pp *Pipe) String() string                { return codeOf(pp) }
func (rf *ReadFrom) String() string            { return codeOf(rf) }
func (rt *RedirectTo) String() string          { return codeOf(rt) }
func (p *Paren) String() string                { return codeOf(p) }
func (un *Unary) String() string               { return codeOf(un) }
func (idc *IncDec) String() string             { return codeOf(idc) }
func (b *Binary) String() string               { return codeOf(b) }
func (t *Transformation) String() string       { return codeOf(t) }
func (si *StringInterpolation) String() string { return codeOf(si) }

func (bl *BasicLit) Visit(cb CodeBuilder) {
	if bl.Mark != 0 {
		if bl.Kind != token.STRING {
			panic("quote mark use on a non string token")
		}
		cb.WriteQuoteString(bl.Lit, bl.Mark, StringTerminateBoth)
	} else {
		cb.WriteString(bl.Lit)
	}
}

func (id *Ident) Visit(cb CodeBuilder) { cb.WriteString(id.Name) }

func (cd *Conditional) Visit(cb CodeBuilder) {
	cd.Cond.Visit(cb)
	cb.WriteString(" ? ")
	cd.True.Visit(cb)
	cb.WriteString(" : ")
	cd.False.Visit(cb)
}

func (fb *Fallback) Visit(cb CodeBuilder) {
	fb.Primary.Visit(cb)
	cb.WriteString(" ?? ")
	fb.Default.Visit(cb)
}

func (sf *SizeOf) Visit(cb CodeBuilder) {
	cb.WriteString("sizeof ")
	sf.X.Visit(cb)
}

func (it *IsType) Visit(cb CodeBuilder) {
	it.X.Visit(cb)
	cb.WriteString(" is ")
	for i, t := range it.Types {
		if i > 0 {
			cb.WriteString(" | ")
		}
		cb.WriteString(t.String())
	}
}

func (tc *TypeCast) Visit(cb CodeBuilder) {
	cb.WriteString(tc.To.String())
	cb.WriteByte('(')
	tc.X.Visit(cb)
	cb.WriteByte(')')
}

func (e *Exit) Visit(cb CodeBuilder) {
	cb.WriteString("exit ")
	e.ExitCode.Visit(cb)
}

func (al *ArrayLiteral) Visit(cb CodeBuilder) {
	cb.WriteByte('[')
	if al.Multiline {
		cb.WriteByte('\n')
		cb.IdentBy(1)
		for _, val := range al.Values {
			cb.WriteIndent()
			val.Visit(cb)
			cb.WriteString(",\n")
		}
		cb.IdentBy(-1)
		cb.WriteIndent()
		cb.WriteString("]\n")
	} else {
		for i, val := range al.Values {
			if i > 0 {
				cb.WriteString(", ")
			}
			val.Visit(cb)
		}
		cb.WriteString("]")
	}
}

func (ml *MapLiteral) Visit(cb CodeBuilder) {
	cb.WriteByte('{')
	if ml.Multiline {
		cb.WriteByte('\n')
		cb.IdentBy(1)
		for i, key := range ml.Keys {
			cb.WriteIndent()
			key.Visit(cb)
			cb.WriteString(": ")
			ml.Values[i].Visit(cb)
			cb.WriteString(",\n")
		}
		cb.IdentBy(-1)
		cb.WriteIndent()
	} else {
		for i, key := range ml.Keys {
			if i > 0 {
				cb.WriteString(", ")
			}
			key.Visit(cb)
			cb.WriteString(": ")
			ml.Values[i].Visit(cb)
		}
	}
	cb.WriteByte('}')
}

func (mm *MergeMap) Visit(cb CodeBuilder) {
	if mm.Op != token.ILLEGAL {
		cb.WriteString(mm.Op.String())
	}
	mm.Value.Visit(cb)
}

func (d *Delete) Visit(cb CodeBuilder) {
	cb.WriteString("delete ")
	d.X.Visit(cb)
	cb.WriteByte('{')
	if d.End != nil {
		d.Indexes[0].Visit(cb)
		cb.WriteString("..")
		d.End.Visit(cb)
	} else {
		for i, ni := range d.Indexes {
			if i > 0 {
				cb.WriteString(", ")
			}
			ni.Visit(cb)
		}
	}
	cb.WriteByte('}')
}

func (ix *Index) Visit(cb CodeBuilder) {
	ix.X.Visit(cb)
	cb.WriteByte('[')
	ix.Index.Visit(cb)
	cb.WriteByte(']')
}

func (sv *SubValue) Visit(cb CodeBuilder) {
	sv.X.Visit(cb)
	sv.Range.Visit(cb)
}

func (r *Interval) Visit(cb CodeBuilder) {
	if r.AInclude {
		cb.WriteByte('[')
	} else {
		cb.WriteByte('(')
	}
	r.A.Visit(cb)
	cb.WriteString("..")
	r.B.Visit(cb)
	if r.BInclude {
		cb.WriteByte(']')
	} else {
		cb.WriteByte(')')
	}
}

func (c *Call) Visit(cb CodeBuilder) {
	cb.WriteString(c.Kind.String())
	cb.WriteString(c.Name)
	for _, arg := range c.Args {
		cb.WriteByte(' ')
		if arg == nil {
			cb.WriteString("\\\n")
			cb.WriteIndent()
		} else {
			arg.Visit(cb)
		}
	}
}

func (pp *Pipe) Visit(cb CodeBuilder) {
	pp.X.Visit(cb)
	if pp.Y != nil {
		cb.WriteString(" | ")
		pp.Y.Visit(cb)
	}
}

func (rf *ReadFrom) Visit(cb CodeBuilder) {
	cb.WriteString("< ")
	rf.File.Visit(cb)
}

func (rt *RedirectTo) Visit(cb CodeBuilder) {
	rt.Caller.Visit(cb)
	cb.WriteString(" >")
	if rt.Append {
		cb.WriteByte('>')
	}
	for _, f := range rt.Files {
		cb.WriteByte(' ')
		if f == nil {
			cb.WriteString("\\\n")
			cb.WriteIndent()
		}
		f.Visit(cb)
	}
}

func (p *Paren) Visit(cb CodeBuilder) {
	cb.WriteByte('(')
	p.Inner.Visit(cb)
	cb.WriteByte(')')
}

func (un *Unary) Visit(cb CodeBuilder) {
	cb.WriteString(un.Op.String())
	un.X.Visit(cb)
}

func (idc *IncDec) Visit(cb CodeBuilder) {
	idc.X.Visit(cb)
	cb.WriteString(idc.Op.String())
}

func (b *Binary) Visit(cb CodeBuilder) {
	if _, ok := b.L.(*Binary); ok {
		cb.WriteByte('(')
		b.L.Visit(cb)
		cb.WriteByte(')')
	} else {
		b.L.Visit(cb)
	}
	cb.WriteByte(' ')
	cb.WriteString(b.Op.String())
	cb.WriteByte(' ')
	if _, ok := b.R.(*Binary); ok {
		cb.WriteByte('(')
		b.R.Visit(cb)
		cb.WriteByte(')')
	} else {
		b.R.Visit(cb)
	}
}

func (t *Transformation) Visit(cb CodeBuilder) {
	t.Ident.Visit(cb)
	t.Fn.Visit(cb)
}

func (si *StringInterpolation) Visit(cb CodeBuilder) {
	cb.WriteByte(si.mark)
	offs := 0
	for i := 0; i < len(si.pos); i++ {
		if offs < si.pos[i] {
			cb.WriteQuoteString(si.raw[offs:si.pos[i]], si.mark, StringTerminateNone)
			offs = si.pos[i]
		}
		cb.WriteString("${")
		si.nodes[i].Visit(cb)
		cb.WriteByte('}')
	}
	if offs < len(si.raw) {
		cb.WriteQuoteString(si.raw[offs:], si.mark, StringTerminateNone)
	}
	cb.WriteByte(si.mark)
}

// =============== Statement Block ===============

func (ifst *IfStatement) String() string           { return codeOf(ifst) }
func (efst *ElseStatement) String() string         { return codeOf(efst) }
func (fst *ForStatement) String() string           { return codeOf(fst) }
func (bcs *BreakContinueStatement) String() string { return codeOf(bcs) }
func (as *AssignStatement) String() string         { return codeOf(as) }
func (bs *BlockStatement) String() string          { return codeOf(bs) }
func (ews *ExprWrapperStatement) String() string   { return codeOf(ews) }
func (rs *ReturnStatement) String() string         { return codeOf(rs) }

func (fst *ForStatement) Visit(cb CodeBuilder) {
	cb.WriteString("for")
	if fst.Label != "" {
		cb.WriteByte(':')
		cb.WriteString(fst.Label)
	}
	if fst.I != nil {
		cb.WriteByte(' ')
		cb.WriteString(fst.I.Name)
		if fst.Value != nil {
			cb.WriteString(", ")
			cb.WriteString(fst.Value.Name)
		}
		cb.WriteString(" in ")
		if fst.Oprnd != nil {
			fst.Oprnd.Visit(cb)
		} else {
			fst.Range.Visit(cb)
		}
	}
	fst.Insts.Visit(cb)
}

func (ifst *IfStatement) Visit(cb CodeBuilder) {
	cb.WriteString("if ")
	ifst.Cond.Visit(cb)
	ifst.Insts.Visit(cb)
	if ifst.Else != nil {
		ifst.Else.Visit(cb)
	}
}

func (efst *ElseStatement) Visit(cb CodeBuilder) {
	cb.WriteString(" else")
	if efst.IfStmt != nil {
		cb.WriteByte(' ')
		efst.IfStmt.Visit(cb)
	} else {
		efst.Insts.Visit(cb)
	}
}

func (bs *BlockStatement) Visit(cb CodeBuilder) {
	if !bs.plain {
		cb.WriteString(" {\n")
	}
	if !bs.root {
		cb.IdentBy(1)
	}
	for _, stmt := range bs.Stmts {
		cb.WriteIndent()
		stmt.Visit(cb)
		cb.WriteByte('\n')
	}
	if !bs.root {
		cb.IdentBy(-1)
	}
	if !bs.plain {
		cb.WriteIndent()
		cb.WriteString("}")
	}
}

func (as *AssignStatement) Visit(cb CodeBuilder) {
	as.Ident.Visit(cb)
	cb.WriteByte(' ')
	cb.WriteString(as.Op.String())
	cb.WriteByte(' ')
	as.Value.Visit(cb)
}

func (bcs *BreakContinueStatement) Visit(cb CodeBuilder) {
	cb.WriteString(bcs.Op.String())
	if bcs.Label != "" {
		cb.WriteString(":")
		cb.WriteString(bcs.Label)
	}
}

func (ews *ExprWrapperStatement) Visit(cb CodeBuilder) { ews.X.Visit(cb) }

func (rs *ReturnStatement) Visit(cb CodeBuilder) {
	cb.WriteString("return ")
	rs.X.Visit(cb)
}

// =============== Cook, Target and Function Block ===============

func (c *cook) String() string      { return codeOf(c) }
func (t *Target) String() string    { return codeOf(t) }
func (fn *Function) String() string { return codeOf(fn) }

func (c *cook) Visit(cb CodeBuilder) {
	c.Insts.Visit(cb)
	// TODO: how to handle multiple file ??
	for _, t := range c.initializeTargets {
		t.Visit(cb)
	}

	for _, t := range c.finalizeTargets {
		t.Visit(cb)
	}

	if c.targetAll != nil {
		c.targetAll.Visit(cb)
	}

	for _, t := range c.targetIndexes {
		t.Visit(cb)
	}

	for _, fn := range c.fns {
		fn.Visit(cb)
	}
}

func (t *Target) Visit(cb CodeBuilder) {
	if cb.Len() > 0 {
		cb.WriteByte('\n')
	}
	cb.WriteString(t.name)
	if t.all {
		cb.WriteString(": *\n")
	} else {
		cb.WriteString(":\n")
		t.Insts.plain = true
		t.Insts.Visit(cb)
	}
}

func (fn *Function) Visit(cb CodeBuilder) {
	if fn.Name != "" {
		cb.WriteString(fn.Name)
	}
	cb.WriteByte('(')
	for i, arg := range fn.Args {
		if i > 0 {
			cb.WriteString(", ")
		}
		arg.Visit(cb)
	}
	cb.WriteByte(')')
	if fn.Lambda == token.LAMBDA {
		cb.WriteString(" => ")
		fn.X.Visit(cb)
	} else {
		fn.Insts.Visit(cb)
	}
}
