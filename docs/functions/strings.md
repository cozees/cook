# String Functions

String functions provide several pre-define functionality that can be used to manipulate the string.

1. [sreplace](#sreplace)
2. [ssplit](#ssplit)
3. [spad](#spad)
## @sreplace

Usage:
```cook
@sreplace [-x] [--line value,...] {regular|string} {replacement} STRING [@OUT]
```

Replace a string of the first given argument in a file or a given string with      a new string given by the second arguments. If the given old string is an empty      string then the new string will be place after each unicode character in the given      string. If the fourth argument is given it must begin with an @ character to indicate      that the replacement should written to that file instead regardless if the third      argument is a string or a file which also begin with an @. Note: when replace the string      by regular expression, the function @sreplace replace each string by line instead of a while      file.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -x, --regx | false | Tell @sreplace that the first argument is a regular expression rather than a normal string.                     Also note that when first argument is a regular expression then second argument can also use         regular expression variable (${number}) as the replacement as well. |
| -l, --line | "" |  |

Example:

```cook
@sreplace -l 1 sample elpmas "sample text"
			  @sreplace -x -l 1,2,3 "(\d+)x" "0x${1}" @file
			  @sreplace -x "(\d+)x" "0x${1}" @file @out
```
[back top](#string-functions)

---

## @ssplit

Usage:
```cook
@ssplit [-l] [--ws] [--by value] [--regx expression] [--rc row:column] STRING
```

Split a string into array or table depend on the given flag. The split function required input to be       a regular string or a unicode string, a redirect syntax to split a non-text file will result with       unknown behavior.

| Options/Flag | Default | Description |
| --- | --- | --- |
| --ws | false | Tell @ssplit to split the string by any whitespace character.      If flag --line is given then @ssplit will split each line into row result in table instead of array. |
| -l, --line | false | Tell @ssplit to split the string by line into string array or table depend on flag --ws. |
| --by | "" | Tell @ssplit to split the string into array or table by given string. If flag --by is space it is     similar to flag --ws except that it ignore other whitespace character such newline or tab. If flag     --ws and --by is given at the same time then @ssplit will ignore flag --by. |
| --regx | "" | Tell @ssplit to split the string into array using the given regular expression. Split with Regular      Expression does not support split by line flag thus it only output array of string. If flag --regx     is given then other flag will be ignored. |
| --rc | "" | Tell @ssplit to return a single string at given row and column instead of array or table. The flag --rc      use conjunction with other flag, for example, if a row value is given then it's also required flag --line      to be given as well otherwise @ssplit will return an error instead. |

Example:

```cook
// result in table or array 2 dimension [[a,b,c],[d,e,f]]
@ssplit --ws -l "a b c
d e f"
```
[back top](#string-functions)

---

## @spad

Usage:
```cook
@spad [--left value] [--right value] [--max value] [--by value] STRING
```

Pads the given string with another string given by "--by" flag until the resulting string is satisfied the        given number to left and the right or it reach the maximum length. If number of total character exceeded the maximum        given by --max flag then the result will be truncated.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -l, --left | 0 | number of string to be pads left of a string. It is number of time a string given with flag                  --by to be repeated and concatenate to left. |
| -r, --right | 0 | number of string to be pads right of a string. It is number of time a string given with flag       --by to be repeated and concatenate to right. |
| -m, --max | 0 | A total maximum number of character allowed. This number of unicode character is compare with the padding result. |
| --by | "" | The string which use for padding, if it is empty or not given then the original argument is return instead. |

Example:

```cook
@spad -l 5 -r 5 -m 12 --by 0 ii
```
[back top](#string-functions)

---

