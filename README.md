# webc - A lightweight, WASM C99 Compiler
webc is a compiler for C99 that outputs WASM and a JS runtime for both servers and browsers.

Currently in development

## Features
Currently it only supports creating a `main()` function that returns an integer, variable declarations, and arithmetic

## Usage
Simply run the compiler with a C file and it will output a WASM binary as well as an HTML/JS file to run the program

You can use `-s` to create a server runtime that can be run with Node.JS or a similar JS runtime

You can use the `-o` flag to change where the output binary will be.

## License
`webc` is licensed under the Apache 2.0 License
