package function

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	osu "os/user"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/cozees/cook/pkg/runtime/args"
	"github.com/cozees/cook/pkg/runtime/parser"
)

const (
	read  = 04   // 0x100
	write = 02   // 0x010
	exec  = 01   // 0x001
	user  = 0700 // 0x000111000000
	group = 0070 // 0x000000111000
	other = 0007 // 0x000000000111
)

const (
	// token start from parser.CUSTOM
	A  parser.Token = parser.CUSTOM + 1 + iota // a
	U                                          // u
	S                                          // s
	G                                          // g
	O                                          // o
	R                                          // r
	W                                          // w
	X                                          // xq
	XU                                         // X
)

// common error
var errMissingOp = fmt.Errorf("missing operator +, - or =")

var fm = &FileModeParser{s: parser.NewSimpleScanner(true)}

func init() {
	fm.s.RegisterSingleCharacterToken('a', A)
	fm.s.RegisterSingleCharacterToken('u', U)
	fm.s.RegisterSingleCharacterToken('s', S)
	fm.s.RegisterSingleCharacterToken('g', G)
	fm.s.RegisterSingleCharacterToken('o', O)
	fm.s.RegisterSingleCharacterToken('r', R)
	fm.s.RegisterSingleCharacterToken('w', W)
	fm.s.RegisterSingleCharacterToken('x', X)
	fm.s.RegisterSingleCharacterToken('X', XU)
}

type FileModeParser struct {
	u, g, o, anyx int
	who           int
	op            parser.Token
	tok           parser.Token
	s             parser.Scanner
}

func (fm *FileModeParser) Parse(origin os.FileMode, mode string) (m os.FileMode, err error) {
	// if the given is already a parsed mode
	if '0' <= mode[0] && mode[0] <= '9' {
		i, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			return 0, err
		}
		return os.FileMode(i), nil
	}
	fm.init(origin.Perm(), mode)
	// release src mode bind to scanner
	defer fm.s.Finalize()

	for fm.tok != parser.EOF {
		// parse who
		fm.who = 0
		if err = fm.next(); err != nil {
			return 0, err
		}
		if err = fm.parseWho(); err != nil {
			return 0, err
		}
		// scan op
	scanOp:
		fm.op = parser.ILLEGAL
		switch fm.tok {
		case parser.ADD, parser.SUB, parser.EQL:
			fm.op = fm.tok
			if err = fm.next(); err != nil {
				return 0, err
			}
		default:
			return 0, errMissingOp
		}
		// parser permission
		if err = fm.parsePermission(); err != nil {
			return 0, err
		}
		if fm.tok != parser.COMMA && fm.tok != parser.EOF {
			goto scanOp
		}
	}
	return os.FileMode(fm.u | fm.g | fm.o), nil
}

func (fm *FileModeParser) compute(perm, permX, usetid, gsetid, override int) bool {
	if override != 0 {
		perm = override
	}
	if permX&fm.anyx == exec {
		perm = permX
	}
	switch fm.op {
	case parser.EQL:
		if fm.who == 0 || fm.who&user == user {
			fm.u = usetid | (perm << 6)
		}
		if fm.who == 0 || fm.who&group == group {
			fm.g = gsetid | (perm << 3)
		}
		if fm.who == 0 || fm.who&other == other {
			fm.o = perm
		}
		return true
	case parser.ADD:
		if fm.who == 0 || fm.who&user == user {
			fm.u |= usetid | (perm << 6)
		}
		if fm.who == 0 || fm.who&group == group {
			fm.g |= gsetid | (perm << 3)
		}
		if fm.who == 0 || fm.who&other == other {
			fm.o |= perm
		}
	case parser.SUB:
		if fm.who == 0 || fm.who&user == user {
			fm.u = fm.u &^ usetid &^ (perm << 6)
		}
		if fm.who == 0 || fm.who&group == group {
			fm.g = fm.g &^ gsetid &^ (perm << 3)
		}
		if fm.who == 0 || fm.who&other == other {
			fm.o &^= perm
		}
	}
	return false
}

