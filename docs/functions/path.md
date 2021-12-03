# Path Functions

Path functions provide several pre-define functionality that can be use to manipulate or extract metadata from file path.

1. [pbase](#pbase)
2. [pabs](#pabs)
3. [pclean](#pclean)
4. [pdir](#pdir)
5. [pext](#pext)
6. [psplit](#psplit)
7. [prel](#prel)
8. [pglob](#pglob)
## @pbase

Usage:
```cook
@pbase FILEPATH
```

Returns the last element of path. Trailing path separators are removed before extracting       the last element. If the path is empty, Base returns ".". If the path consists entirely       of separators, Base returns a single separator.

| Options/Flag | Default | Description |
| --- | --- | --- |

Example:

```cook
@pbase dir/file.txt
```
[back top](#path-functions)

---

## @pabs

Usage:
```cook
@pabs FILEPATH
```

Returns an absolute representation of path. If the path is not absolute it will be joined       with the current working directory to turn it into an absolute path. The absolute path name       for a given file is not guaranteed to be unique. The path is also being clean as well.

| Options/Flag | Default | Description |
| --- | --- | --- |

Example:

```cook
@pabs dir/file.txt
```
[back top](#path-functions)

---

## @pclean

Usage:
```cook
@pclean FILEPATH
```

Returns the shortest path name without "." or ".."

| Options/Flag | Default | Description |
| --- | --- | --- |

Example:

```cook
@pclean ./dir/../file.txt
```
[back top](#path-functions)

---

## @pdir

Usage:
```cook
@pdir FILEPATH
```

Returns all but the last element of path, typically the path's directory. After dropping       the final element, Dir calls Clean on the path and trailing slashes are removed.       If the path is empty, Dir returns ".". If the path consists entirely of separators, Dir       returns a single separator. The returned path does not end in a separator unless it is the root directory

| Options/Flag | Default | Description |
| --- | --- | --- |

Example:

```cook
@pdir dir/file.txt
```
[back top](#path-functions)

---

## @pext

Usage:
```cook
@pext FILEPATH
```

Returns the file name extension used by path. The extension is the suffix beginning at       the final dot in the final element of path; it is empty if there is no dot.

| Options/Flag | Default | Description |
| --- | --- | --- |

Example:

```cook
@pext dir/file.txt
```
[back top](#path-functions)

---

## @psplit

Usage:
```cook
@psplit FILEPATH
```

Returns array string of each segment in file path which separate by path separator.

| Options/Flag | Default | Description |
| --- | --- | --- |

Example:

```cook
@psplit dir/file.txt
```
[back top](#path-functions)

---

## @prel

Usage:
```cook
@prel REFERENCE_PATH TO_PATH
```

Returns a relative path that is lexically equivalent to targpath when joined to basepath       with an intervening separator. On success, the returned path will always be relative to       reference path, even if reference path and to path share no elements

| Options/Flag | Default | Description |
| --- | --- | --- |

Example:

```cook
@prel dir/a dir/sample/../a/file.txt
```
[back top](#path-functions)

---

## @pglob

Usage:
```cook
@pglob GLOB_PATTERN
```

Returns the names of all files matching pattern or nil if there is no matching file.       The syntax of patterns is the same as in Match. The pattern may describe hierarchical       names such as /usr/*/bin/ed (assuming the Separator is '/').

| Options/Flag | Default | Description |
| --- | --- | --- |

Example:

```cook
@pglob dir/*.txt
```
[back top](#path-functions)

---

