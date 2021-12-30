# Cook Introduction

![icon](docs/icon.mark.png)

A simple interpreter language to read and execute cook statement/instruction in Cookfile. 
Cookfile syntax was inspire by Go and the tools gnu make. Cook aiming to provide cross-platform
compatibility and simplicity.

Although all Cook functionality is being test on Linux, MacOS and Windows, it is currently at it early stage.

# Languages

More about Cook, check language [specification](spec/language.md). For built-in function visit [here](docs/functions/all.md)

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

Create a file name "Cookfile" with content below

```cook
A = 12

target:
    @print A
```

Then to execute above code run

```bash
cook target

// or

cook -c Cookfile target
```