func (fm *FileModeParser) parsePermission() (err error) {
	perm, permX, usetid, gsetid, computedEqual := 0, 0, 0, 0, false
	for {
		switch fm.tok {
		case R:
			perm |= read
		case W:
			perm |= write
		case X:
			perm |= exec
		case S:
			usetid = int(fs.ModeSetuid)
			gsetid = int(fs.ModeSetgid)
		case XU:
			if fm.op == parser.ADD {
				permX |= exec
			}
		case U:
			computedEqual = fm.compute(perm, permX, usetid, gsetid, fm.u>>6)
		case G:
			computedEqual = fm.compute(perm, permX, usetid, gsetid, fm.g>>3)
		case O:
			computedEqual = fm.compute(perm, permX, usetid, gsetid, fm.o)
		default:
			if usetid != 0 || gsetid != 0 || perm != 0 || fm.op == parser.EQL && !computedEqual {
				fm.compute(perm, permX, usetid, gsetid, 0)
			} else if permX != 0 {
				fm.compute(perm, permX, usetid, gsetid, 0)
			}
			return
		}
		if err = fm.next(); err != nil {
			break
		}
	}
	return
}

func (fm *FileModeParser) parseWho() (err error) {
	for {
		switch fm.tok {
		case A:
			fm.who = user | group | other
		case U:
			fm.who |= user
		case G:
			fm.who |= group
		case O:
			fm.who |= other
		default:
			return
		}
		if err := fm.next(); err != nil {
			return err
		}
	}
}

func (fm *FileModeParser) init(origin os.FileMode, mode string) {
	fm.u, fm.g, fm.o, fm.anyx = 0, 0, 0, 0
	if origin != 0 {
		efm := int(origin)
		fm.u = ((efm >> 6) % 010) << 6
		fm.g = ((efm >> 3) % 010) << 3
		fm.o = efm % 010
		fm.anyx = ((fm.u >> 6) & exec) | ((fm.g >> 3) & exec) | (fm.o & exec)
	}
	fm.op = parser.ILLEGAL
	fm.tok = parser.ILLEGAL
	fm.s.Init([]byte(mode))
}

func (fm *FileModeParser) next() (err error) {
	_, _, _, fm.tok, _, err = fm.s.Scan()
	return
}

func UnixStringPermission(m os.FileMode, isDir bool) string {
	var buf [10]byte
	var w = 1
	const rwx = "rwxrwxrwx"
	for i, c := range rwx {
		isUGset := (i == 2 || i == 5) && (m&fs.ModeSetuid == fs.ModeSetuid || m&fs.ModeSetgid == fs.ModeSetgid)
		if m&(1<<uint(9-1-i)) != 0 {
			if isUGset {
				buf[w] = 's'
			} else {
				buf[w] = byte(c)
			}
		} else {
			if isUGset {
				buf[w] = 'S'
			} else {
				buf[w] = '-'
			}
		}
		w++
	}
	buf[0] = '-'
	if isDir {
		buf[0] = 'd'
	}
	return string(buf[:w])
}

var (
	errMissingTarget    = fmt.Errorf("target is required")
	errSourceNotExisted = fmt.Errorf("source is not exist")
	errMissingPath      = fmt.Errorf("path argument is required")
)

func readMode(file, givenMode string) (os.FileMode, error) {
	mode := os.FileMode(0)
	if fmode, err := os.Stat(file); err == nil {
		mode = fmode.Mode()
	}
	return fm.Parse(mode, givenMode)
}

func readPath(f Function, o *fdOptions, nonFlagCount, from int) ([]string, error) {
	if remains := len(o.Args); nonFlagCount != -1 && remains != nonFlagCount {
		return nil, fmt.Errorf("%s required directory path to be provided", f.Name())
	} else if remains == 0 {
		return nil, errMissingPath
	} else if from > 0 {
		return o.Args[from:], nil
	} else {
		return o.Args, nil
	}
}

func readUserGroup(o *fdOptions) (u int, g int, err error) {
	u, g = -1, -1
	if len(o.Args) == 0 {
		return u, g, fmt.Errorf("user or group is required")
	}
	raw := o.Args[0]
	index := strings.IndexByte(raw, ':')
	un, gn := raw, ""
	if index != -1 {
		un = raw[0:index]
		gn = raw[index+1:]
	}
	if un != "" {
		if !o.Numguid {
			user, err := osu.Lookup(un)
			if err != nil {
				return -1, -1, err
			}
			un = user.Uid
		}
		i64, err := strconv.ParseInt(un, 10, 64)
		if err != nil {
			return -1, -1, err
		}
		u = int(i64)
	}
	if gn != "" {
		if !o.Numguid {
			group, err := osu.LookupGroup(gn)
			if err != nil {
				return -1, -1, err
			}
			gn = group.Gid
		}
		i64, err := strconv.ParseInt(gn, 10, 64)
		if err != nil {
			return -1, -1, err
		}
		g = int(i64)
	}
	return
}

