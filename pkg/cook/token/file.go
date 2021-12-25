package token

import (
	"fmt"
	"sort"
	"sync"
)

type File struct {
	name string
	size int

	mutex sync.Mutex
	lines []int
}

func NewFile(name string, size int) *File { return &File{name: name, size: size} }

func (f *File) Name() string { return f.name }

func (f *File) LineCount() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	n := len(f.lines)
	return n
}

func (f *File) AddLine(offset int) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if i := len(f.lines); (i == 0 || f.lines[i-1] < offset) && offset < f.size {
		f.lines = append(f.lines, offset)
	}
}

func (f *File) ValidateOffset(offset int) int {
	if offset > f.size {
		panic(fmt.Sprintf("invalid file offset %d (should be <= %d)", offset, f.size))
	}
	return offset
}

func (f *File) Position(offset int) (pos Position) {
	pos.Offset = offset
	pos.Filename, pos.Line, pos.Column = f.unpack(offset)
	return
}

func (f *File) unpack(offset int) (filename string, line, column int) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	filename = f.name
	i := sort.Search(len(f.lines), func(i int) bool { return f.lines[i] > offset }) - 1
	if i >= 0 {
		line, column = i+1, offset-f.lines[i]+1
	} else {
		line, column = 1, offset
	}
	return
}

type Position struct {
	Filename string
	Offset   int
	Line     int
	Column   int
}

func (pos Position) String() string {
	s := pos.Filename
	if s != "" {
		s += ":"
	}
	s += fmt.Sprintf("%d", pos.Line)
	if pos.Column != 0 {
		s += fmt.Sprintf(":%d", pos.Column)
	}
	return s
}
