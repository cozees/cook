# Cook Language Specification

This document regarding cook language specification.

# Operand Value Syntax

```cook
123
true
false
1.43
"string"
'string'
`string`
[1, 2, 3]               // array
{1: 2, 3: 'text'}       // map or dictionary
['*.txt']               // glob pattern. result an array of file path
```

# Declare Variable Syntax

```cook
A = 123
A = true
A = [123, 1.23, "xyz"]
A = {1.2:'abc', true: 4.2}
A = `test/file.txt`
B = 1 + 2
C = A * 3
D = ['folder/*.txt']
E = ['folder/**/test/*.txt']
F = {'folder/**/test/*.txt'}
```

# Built-in types

```cook
123                         // integer value
1.23                        // float value
true/false                  // boolean value
'text', "text", `text`      // string value
[1,2,3]                     // array value
{1:2, 3:'abc'}              // map or dictionary value
```

In cook, there are two ways to define multiple line string. It can be done by simply wrap multiple
line string in a backquote (`) or written them in line. The backquote string will preserve everything except string interpolation while the simple multiple line string does not perserve everyting, it elimiate any leading and trailing whitespace.

```cook
// Multiline string below is equal to "\n\ttext line1\n\ttext line2\n"
A = `
	text line 1
	text line 2
`

// Multiline string belew is equal to "CREATE TABLE A (\n\tid int\n)"
A = "CREATE TABLE A ("
    "	id int"
    ")"
```

String interpolation can declare using dollar sign within the string. Unlike other language, Cook allow string interpolation on any form of string declaration, so to escape it you've to use string escape \ backslash before the dollar sign.

```cook
// B result with "Simple text 123 with escape $A"
A = 123
B = "Simple text $A with escape \$A"

// B result with "Simple text 124 with escape ${A}"
B = "Simple text ${A+1} with escape \${A}"
```

### Compatible conversion between buit-in type
|         | String | Float  | Integer | Boolean |
|---------|--------|--------|---------|---------|
| String  | Yes    | Yes    | Yes     | Yes     |
| Float   | Yes    | Yes    | Yes     |         |
| Integer | Yes    | Yes    | Yes     |         |
| Boolean | Yes    |        |         | Yes     |

You can also casting value explicitly:

```cook
A = 123
B = float(A)            // equal to B = 123.0
A = string(123)         // equal to A = "123"
A = string(true)        // equal to A = "true"
A = boolean("true")     // equal to A = true
A = integer("123")      // equal to A = 123
```

## String

Cook string is utilize Go string directly therefore strings operation has similar effect to Go.

```cook
A = "sample text"

A += " extra text"      // A is now "sample text extra text"
B = A{1..3}             // substring A to B is "amp"

```

## Array

An array is a dynamic list that can contain any value and unlimited size.

```cook
A = [1, 2.34, myfile.txt]

// append an item to array, result [1, 2.34, myfile.txt, 3]
A += 3

// append 3 element to array, result [1, 2.34, myfile.txt, 3, 4, 2, a]
// a is variable, it value is determine during evaluation.
A += [4, 2, a]

// append a map to array, result [1, 2.34, myfile.txt, 3, 4, 2, a, {1:2, 3:true}]
M = {1:2, 3:true}
A += M

// remove one or more item from array, result [2.34, myfile.txt]
A -= M
A -= [1, 3, 4, 2, a]

// insert one or multiple item into array at position 0, result [1, 2, 11, 2.34, myfile.txt]
// it's easy to make mistake with syntax "A[0]" which mean modifer value of array index 0
A{0} += 11
A{0} += [1, 2]

// remove one or multiple item from array
delete A{0}               // removing 1 item at position 0
delete A{1..3}            // removing 2 item at position 1, 2, 3
delete A{1,2,3}           // removing 2 item at position 1, 2, 3

// slice array, since we built using Golang, we also take advantage of using Go slice too
// A current value is [2, myfile.txt]
A += [5, 3, 6]      // A is [2, myfile.txt, 5, 3, 6]
B = A{2..4}         // B is [5, 3, 6]

// accessing or modifying array value
A[0] = 2
A[0] += 1
```

## Map

As Cook build using Go, the map in Cook is behavor just like map in Go. The Map item is not
ordered so be mind full using it.

```cook
A = {1:2, 2:"text"}

// merge new map into exist map A. + allowed merge new map (right) to existing map (left)
// with no conflict key on both map.
A += {3:"21", "a": 8.2}         // {1:2, 2:"text", 3:"21", "a":8.2}

// Use "<" to resolve conflict in favor or a new map or "?" to resolve conflict in favor
// of the existing map.
A += < {2:"21"}                 // {1:2, 2:"21", 3:"21", "a":8.2}
A += ? {1:true}                 // {1:2, 2:"21", 3:"21", "a":8.2}

// like array, to remove an item from map we can also use similar syntax except that
// map does allow or support with range.
delete A{2}                     // {1:2, 3:"21", "a":8.2}
delete A{"a"}                   // {1:2, 3:"21"}
```

## Transform array or map element

Transform 

```cook
A = [1, 2, 3]

// B is an array immultabl, its value is based on value of array A
B = A(i, v) => (i + 1) * v

// Cook raise an error if we try to modify an immultable array
B[1] = 1

// print out "4 9"
@print B[1] B[2]                    

// Change source of transform array effect all sub transform array
A[0] = 5
@print B                            // print "5, 4, 9"

// Even update array A in place, the transform array B values is still
// base on array A.
A = A(i, v) => (i + 1) * v          // A is now [5, 4, 9]
@print B                            // print "5, 8, 27"

// Transformation can be applied map or dictionary as well
A = {1:2, 2:4}
B = A(k, v) => v * k                // B is {1:2, 2:8}
@print B[1] B[2] A[1] A[2]          // print "2 8 2 4"
A[1] = 4
@print B[1] A[1]                    // print "4 4"

// complex transformation can be wrapped in a block function literal
B = A(key, value) => {
    if $key == 1 {
        return $value * 4
    } else {
        return $value * ($key + 2)
    }
}
```

# Operator

| Operation      | Symbol Operator           |
| ------------   | ------------------------- |
| additive       | **+**, **-**              |
| bit            | &, \|, >>, <<, ^, &^      |
| multiplicative | *, /, %                   |
| comparative    | ==, !=, <, >, <=, >=      |
| logic          | &&, ||, is, is!           |
| ternary        | ?, ??                     |
| assignment     | =, +=, -=, ...etc         |
| unary          | -expr, +expr, !expr       |

Special exception case when writting call argument.

| Operation               | Symbol Operator           |
| ----------------------- | ------------------------- |
| Create/Overrite to file | >                         |
| Create/Append to file   | >>                        |
| Read from file          | <                         |

# Function

Cook function is similar other language except it does not required explicitly return type or argument type.

For built-in function document visit [here](docs/functions/all.md). Our built-in function is defined
in the form of Command call there you can literally print out function description or documentation.

```cook
// function lamda syntax
sample(a, b) => a + b

// block syntax
sample(a, b) {
    return a + b
}
```

# Target

A target is similar to a function exception is does not allow explicit argument declaration and it also forbid from return any value. However you can still call and pass argument to target the same way that you pass argument to a function. To access argument in target, use dollar sign "$" follow by number of index variable which pass to. The argument "$0" represent total number of argument pass to the target.

```cook
target:
    A = 123 * $2 + $0
```

# Control Flow

## If Else statement

Like most of language, Cook also provide an if and else statement.

```cook
if a > 0 {
    // block execution
} else if 0 < b < 10 {
    // short way to write condition check like below
    // block execution
} else if 0 < c && c < 10 {
    // block execution
} else {
    // block execution
}
```

Using Ternary syntax for short if else. Although, Cook support nested ternary you avoid it if possible.

```cook
A = expression ? true : false
```

Using fallback syntax for default value. Note Cook allow expression syntax on default value as well
thus it is up to developer to make sure that default value is absolutely return a value and not an
error of any kind. Fallback syntax occurred when the primary value or expression match condition below:

- Variable not existed
- Index out of range or index not existed
- There is an error when evaluate expression/primary value

The below example express that if the evaluation of "expression" raise error that match above description then the variable A is set to default value integer 0.

```cook
A = expression ?? 0
```

## For loop

```cook
// Looping using range syntax.
// Variable i can be modified in the loop to purposely increasing step.
// By default, range loop syntax has a default step where it increase or decrease by 1
for i in 1..10 {
    // block execution
    if i > 5 {
        // suppose i value is 6, next execution i value is 9 rather than 7
        i += 2
    }
}

// using math interval syntax to exclude value in range. Below index i is loop
// starting from 0 until 2 rather 3.
// The loop step is increase by 2 rather than the default value 1
A = [1, 2, 3]
for i in [0..sizeof A):2 {
    // block execution
}

// loop in item in array
for index, value in [1, 2, 3] {
    // block execution
    // it's safe to remove item from array
}

// loop in item in map
for index, value in {1:2, 3:3} {
    // block execution
    // it's safe to remove item from map
}

// infinite loop
a = 1
for {
    if a > 3 {
        break
    }
    a++
}
```