func AllFileDirectoryFlags() []*args.Flags {
	return []*args.Flags{rmFlags, mkdirFlags, rmdirFlags, chmodFlags, chownFlags, cpFlags, mvFlags, chdirFlags}
}

type fdOptions struct {
	Recursive bool   `flag:"recursive"`
	Mode      string `flag:"mode,740"`
	Numguid   bool   `flag:"guinum"`
	Args      []string
}

func removeAll(folderOnly bool, f Function, i interface{}) (interface{}, error) {
	opts := i.(*fdOptions)
	paths, err := readPath(f, opts, -1, 0)
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		if folderOnly || !opts.Recursive {
			stat, err := os.Stat(path)
			if err != nil {
				return nil, err
			}
			if folderOnly {
				if !stat.IsDir() {
					return nil, fmt.Errorf("%s is not a directory", path)
				}
			} else if stat.IsDir() {
				return nil, fmt.Errorf("%s is a directory use rmdir or add flag -r", path)
			}
		}
		if opts.Recursive {
			if folderOnly {
				if filepath.IsAbs(path) {
					return nil, fmt.Errorf("%s is an aboslute path, only relative path is allowed when -p flag is given", path)
				}
				// get top directory name from path: test/abc/dir1 => test
				for a := filepath.Dir(path); a != "."; a = filepath.Dir(path) {
					path = a
				}
			}
			if err = os.RemoveAll(path); err != nil {
				return nil, err
			}
		} else if err = os.Remove(path); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func copyFile(a, b string) error {
	f1, err := os.Open(a)
	if err != nil {
		return err
	}
	defer f1.Close()
	f1Stat, err := f1.Stat()
	if err != nil {
		return err
	} else if f1Stat.IsDir() {
		return fmt.Errorf("%s is not a file, to copy directory use -R", a)
	}
	f2, err := os.OpenFile(b, os.O_CREATE|os.O_WRONLY, f1Stat.Mode())
	if err != nil {
		return err
	}
	defer f2.Close()
	if cp, err := io.Copy(f2, f1); err != nil {
		return err
	} else if cp != f1Stat.Size() {
		return fmt.Errorf("copy failed, only %d out of %d bytes was copied", cp, f1Stat.Size())
	}
	return nil
}

func moveOrCopyGlob(sources []string, dir string, action func(a, b string) error) error {
	for _, path := range sources {
		if err := action(path, filepath.Join(dir, filepath.Base(path))); err != nil {
			return err
		}
	}
	return nil
}

func moveOrCopy(f Function, i interface{}, action func(a, b string) error) (interface{}, error) {
	paths, err := readPath(f, i.(*fdOptions), -1, 0)
	if err != nil {
		return nil, err
	}
	ic := len(paths)
	if ic < 2 {
		return nil, errMissingTarget
	}
	stat, err := os.Stat(paths[ic-1])
	isDirAndExist := func(stat os.FileInfo, err error) error {
		if os.IsNotExist(err) || err != nil {
			return err
		}
		if !stat.IsDir() {
			return fmt.Errorf("last argument %s must be an existed directory", paths[ic-1])
		}
		return nil
	}
	if ic > 2 {
		if err = isDirAndExist(stat, err); err != nil {
			return nil, err
		}
		ic--
		for i := 0; i < ic; i++ {
			gpaths, err := filepath.Glob(paths[i])
			if err != nil {
				return nil, err
			}
			if err = moveOrCopyGlob(gpaths, paths[ic], action); err != nil {
				return nil, err
			}
		}
		return nil, nil
	} else if ls, err := filepath.Glob(paths[0]); err != nil {
		return nil, err
	} else if ls == nil {
		return nil, errSourceNotExisted
	} else if len(ls) > 1 || ls[0] != paths[0] {
		if err = isDirAndExist(stat, err); err != nil {
			return nil, err
		}
		return nil, moveOrCopyGlob(ls, paths[1], action)
	} else {
		// not glob pattern
		return nil, action(paths[0], paths[1])
	}
}

const (
	pathDesc           = `It fine to use linux file path syntax on any platform.`
	permissoinDesc     = `The function utilize linux permission mode syntax to set permission.`
	permissionFlagDesc = `The linux permission syntax is required in order provide the permission other than default permission 740.`
)

const (
	mkdirRecurDesc = `Create directories recursively if any directory in the given path is not exist.
					  By default, if permission mode is not given then a permission 740 is used.`
	mkdirMode = `Set directory permission. ` + permissionFlagDesc
	mkdirDesc = `Create one or multiple directories. ` + permissoinDesc + ` ` + pathDesc
)

var fdOptionsType = reflect.TypeOf((*fdOptions)(nil)).Elem()

var mkdirFlags = &args.Flags{
	Flags: []*args.Flag{
		{Short: "p", Long: "recursive", Description: mkdirRecurDesc},
		{Short: "-m", Long: "mode", Description: mkdirMode},
	},
	Result:      fdOptionsType,
	FuncName:    "mkdir",
	ShortDesc:   "create directories",
	Usage:       "@mkdir [-p] [-m permission] PATH",
	Example:     "@mkdir -p -m 755 dir1/dir2/dir3",
	Description: mkdirDesc,
}

var rmdirFlags = &args.Flags{
	Flags: []*args.Flag{
		{Short: "p", Long: "recursive", Description: "Remove all empty child directories include current directory as well."},
	},
	Result:      fdOptionsType,
	FuncName:    "rmdir",
	ShortDesc:   "remove directories",
	Usage:       "@rmdir [-p] PATH",
	Example:     "@rmdir -p dir1/dir2/dir3",
	Description: `Remove one or multiply empty directory in the hierarchy. ` + pathDesc,
}

var rmFlags = &args.Flags{
	Flags: []*args.Flag{
		{Short: "r", Long: "recursive", Description: "Remove all file or directory in hierarchy of the given directory."},
	},
	Result:    fdOptionsType,
	FuncName:  "rm",
	ShortDesc: "remove files or directories",
	Usage:     "@rm [-p] PATH",
	Example:   "@rm -p dir1/dir2/dir3",
	Description: `Remove one or more files and directory in the hierarchy. If the given path is a file then only that that is remove
				  however it was a directory then it content including the directory itself will be remove.` + pathDesc,
}

var chdirFlags = &args.Flags{
	Result:      fdOptionsType,
	FuncName:    "workin",
	ShortDesc:   "change working directory",
	Usage:       "@workin PATH",
	Example:     "@workin dir1/dir2/dir3",
	Description: `Change current working directory to the given directory.` + pathDesc,
}

var chownFlags = &args.Flags{
	Flags: []*args.Flag{
		{Short: "n", Long: "guinum", Description: `Tell @chown that the given user and/or group id is a numeric id. 
												   By default, @chown treat the given user or group as a username or group name
												   which required lookup to find a numeric representation of user or group id.`},
		{Short: "r", Long: "recursive", Description: `Tell @chown to change owner of all file or directory in the hierarchy.`},
	},
	Result:      fdOptionsType,
	FuncName:    "chown",
	ShortDesc:   "Change owner and/or group of file or directory",
	Usage:       "@chown [-rn] [user][:group] PATH",
	Example:     "@chown -r user1:group1 dir1/dir2/dir3",
	Description: `Change owner and/or group of file or directory.` + pathDesc,
}

var chmodFlags = &args.Flags{
	Flags: []*args.Flag{
		{Short: "r", Long: "recursive", Description: `Tell @chmode to change permission of all file or directory in the hierarchy.`},
	},
	Result:      fdOptionsType,
	FuncName:    "chmod",
	ShortDesc:   "Change permission of files or directories",
	Usage:       "@chown [-r] mode PATH [PATH ...]",
	Example:     "@chown -r u+x,g-w dir1 dir2/dir3",
	Description: `Change permission mode of files or directories.` + pathDesc,
}

var mvFlags = &args.Flags{
	Result:      fdOptionsType,
	FuncName:    "mv",
	ShortDesc:   "Move a file or directorie",
	Usage:       "@mv TARGET_PATH DESTRINATION_PATH",
	Example:     "@mv dir1 dir2/dir3",
	Description: `Move a target file or directory to the destination path which must be an existed directory.` + pathDesc,
}

var cpFlags = &args.Flags{
	Flags: []*args.Flag{
		{Short: "r", Long: "recursive", Description: `Copies the directory and the entire sub-tree to the target. To copy the content only add trailing /.`},
	},
	Result:      fdOptionsType,
	FuncName:    "cp",
	ShortDesc:   "Copy files or directories",
	Usage:       "@cp [-r] mode PATH [PATH ...] NEW_PATH",
	Example:     "@cp dir1 file.txt dir2/dir3",
	Description: `Copy one or more of files or directories. If the target is not exist the @cp will create like call @mkdir -p. ` + pathDesc,
}

func init() {
	registerFunction(NewBaseFunction(mkdirFlags, func(f Function, i interface{}) (interface{}, error) {
		opts := i.(*fdOptions)
		paths, err := readPath(f, opts, -1, 0)
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			m, err := readMode(path, opts.Mode)
			if err != nil {
				return nil, err
			}
			if opts.Recursive {
				if err = os.MkdirAll(path, os.FileMode(m)); err != nil {
					return nil, err
				}
			} else if err = os.Mkdir(path, os.FileMode(m)); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}))

	registerFunction(NewBaseFunction(rmdirFlags, func(f Function, i interface{}) (interface{}, error) {
		return removeAll(true, f, i)
	}))

	registerFunction(NewBaseFunction(rmFlags, func(f Function, i interface{}) (interface{}, error) {
		return removeAll(false, f, i)
	}))

	registerFunction(NewBaseFunction(chdirFlags, func(f Function, i interface{}) (interface{}, error) {
		paths, err := readPath(f, i.(*fdOptions), 1, 0)
		if err != nil {
			return nil, err
		}
		return nil, os.Chdir(paths[0])
	}, "chdir"))

	registerFunction(NewBaseFunction(chownFlags, func(f Function, i interface{}) (interface{}, error) {
		opts := i.(*fdOptions)
		// must call ownergroup before path
		u, g, err := readUserGroup(opts)
		if err != nil {
			return nil, err
		}
		paths, err := readPath(f, opts, -1, 1)
		if err != nil {
			return nil, err
		}
		if opts.Recursive {
			for _, path := range paths {
				err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
					return os.Chown(path, u, g)
				})
				if err != nil {
					return nil, err
				}
			}
		} else {
			for _, path := range paths {
				if err = os.Chown(path, u, g); err != nil {
					return nil, err
				}
			}
		}
		return nil, nil
	}))

	registerFunction(NewBaseFunction(chmodFlags, func(f Function, i interface{}) (interface{}, error) {
		opts := i.(*fdOptions)
		paths, err := readPath(f, opts, -1, 1)
		if err != nil {
			return nil, err
		}
		chmod := func(path string) error {
			m, err := readMode(path, opts.Args[0])
			if err != nil {
				return err
			} else if err = os.Chmod(path, m); err != nil {
				return err
			}
			return nil
		}
		if opts.Recursive {
			for _, path := range paths {
				err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
					return chmod(path)
				})
				if err != nil {
					return nil, err
				}
			}
		} else {
			for _, path := range paths {
				if err = chmod(path); err != nil {
					return nil, err
				}
			}
		}
		return nil, nil
	}))

	registerFunction(NewBaseFunction(mvFlags, func(f Function, i interface{}) (interface{}, error) {
		return moveOrCopy(f, i, func(a, b string) error { return os.Rename(a, b) })
	}, "move"))

	registerFunction(NewBaseFunction(cpFlags, func(f Function, i interface{}) (interface{}, error) {
		opts := i.(*fdOptions)
		return moveOrCopy(f, i, func(a, b string) error {
			if opts.Recursive {
				stata, err := os.Stat(a)
				if os.IsNotExist(err) || err != nil {
					return err
				}
				if stata.IsDir() {
					// copy recursive include directory as well
					if a[len(a)-1] != os.PathSeparator {
						b = filepath.Join(b, filepath.Base(a))
					}
					statb, err := os.Stat(b)
					if os.IsNotExist(err) {
						if err = os.MkdirAll(b, stata.Mode()); err != nil {
							return err
						}
					} else if !statb.IsDir() {
						return fmt.Errorf("target %s is not a directory", b)
					}
					return filepath.WalkDir(a, func(path string, d fs.DirEntry, err error) error {
						var rel string
						if err == nil {
							rel, err = filepath.Rel(a, path)
							if err != nil {
								return err
							}
							if d.IsDir() {
								var di os.FileInfo
								di, err = d.Info()
								if err != nil {
									return err
								}
								err = os.MkdirAll(filepath.Join(b, rel), di.Mode())
							} else {
								err = copyFile(path, filepath.Join(b, rel))
							}
						}
						return err
					})
				}
			}

			statb, err := os.Stat(b)
			// check if b is directory the copy in form b/base(a)
			if (err == nil || os.IsExist(err)) && statb.IsDir() {
				b = filepath.Join(b, filepath.Base(a))
			}
			return copyFile(a, b)
		})
	}, "copy"))
}
