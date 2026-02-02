<div style="text-align: center">
    <h1>webc - A lightweight, WASM C99 Compiler</h1>
</div>

[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/grqphical/webc.svg)](https://github.com/grqphical/webc)
[![GitHub license](https://img.shields.io/github/license/grqphical/webc.svg)](https://github.com/grqphical/webc/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/grqphical/webc.svg)](https://GitHub.com/grqphical/webc/releases/)

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
- [ ] Long (64 bit Integer) variables, arthimetic, and variable modification
- [ ] Unsigned Integer/Long variables, arthimetic, and variable modification
- [ ] If/Else-If/Else Statements
- [ ] For/While loops
- [ ] Functions
- [ ] Preprocessor (include, define, ifdef, etc.)
- [ ] Dynamic Memory (malloc/free)
- [ ] Structs, Unions, Typedefs
- [ ] Static variables

## Contributions
Any contribution is welcome, just make sure to run `make format` before you commit to ensure the code style remains consistent

## License
`webc` is licensed under the Apache 2.0 License
