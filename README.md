# Cook Introduction

![icon](docs/icon.mark.png)

A simple interpreter language to read and execute cook statement/instruction in Cookfile. 
Cookfile syntax was inspire by Go and the tools gnu make. Cook aiming to provide cross-platform
compatibility and simplicity.

Although all Cook functionality is being test on Linux, MacOS and Windows, it is currently at it early stage.

# Usage

Download the binary from the [release](https://github.com/cozees/cook/releases/tag/0.0.1.alpha) page on github and add the path to the binary
executable in your variable environment.

Ultimately, if you have [Go](https://github.com/golang/go) installed on your machine then you can build Cook from source code with command below:

```shell
git clone https://github.com/cozees/cook.git
cd cook
go build -o cook cmd/main.go
```

**Note:** in our release page, we include a binary compression with smaller size foot print which tested against all cook functionality to ensure that is it running fine on major plaform such as Linux, MacOS and Windows.

# Languages

Cook design to be simple while having most basic need to perform as scripting language thus Cook does not support complex data type such class struct or other similar data type.

## Basic Data Type

1. **Integer**, decimal syntax only
2. **Float** or Double
3. **Boolean** (True/False)
4. **String**
5. **Array**
6. **Map** or Dictionary

There are only 4 types **integer**, **float**, **boolean** and **string** can have type casting other type will be rejected.

|         | String | Float  | Integer | Boolean |
|---------|--------|--------|---------|---------|
| String  | Yes    | Yes    | Yes     | Yes     |
| Float   | Yes    | Yes    | Yes     |         |
| Integer | Yes    | Yes    | Yes     |         |
| Boolean | Yes    |        |         | Yes     |

```cook
V = float(123)       // sample as V = 123.0
V = integer(V)       // cast float number V to integer
V = string(V)        // format V integer into string same as V + ""
V = boolean(V)       // error
V = integer(V)       // parse string V into integer value
V = boolean("true")  // same as V = true
```

## Declare Variable

Declare variable in Cook is very simple, a single equal sign indicate will create a new variable if not exist otherwise it value will be repalce with a new value.

Cook varable can hold any type of value and it can be use to operate one another when type conversion is compatible.

```cook
a = 24.24
VAR = 123
text = "double quote string"
text = 'single quote string'
valid = false
array = [
    1, 2.3, 'abc',
    [33.32, 182],
    {1: "49", 2: "50"}
]
map = { "key": "value" }
```

Result when operate different type together

|         | String | Float  | Integer | Boolean |
|---------|--------|--------|---------|---------|
| String  | String | String | String  | String  |
| Float   | String | Float  | Float   | Error   |
| Integer | String | Float  | Integer | Error   |
| Boolean | String | Error  | Error   | Boolean |

Allowed operator on different type

|         | String | Float      | Integer                             | Boolean  |
|---------|--------|------------|-------------------------------------|----------|
| String  | +      | +          | +                                   | +        |
| Float   | +      | +, -, *, / | +, -, *, /                          |          |
| Integer | +      | +, -, *, / | +, -, *, %, /, <<, >>, &, \|, &^, ^ |          |
| Boolean | +      |            |                                     | &&, \|\| |


For example:

```cook
a = "text a" + " b"   // result string "text a b"
a = "text a " + 123   // result string "text a 123"
a = "text a " + 3.3   // result string "text a 3.3"
a = "text a " + true  // result string "text a true"
a = 2.3 + 13          // result a float or double 15.3
a = 34 + 24.2         // result a float or double 58.2
a = 1 + 2             // result an integer 3
a = true && false     // result a boolean false
a = true + 1          // syntax error
```

## Operator
| Symbol    | Description                  |
| --------- | ---------------------------- |
| **\+**    | add or concatination         |
| **\-**    | substract                    |
| **\***    | multiply                     |
| **%**     | modulo or remaining          |
| **/**     | devide                       |
| **&**     | AND binary operator          |
| **\|**    | OR binary operator           |
| **^**     | XOR binary operator          |
| **&^**    | AND XOR binary operator      |
| **<<**    | Shift Left binary operator   |
| **>>**    | Shift Right binary operator  |
| **>**     | Greater operator             |
| **<**     | Smaller operator             |
| **<=,≤**  | Smaller or equal operator    |
| **>=,≥**  | Greater or equal operator    |
| **==**    | Equal operator               |
| **!=,≠**  | Different operator           |

The symbol below interprete differently in a special case for call expression. Similar to linux redirect syntax.

| Symbol    | Description                                                        |
| --------- | ------------------------------------------------------------------ |
| **<**     | redirect the content of a file as argument                         |
| **>**     | redirect the result of a call to override file content if exist    |
| **>>**    | redirect the result of a call to append to a file content if exist |

For example:

```cook
@get https://www.example.com/data.zip > data.zip
@get https://www.example.com/data.zip > $FILENAME
dataHash = @hash -sha256 < data.zip

// print -e echo or return the the value holding by variable `dataHash`
// with a newline feed and then the result is append to the file `allhash.txt`
@print -e $dataHash >> allhash.txt
```

## Control Flow

Cook favor Go syntax a lot and it build using Go thus Cook have only 1 loop keyword, like Go, `FOR`.

#### Loop as index (1 and 5 are included)
```cook
// below code print 1 to 5 with line feed at the end of each index number
for index in 1..5 {
    @print $index
}
```

#### Loop an array or map
```cook
for index, value in ARRAY {
    // ...
}
for key, value in MAP {
    // ...
}
```

#### Loop infinite
```cook
for {
    // ...
    break
    continue
}
```

#### Nested Loop
```cook
for@label1 index in 1..100 {
    for key, val in MAP {
        if val > 40 {
            // break out of current loop and continue immediately
            // continue the outer loop
            continue label1
        } else if val > 20 {
            // break out of current and outer loop
            break label1
        } else if val > 10 {
            // break or continue with labal working against
            // current loop only
            break
        }
        @print $key
    }
    @print success
}
```

#### If Else Statement

```cook
if VAR {
    // ...
} else if V1 > 2 {
    // ...
} else {
    // ...
}

// short if or ternary expression
A = B > 2 ? 1 : 2
```

#### Fallback Expression

```cook
// if B fail due to evaluation error or B not existed or nil
// then value C is assign A. C must be a non nil expression
// otherwise it will still cause cook to terminated
A = B ?? C

// if ARRAY index 100 is not exist or operation + 1 failed
// then value 0 is assign to variable A
A = ARRAY[100] + 1 ?? 0

// fallback can nested as well
A = Failure1 ?? Failure2 ?? Failure3 ?? false
```

## Built-In Function

See [here](docs/functions/all.md) all built-in functions.

Use `sizeof` to determine a lenght of a variable.

```cook
a = [1, 2, 3]
b = "abcxyz"
c = sizeof a    // result, 3
d = sizeof b    // result, 6
```

## Target

A target is a group instruction similar to function, the only different different is that target cannot be nested and there is no result return from each target. Getting result value computed by other target should be avoided, a target is mean to produce output directly or inderectly into filesystem. You can however share value between target by using Global variable.

```cook
target:
    //....

target@windows:
    // This target will be used instead of default platform target above
    // when executed on Windows platform.
    // ....

initialize:
    // special target always executed before execute any other target

finalize:
    // special target always executed after all other target finish execution
    // regardless whether the it failed or success
    // finalize is mean for a clean up task thus avoid calling other target
    // that produce for file to clean up.

// special target `all` use to define default behavior execution when
// no target is given when execute Cookfile
// * mean execute all other target by it's defined order
all: *
all: @target
```

## Calling a Target, Function or External Command line

Cook allow accessing to a defined Cook target the away that developer wanted. That's it you can call Cook target at any point in your Cookfile, A target is similar to function or external command line as it also accept argument input.

**Note:** avoid create target that have the same name as predefined function as it will cause the Cook program to execute target rather that a built-in function.

Argument given to target is provided in order of number starting from 1, the zero variable is define the length of argument input. In order to access this varaible a `$` symbol is required

```cook
target:
    // $0 return the lenght of the argument input
    if $0 > 1 {
        // access first argument
        @print $1
        for i in 1..$0 {
            // access argument by index
            @print $[$i]
        }
    } else {
        // no input provided, is treated as 3 string arguments
        @print no input provided
    }

main:
    a = 43

    // double quote allow variable interpolate string
    @target 1.23 "Text with $a or ${a} space"

    // single quote treat everything as string
    @target 1.23 'Text with $a or ${a} space'

    // interpolate string work on double quote string or none quoted string
    // use trailing slash (\) for write multiple argument on multiple line
    @target 1.23 long${a}text \
                 -sample \
                 -sample

    // calling a target without arguments
    @target

    // sample as `@print 123` except it call external command echo instead
    #echo 123
```

Any value given as call argument is treat as a string type, You don't need to quote a string however if a string contain whitespace then you should wrap the string within single or double quote.


## Cook Arguments

When execute Cookfile you can also provide addition global variable by defined them in command argument.

The variable argument syntax is similar to Go flag syntax. Each variable name must start with either a `-` or `--` follow by the name of variable and a space or equal sign and finally the value of it.
If multiple name was given then the variable provide in global scope as an array.

```cook
// `var1` is an array `["simple text", 34]`
// `var2` is a float `123.34`
// `ignore` is a boolean type, value `true`
// cook will execute 2 target name `target` and `main`
cook -var1 "simple text" -var2:f=123.34 --var1:i=34 -ignore target main

// custom Cookfile name, default to Cookfile
cook -c myCookfile

// No variable, no explicit target, no custom Cookfile name
// Cookfile is file to be executed
// all target must be defined in `Cookfile`
cook
```

**Note**: In the argument given to cook command line, a suffix `i` and `f` at the end of variable name `var1` is indicated that Cook should convert it value into integer and float value rather string. If no suffix then all value is treat as string.