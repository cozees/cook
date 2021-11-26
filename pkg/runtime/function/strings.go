package function

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/cozees/cook/pkg/runtime/parser"
)

type stringSplitOption struct {
	ws   bool
	line bool
	by   string
	reg  string
	rc   string
	args []string
	*options
}

func (so *stringSplitOption) reset() {
	so.ws = false
	so.line = false
	so.by = ""
	so.reg = ""
	so.rc = ""
	so.options.args = nil
}

func (so *stringSplitOption) copy() interface{} {
	return &stringSplitOption{
		ws:   so.ws,
		line: so.line,
		by:   so.by,
		reg:  so.reg,
		rc:   so.rc,
		args: so.options.args,
	}
}

func (so *stringSplitOption) rowColumn() (row, column int, err error) {
	row, column = -1, -1
	if so.rc == "" {
		return
	}
	index, i := strings.IndexByte(so.rc, ':'), int64(-1)
	if index == 0 {
		if i, err = strconv.ParseInt(so.rc[1:], 10, 32); err != nil {
			return
		}
		column = int(i)
	} else if index > 0 {
		if i, err = strconv.ParseInt(so.rc[:index], 10, 32); err != nil {
			return
		}
		row = int(i)
		if i, err = strconv.ParseInt(so.rc[index+1:], 10, 32); err != nil {
			return
		}
		column = int(i)
	} else if i, err = strconv.ParseInt(so.rc, 10, 32); err == nil {
		row = int(i)
	}
	return
}

func newStringSplitOptions(fs *flag.FlagSet) *stringSplitOption {
	opts := &stringSplitOption{options: &options{}}
	opts.options.opts = opts
	return opts
}

type stringReplaceOption struct {
	reg     bool   // true if search is regular expression
	line    string // a line to replace, format single line numer 1 or multiple line with comma 1,1,11
	inPlace bool   // replace string in file in place, do not create new file
	args    []string
	*options
}

func (so *stringReplaceOption) reset() {
	so.reg = false
	so.inPlace = false
	so.line = ""
	so.options.args = nil
}

func (so *stringReplaceOption) copy() interface{} {
	notInPlace := len(so.options.args) == 4 &&
		strings.HasPrefix(so.options.args[2], "@") &&
		strings.HasPrefix(so.options.args[3], "@")
	return &stringReplaceOption{
		reg:     so.reg,
		line:    so.line,
		inPlace: !notInPlace,
		args:    so.options.args,
	}
}

const bufferSize = 2048

func newStringReplaceOptions(fs *flag.FlagSet) *stringReplaceOption {
	opts := &stringReplaceOption{options: &options{}}
	opts.options.opts = opts
	return opts
}

func replaceString(src, search, replace string) (pending, result string) {
	n := strings.Count(src, search)
	if n == 0 {
		if len(src) < len(search) {
			return src, ""
		} else {
			l := len(src) - len(search)
			return src[l:], src[:l]
		}
	}
	var b strings.Builder
	acc := len(src)
	b.Grow(acc + n*(len(replace)-len(search)))
	start := 0
	for i := 0; i < n; i++ {
		j := start
		if len(search) == 0 {
			if i > 0 {
				_, wid := utf8.DecodeRuneInString(src[start:])
				j += wid
			}
		} else {
			j += strings.Index(src[start:], search)
		}
		b.WriteString(src[start:j])
		b.WriteString(replace)
		start = j + len(search)
	}
	minial := acc - len(search)
	if start < minial {
		b.WriteString(src[start:minial])
		start = minial
	}
	return src[start:], b.String()
}

type stringPadOptions struct {
	left  int
	right int
	by    string
	max   int
	args  []string
	*options
}

func (spo *stringPadOptions) reset() {
	spo.left = 0
	spo.right = 0
	spo.by = ""
	spo.max = -1
	spo.options.args = nil
}

func (spo *stringPadOptions) copy() interface{} {
	return &stringPadOptions{
		left:  spo.left,
		right: spo.right,
		by:    spo.by,
		max:   spo.max,
		args:  spo.options.args,
	}
}

