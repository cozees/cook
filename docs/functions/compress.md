# Compress/Archive Functions

Compress, Extract or Archive functions provide several pre-define functionality that can be used to archive, compress or extract of file type tarbal, gzip, zip.

1. [compress](#compress)
2. [extract](#extract)
## @compress

Usage:
```cook
@compress [-v] [-m 0700] [-f] [--tar] [-o DIRECTORY|FILE] [-k algo] FILE
```

The Compress function compress the file or directory.        It supported format 7z(lzma), xz, zip, gzip, bzip2, tar, rar.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -k, --kind | "" | Providing compressor the algorithms to compress the data. By default gzip is used. |
| -o, --out | "" | Tell compressor where to produce the output result. It is         file name or path to the output file. |
| -t, --tar | false | Tell compressor to output as tar file |
| -f, --override | false | Tell compressor to override the output file if its exist |
| -m, --mode | "" | providing a unix like permission to apply to the output file. By default, the permission is set to 0777. |

Example:

```cook
@compress -a gzip --tar folder
```
[back top](#compressarchive-functions)

---

## @extract

Usage:
```cook
@extract [-v] [-m 0700] [-o DIRECTORY|FILE] FILE
```

The extractor function extract the file or directory from the compressed file.       It support format 7z(lzma), xz, zip, gzip, bzip2, tar, rar.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -o, --out | "" | Tell extractor where to extract file and/or folder to. If folder is not exist        extractor will create it. |
| -m, --mode | "" | override/provide permission to all file or folder extracted from compress/archive file.        By default, it apply the permission based on the permission available in the archive/compressed file        however if there is no permisson available then 0777 permission is used. |
| -v, --verbose | false | Tell extractor to display each extracted file or folder |

Example:

```cook
@extract sample.tar.gz
```
[back top](#compressarchive-functions)

---

