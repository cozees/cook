# Log Functions

Log functions provide several pre-define functionality print or format variable to the standard output.

1. [print](#print)
## @print

Usage:
```cook
@print [-ens] ARG [ARG ...]
```

The print function write the arguments as the string into standard output if flag "echo"      is not given otherwise the string result is return from the function instead.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -e, --echo | false | Tell print function to return the result instead of writing the result in standard output. |
| -n, --omitln | false | Tell print function to not add a newline at the end of the result. |
| -s, --strip | false | Tell print function to remove all leading and trailing whitespace from each given argument. |

Example:

```cook
@print -e text
```
[back top](#log-functions)

---

