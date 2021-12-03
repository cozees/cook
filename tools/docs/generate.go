package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cozees/cook/pkg/runtime/args"
	"github.com/cozees/cook/pkg/runtime/function"
)

var docPaths = filepath.Join("docs", "functions")

type functionGroup struct {
	Name        string
	File        string
	Description string
	Flags       func() []*args.Flags
}

const (
	stringDesc = `String functions provide several pre-define functionality that can be used to manipulate the string.`
	httpDesc   = `Http functions provide pre-define function to send get, head, options, post, patch, put and delete request to the server.`
	logDesc    = `Log functions provide several pre-define functionality print or format variable to the standard output.`
	pathDesc   = `Path functions provide several pre-define functionality that can be use to manipulate or extract metadata from file path.`
	fdDesc     = `File and Directory functions provide several pre-define functionality create, delete or modified ones or more files and directories.`
)

var functions = []*functionGroup{
	{Name: "String Functions", File: "strings", Flags: function.AllStringFlags, Description: stringDesc},
	{Name: "Http Functions", File: "http", Flags: function.AllHttpFlags, Description: httpDesc},
	{Name: "Log Functions", File: "log", Flags: function.AllLogFlags, Description: logDesc},
	{Name: "Path Functions", File: "path", Flags: function.AllPathFlags, Description: pathDesc},
	{Name: "File and Directory Functions", File: "fd", Flags: function.AllFileDirectoryFlags, Description: fdDesc},
}

func main() {
	// find root path base on go.mod
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	file := filepath.Join(root, "go.mod")
	for stat, err := os.Stat(file); os.IsNotExist(err) || stat.IsDir(); stat, err = os.Stat(file) {
		if root == "/" || root == "." {
			fmt.Println("Cannot find root directory of the project.")
			os.Exit(1)
		}
		root = filepath.Dir(root)
		file = filepath.Join(root, "go.mod")
	}

	buf := bytes.NewBufferString("# Table Content\n\n")
	for i, fn := range functions {
		buf.WriteString(fmt.Sprintf("%d. [%s](%s.md)\n", i+1, fn.Name, fn.File))
		content := bytes.NewBufferString(fmt.Sprintf("# %s\n\n", fn.Name))
		content.WriteString(fn.Description)
		content.WriteString("\n\n")
		gbuf := bytes.NewBufferString("")
		for fi, flags := range fn.Flags() {
			io.Copy(gbuf, flags.HelpAsReader(true, markdownAnchor(nil, fn.Name)))
			gbuf.WriteString("\n---\n\n")
			anchor := markdownAnchor(flags, "")
			if len(flags.Aliases) > 0 {
				content.WriteString(fmt.Sprintf("%d. [%s, %s](#%s)\n", fi+1, flags.FuncName, strings.Join(flags.Aliases, ","), anchor))
			} else {
				content.WriteString(fmt.Sprintf("%d. [%s](#%s)\n", fi+1, flags.FuncName, anchor))
			}
		}
		io.Copy(content, gbuf)
		ioutil.WriteFile(filepath.Join(root, docPaths, fmt.Sprintf("%s.md", fn.File)), content.Bytes(), 0744)
	}
	ioutil.WriteFile(filepath.Join(root, docPaths, "all.md"), buf.Bytes(), 0744)
}

func markdownAnchor(flags *args.Flags, s string) string {
	if flags != nil {
		s = "@" + flags.FuncName
		if len(flags.Aliases) > 0 {
			s += ", @" + strings.Join(flags.Aliases, ", @")
		}
	}
	buf := strings.Builder{}
	for _, c := range s {
		if c == 45 || 47 < c && c < 58 || c == 95 || 96 < c && c < 123 {
			buf.WriteRune(c)
		} else if 64 < c && c < 91 {
			buf.WriteRune(c + 32)
		} else if c == ' ' || c == '\t' || c == '\n' {
			buf.WriteByte('-')
		}
	}
	return buf.String()
}
