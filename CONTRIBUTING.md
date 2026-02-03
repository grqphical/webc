# webc Contribution Guidelines
Thank you for considering contributing to webc! The contents of this document are guidelines for contributing to this repository.

## Prerequisites
You need to have Go version `1.25` or later installed (I have not tested with previous version, let me know if previous versions do work)

I **HIGHLY RECOMMEND** having GNU `make` installed; there should be versions for most major Operating Systems

### Recommendations
I also recommend using [The WebAssembly Binary Toolkit](https://github.com/WebAssembly/wabt) to debug WASM binaries, not necessary but will save you countless hours

## Code Style
All conventions should be based on Go's standard style conventions which you can read [here](https://google.github.io/styleguide/go/guide)

**tl;dr:**
- Structs, Functions, Interfaces, Types, Variables, and Constants are `CamelCase`
- Private members of structs or modules are `lowerCamelCase`

### Formatting
I use `gofmt` to format webc, it comes with `go` so there is no need to install extra tools

For any files in `templates/` (HTML and JS files), use `prettier` to format them (`npm install -g prettier`).

`make test` should do all the formatting for you, you just need to make sure the formatters are instaled.

### File Organization
`main.go` contains all functionality related to the compiler executable, things like compiler flags should be put into here

`internal/lexer` is code for the lexer

`internal/parser` is code for the parser

`internal/codegen` is code for the WASM code generator

`templates` contains templates for the HTML/JS wrapper code that gets created with the compiler

Unit tests should be included in a file inside of it's respective package. The file must end with `_test.go` otherwise `go test` will not detect it

Integration tests (testing pre-built binaries) should be placed inside of the `testing/` folder

## Bug Reports/Feature Requests
I have included templates for these kinds of issues, when you create a new issue you should be able to select a template