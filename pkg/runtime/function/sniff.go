package function

import "bytes"

type FileType uint

const (
	Unknown FileType = iota
	ZipFile
	TarFile
	GzipFile
	RarFile
)

var ftstring = [...]string{
	Unknown:  "unknown",
	ZipFile:  "zip",
	TarFile:  "tarball",
	GzipFile: "gzip",
	RarFile:  "rar",
}

func (ft FileType) String() string { return ftstring[ft] }

type FileSig struct {
	sig    [][]byte
	offset int
	ftype  FileType
}

func (fs *FileSig) match(data []byte) bool {
	if fs.offset < 0 || len(data) == 0 {
		return false
	}
	if fs.offset > 0 {
		if fs.offset > len(data) {
			return false
		}
	}
	for _, sp := range fs.sig {
		if bytes.HasPrefix(data[fs.offset:], sp) {
			return true
		}
	}
	return false
}

var sniffSignatures = []*FileSig{
	{[][]byte{[]byte("PK\x03\x04")}, 0, ZipFile},
	{[][]byte{[]byte("ustar\u000000"), []byte("ustar\040\040\u0000")}, 257, TarFile},
	{[][]byte{[]byte("\x1F\x8B\x08")}, 0, GzipFile},
	{[][]byte{[]byte("Rar!\x1A\x07\x00")}, 0, RarFile},
	{[][]byte{[]byte("Rar!\x1A\x07\x01\x00")}, 0, RarFile},
}

func FileDataType(data []byte) FileType {
	for _, sns := range sniffSignatures {
		if sns.match(data) {
			return sns.ftype
		}
	}
	return Unknown
}
