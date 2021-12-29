//go:build !windows

package function

import (
	"fmt"
	"os"
	"strconv"
)

func Chmod(file string, raw string) error {
	stat, err := os.Stat(file)
	if err != nil {
		return err
	}
	mode := stat.Mode()
	perm, err := fm.Parse(mode.Perm(), raw)
	if err != nil {
		return err
	}
	return os.Chmod(file, (mode>>9)<<9|perm)
}

func chown(file string, uid, gid string) (err error) {
	var u, g = int64(-1), int64(-1)
	if uid != "" {
		if u, err = strconv.ParseInt(uid, 10, 64); err != nil {
			return err
		}
	}
	if gid != "" {
		if g, err = strconv.ParseInt(gid, 10, 64); err != nil {
			return err
		}
	}
	if u == -1 && g == -1 {
		// this should be have by optional that make sure uid and gid at one is provided
		return fmt.Errorf("user id %s and group id %s is not available", uid, gid)
	}
	return os.Chown(file, int(u), int(g))
}

// provide for test file only
func GetFDModePerm(file string) (os.FileMode, error) {
	stat, err := os.Stat(file)
	if err != nil {
		return 0, err
	} else {
		return stat.Mode().Perm(), nil
	}
}

func GetFDStat(file string) (stat os.FileInfo, err error) { return os.Stat(file) }
