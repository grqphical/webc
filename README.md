<div style="text-align: center">
    <h1>webc - A lightweight, WASM C99 Compiler</h1>
</div>

[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/grqphical/webc.svg)](https://github.com/grqphical/webc)
[![GitHub license](https://img.shields.io/github/license/grqphical/webc.svg)](https://github.com/grqphical/webc/blob/master/LICENSE)
[![Run Go Tests](https://github.com/grqphical/webc/actions/workflows/tests.yml/badge.svg)](https://github.com/grqphical/webc/actions/workflows/tests.yml)

webc is a compiler for C99 that outputs WASM and a JS runtime for both servers and browsers. The goal of webc is to create a runtime similar to native C that can be run either
in the browser or on the server (with Node.js, Bun etc.) with features such as a virtual filesystem, C style error handling (errno), and support for signals from JavaScript.

Currently in active development

## Installation
The only requirements for webc is having Go installed on your system, other than that you can install it with:
```bash
git clone https://github.com/grqphical/webc
cd webc
make build
```

## Usage
Simply run the compiler with a C file and it will output a WASM binary as well as an HTML/JS file to run the program

You can use `-s` to create a server runtime that can be run with Node.JS or a similar JS runtime

You can use the `-o` flag to change where the output binary will be.

## Features/Roadmap
- [x] Main function
- [x] Integer variables, arthimetic, and variable modification
- [x] Floating point variables, arthimetic, and variable modification
- [x] Character variables, arthimetic, and variable modification
- [x] Constant variables
- [ ] Long (64 bit Integer) variables, arthimetic, and variable modification
- [ ] Unsigned Integer/Long variables, arthimetic, and variable modification
- [ ] If/Else-If/Else Statements
- [ ] For/While loops
- [ ] Functions
- [ ] Preprocessor (include, define, ifdef, etc.)
- [ ] Dynamic Memory (malloc/free)
- [ ] Structs, Unions, Typedefs
- [ ] Static variables

## Development
To build webc, run `make build`

To run webc's tests, run `make test`. webc includes both unit tests and a custom integration test system for testing the output of functions in compiled WASM binaries.
This system is inside of the `testing/` directory

## Contributions
Any contribution is welcome, please make sure to go through the steps below before making a Pull Request:
1. Any new code has tests (if necessary, which 99 times out of 100 it is)
2. Make sure EVERY test passes by running `make test`
3. Make sure the code is formatted by running `make format` (this does require `gofmt` which should come with every Go installation)

## License
`webc` is licensed under the Apache 2.0 License