func newStringPadOption(fs *flag.FlagSet) Option {
	opts := &stringPadOptions{options: &options{}}
	opts.options.opts = opts
	fs.IntVar(&opts.left, "l", 0, "padding left side by character or string given by \"by\" flag")
	fs.IntVar(&opts.right, "r", 0, "padding right side by character or string given by \"by\" flag")
	fs.IntVar(&opts.max, "m", -1, "maximum length character of the ouput result")
	flagString(fs, &opts.by, "by", " ", "character or string use to add to left or right the string")
	return opts
}

func pad(left, right, max int, by, arg string) string {
	if max > 0 && len(arg) >= max {
		return arg
	}
	// reduce left and right paddingt o fit max length
	if max > 0 {
		a := false
		for left*len(by)+right*len(by)+len(arg) > max {
			if a {
				left--
			} else {
				right--
			}
			a = !a
		}
	}
	if left > 0 {
		arg = strings.Repeat(by, left) + arg
	}
	if right > 0 {
		arg += strings.Repeat(by, right)
	}
	return arg
}

func init() {
	registerFunction(&GeneralFunction{
		name:     "spad",
		flagInit: func(fs *flag.FlagSet) Option { return newStringPadOption(fs) },
		handler: func(gf *GeneralFunction, i interface{}) (interface{}, error) {
			opts := i.(*stringPadOptions)
			switch si := len(opts.args); si {
			case 0:
				return nil, fmt.Errorf("spad required at least once argument")
			case 1:
				if opts.left == 0 && opts.right == 0 {
					return opts.args[0], nil
				}
				return pad(opts.left, opts.right, opts.max, opts.by, opts.args[0]), nil
			default:
				if opts.left == 0 && opts.right == 0 {
					return opts.args, nil
				}
				result := make([]string, si)
				for i, arg := range opts.args {
					result[i] = pad(opts.left, opts.right, opts.max, opts.by, arg)
				}
				return result, nil
			}
		},
	})

	registerFunction(&GeneralFunction{
		name: "ssplit",
		flagInit: func(fs *flag.FlagSet) Option {
			opts := newStringSplitOptions(fs)
			flagBool(fs, &opts.ws, "S", false, "split string into array or column by whitespace.")
			flagBool(fs, &opts.line, "L", false, "split string into array or rows.")
			flagString(fs, &opts.by, "by", "", "split string into array by the given delimited")
			flagString(fs, &opts.reg, "reg", "", "split string into array with regular expression")
			flagString(fs, &opts.rc, "rc", "", "A row:column value use inconjuntion to flag S or by and L."+
				"It only effected if S or by flag is specified along with L flag."+
				"result as string if row and column specified otherwise if a row or a column is given then a string array is return."+
				"if not given at all then a string array of two dimension is return.")
			return opts
		},
		handler: func(gf *GeneralFunction, i interface{}) (interface{}, error) {
			opts := i.(*stringSplitOption)
			// validate the argument
			if len(opts.args) > 1 || len(opts.args) == 0 {
				return nil, fmt.Errorf("split required a single ascii or unicode utf-8 argument")
			} else if !utf8.ValidString(opts.args[0]) {
				return nil, fmt.Errorf("argument %s is not a valid unicode utf-8", opts.args[0])
			} else if opts.line && opts.by != "" && strings.ContainsAny(opts.by, "\n\r") {
				return nil, fmt.Errorf("flag \"-by\" must not contain any newline character if flag \"-L\" is given")
			}
			// regular expression is produce array of string only
			if opts.reg != "" {
				reg, err := regexp.Compile(opts.reg)
				if err != nil {
					return nil, fmt.Errorf("invalid regular expression %s: %w", opts.reg, err)
				}
				return reg.Split(opts.args[0], -1), nil
			} else if opts.by != "" || opts.ws || opts.line {
				row, col, err := opts.rowColumn()
				if err != nil {
					return nil, err
				} else if !opts.line && (row != -1 || col != -1) {
					return nil, fmt.Errorf("flag \"-L\" is required when specified flag \"-rc\"")
				}
				s := parser.NewSimpleScanner(false)
				s.Init([]byte(opts.args[0]))
				offs, cur, cr, cc := 0, 0, 0, 0
				var result, array []interface{}
				var seg string
				isCell := row != -1 && col != -1
				var asg func(ln bool, s string) bool
				if !isCell {
					result = make([]interface{}, 0, len(opts.args))
					array = make([]interface{}, 0, 1)
					asg = func(ln bool, s string) bool {
						if row != -1 || col != -1 {
							if row == cr || col == cc {
								result = append(result, s)
								return row == cr && ln
							}
						} else if opts.line {
							array = append(array, s)
							if ln {
								result = append(result, array)
								array = make([]interface{}, 0, 1)
							}
						} else {
							result = append(result, s)
						}
						return false
					}
				} else {
					asg = func(ln bool, s string) bool { return row == cr && col == cc }
				}
			innerLoop:
				for ch, err := s.Next(); err == nil && ch != parser.RuneEOF; ch, err = s.Next() {
				revisit:
					cur = s.Offset()
					switch {
					case ch == '\r':
						if s.Peek() == '\n' {
							if _, err = s.Next(); err != nil {
								return nil, err
							}
						}
						fallthrough
					case ch == '\n':
						if opts.line {
							if seg = opts.args[0][offs:cur]; asg(true, seg) {
								goto conclude
							}
							offs = s.NextOffset()
							cr++
							cc = 0
							continue innerLoop
						}
						fallthrough
					case unicode.IsSpace(ch):
						if opts.ws {
							if seg = opts.args[0][offs:cur]; asg(false, seg) {
								goto conclude
							}
							cc++
							offs = s.NextOffset()
						}
					case strings.IndexRune(opts.by, ch) == 0:
						// only if it start with ch
						ni := utf8.RuneLen(ch)
						if ni == len(opts.by) {
							if seg = opts.args[0][offs:cur]; asg(false, seg) {
								goto conclude
							}
							offs = s.NextOffset()
							continue innerLoop
						}
						for {
							if ch, err = s.Next(); err != nil {
								return nil, err
							} else if ni < len(opts.by) && strings.IndexRune(opts.by[ni:], ch) == 0 {
								ni += utf8.RuneLen(ch)
								if ni == len(opts.by) { // match delimiter "by"
									if seg = opts.args[0][offs:cur]; asg(false, seg) {
										goto conclude
									}
									offs = s.NextOffset()
									continue innerLoop
								}
								continue
							}
							break
						}
						if opts.line && (ch == '\r' || ch == '\n') {
							goto revisit
						}
					}
				}
				if offs < len(opts.args[0]) {
					if seg = opts.args[0][offs:]; asg(false, seg) {
						goto conclude
					}
				}
				if len(array) > 0 {
					if opts.line {
						result = append(result, array)
					} else {
						result = append(result, array...)
					}
				}
			conclude:
				if isCell {
					return seg, nil
				} else {
					return result, err
				}
			} else {
				return nil, fmt.Errorf("required at least one split flag available")
			}
		},
	})

	registerFunction(&GeneralFunction{
		name: "sreplace",
		flagInit: func(fs *flag.FlagSet) Option {
			opts := newStringReplaceOptions(fs)
			flagBool(fs, &opts.reg, "e", false, "indicate that the search in put is an regular expression")
			flagString(fs, &opts.line, "L", "", "line which replace should occurred")
			return opts
		},
		handler: func(gf *GeneralFunction, i interface{}) (interface{}, error) {
			opts := i.(*stringReplaceOption)
			sargs := len(opts.args)
			if (opts.inPlace && sargs != 3) || (!opts.inPlace && sargs != 4) {
				return nil, fmt.Errorf("required three or four argument")
			}
			// if search and replace value is the same then there is need to run replace function
			if opts.args[0] == opts.args[1] {
				if strings.HasPrefix(opts.args[2], "@") {
					return nil, nil
				} else {
					return opts.args[2], nil
				}
			}
			// a file input
			if strings.HasPrefix(opts.args[2], "@") {
				var lines []int
				if opts.line != "" {
					if strings.IndexByte(opts.line, ',') != -1 {
						for _, lstr := range strings.Split(opts.line, ",") {
							l, err := strconv.ParseInt(lstr, 10, 32)
							if err != nil {
								return nil, fmt.Errorf("invalid line format %s from %s: %w", lstr, opts.line, err)
							}
							lines = append(lines, int(l))
						}
						sort.Ints(lines)
					} else if l, err := strconv.ParseInt(opts.line, 10, 32); err != nil {
						return nil, fmt.Errorf("invalid line format %s: %w", opts.line, err)
					} else {
						lines = append(lines, int(l))
					}
				}
				// result is alway written to a newfile
				var file = opts.args[2][1:]
				var fileBC string
				if opts.inPlace {
					fileBC = filepath.Join(os.TempDir(), ".cook-replace."+file)
					defer os.Remove(fileBC)
				} else {
					fileBC = opts.args[3][1:]
				}
				// open file read & fileBC write
				f, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0700)
				if err != nil {
					return nil, err
				}
				defer f.Close()
				fstat, err := f.Stat()
				if err != nil {
					return nil, err
				}
				f2, err := os.OpenFile(fileBC, os.O_CREATE|os.O_RDWR, fstat.Mode())
				if err != nil {
					return nil, err
				}
				defer f2.Close()
				// this probably no perfect there is corner case there the unicode string was happen to split
				// at the end of bufferSize the regular expression or normal replace can be found
				byteCount, cce := int64(0), 0
				if opts.reg {
					buf := bufio.NewReader(f)
					var reg *regexp.Regexp
					reg, err = regexp.Compile(opts.args[0])
					if err != nil {
						return nil, err
					}
					line := 1
					for {
						src, errRead := buf.ReadString('\n')
						if src != "" {
							switch {
							case len(lines) > 0 && lines[0] == line:
								lines = lines[1:]
								fallthrough
							case lines == nil:
								if cce, err = f2.WriteString(reg.ReplaceAllString(src, opts.args[1])); err != nil {
									return nil, err
								}
							default:
								if cce, err = f2.WriteString(src); err != nil {
									return nil, err
								}
							}
							line++
							byteCount += int64(cce)
						}
						if errRead == io.EOF {
							break
						}
						if errRead != nil {
							return nil, err
						}
					}
				} else {
					buf := [bufferSize]byte{}
					prev, result, n, line := "", "", 0, 1
					ireader := bufio.NewReader(f)
					var errRead error
					for {
						if len(lines) == 0 {
							n, errRead = ireader.Read(buf[:])
							if errRead != nil && errRead != io.EOF {
								return nil, err
							}
							if lines == nil {
								prev, result = replaceString(prev+string(buf[:n]), opts.args[0], opts.args[1])
							} else {
								result = string(buf[:n])
							}
							if cce, err = f2.WriteString(result); err != nil {
								return nil, err
							}
						} else {
							n = bufferSize
							lstr, errRead := ireader.ReadString('\n')
							if errRead != nil && errRead != io.EOF {
								return nil, err
							} else if err == io.EOF {
								n = 0
							}
							if lines[0] == line {
								if cce, err = f2.WriteString(strings.ReplaceAll(lstr, opts.args[0], opts.args[1])); err != nil {
									return nil, err
								}
								lines = lines[1:]
							} else if cce, err = f2.WriteString(lstr); err != nil {
								return nil, err
							}
							line++
						}
						byteCount += int64(cce)
						// n is less than bufferSize then there is no more data to read
						if errRead == io.EOF {
							if len(prev) > 0 {
								if cce, err = f2.WriteString(prev); err != nil {
									return nil, err
								}
								byteCount += int64(cce)
							}
							break
						}
					}
				}
				// copy content of new file to old if in place
				if opts.inPlace {
					if err = f.Truncate(byteCount); err == nil {
						var exn int64
						if exn, err = f.Seek(0, 0); err == nil && exn == 0 {
							if exn, err = f2.Seek(0, 0); err == nil && exn == 0 {
								if exn, err = io.Copy(f, f2); exn == byteCount && err == nil {
									return nil, nil
								}
								err = fmt.Errorf("only %d out of %d byte(s) have been copied", exn, byteCount)
							}
						}
						if err == nil {
							err = fmt.Errorf("file seek offset expected 0 but get %d", exn)
						}
					}
				}
				return nil, err
			}
			if opts.reg {
				reg, err := regexp.Compile(opts.args[0])
				if err != nil {
					return nil, err
				}
				return reg.ReplaceAllString(opts.args[2], opts.args[1]), nil
			} else {
				return strings.ReplaceAll(opts.args[2], opts.args[0], opts.args[1]), nil
			}
		},
	})
}
