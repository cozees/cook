package token

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

type Files map[string]*File

func NewFiles() Files { return make(map[string]*File) }

func (fls Files) IsExisted(name string) bool {
	absPath, err := filepath.Abs(name)
	if err != nil {
		return false
	}
	_, ok := fls[absPath]
	return ok
}

func (fls Files) AddFile(name string) (*File, error) {
	absPath, err := filepath.Abs(name)
	if err != nil {
		return nil, err
	}
	fts, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	if fts.IsDir() {
		return nil, fmt.Errorf("%s is not a file", name)
	}
	return fls.AddFileInMemory(absPath, nil, int(fts.Size()))
}

func (fls Files) AddFileInMemory(path string, src []byte, size int) (file *File, err error) {
	if !filepath.IsAbs(path) {
		if path, err = filepath.Abs(path); err != nil {
			return nil, err
		}
	}
	file = &File{
		name:  filepath.Base(path),
		dir:   filepath.Dir(path),
		size:  size,
		lines: make([]int, 0),
		src:   src,
	}
	fls[path] = file
	return file, nil
}

type File struct {
	name  string
	dir   string
	size  int
	lines []int
	src   []byte
}

func (f *File) Abs() string { return filepath.Join(f.dir, f.name) }

func (f *File) Name() string { return f.name }

func (f *File) Dir() string { return f.dir }

func (f *File) Size() int { return f.size }

func (f *File) ReadFile() ([]byte, error) {
	if len(f.src) > 0 {
		return f.src, nil
	}
	return ioutil.ReadFile(filepath.Join(f.dir, f.name))
}

func (f *File) AddLine(offset int) { f.lines = append(f.lines, offset) }

func (f *File) Position(offset int) (filename string, line, column int) {
	filename = f.name
	if i := sort.Search(len(f.lines), func(i int) bool { return f.lines[i] > offset }) - 1; i >= 0 {
		line, column = i+1, offset-f.lines[i]+1
	}
	return
}
