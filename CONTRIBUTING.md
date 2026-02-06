# webc Contribution Guidelines
Thank you for considering contributing to webc! The contents of this document are guidelines for contributing to this repository.

## Prerequisites
You need to have Go version `1.25` or later installed (I have not tested with previous version, let me know if previous versions do work)

### Recommendations
I **HIGHLY RECOMMEND** having GNU `make` installed; there should be versions for most major Operating Systems

I also recommend using [The WebAssembly Binary Toolkit](https://github.com/WebAssembly/wabt) to debug WASM binaries, not necessary but they will save you countless hours

## Code Style
All conventions should be based on Go's standard style conventions which you can read [here](https://google.github.io/styleguide/go/guide)

**tl;dr:**
- Structs, Functions, Interfaces, Types, Variables, and Constants are `CamelCase`
- Private members of structs or modules are `lowerCamelCase`

### Formatting
I use `gofmt` to format webc, it comes with `go` so there is no need to install extra tools

For any files in `templates/` (HTML and JS files), use `prettier` to format them (`npm install -g prettier`).

`make format` should do all the formatting for you, you just need to make sure the formatters are instaled.

### File Organization
`main.go` contains all functionality related to the compiler executable, things like compiler flags should be put into here

`internal/lexer` is code for the lexer

`internal/parser` is code for the parser

`internal/codegen` is code for the WASM code generator

`internal/ast` contains structs that represent nodes in the Abstract Syntax Tree

`templates` contains templates for the HTML/JS wrapper code that gets created with the compiler

Unit tests should be included in a file inside of it's respective package. The file must end with `_test.go` otherwise `go test` will not detect it

Integration tests (testing pre-built binaries) should be placed inside of the `testing/` folder

## Makefile Actions
### `make test`
Runs all unit/integration tests

### `make format`
Runs the formatters, I recommend having this run as a pre-commit hook

### `make build`
Builds a binary for your current platform

### `make build-all`
Build a binary for every supported platform/architecture

### `make build-target OS=os ARCH=arch`
Builds a binary for a specific platform, thanks to Go's cross compilation capabilities

Below is a table of the commands for every supported OS/Architecture:

| OS      | Architecture          | Command                                 |
|---------|-----------------------|-----------------------------------------|
| Windows | x86_64                | make build-target OS=windows ARCH=amd64 |
| Linux   | x86_64                | make build-target OS=linux ARCH=amd64   |
| Linux   | arm64                 | make build-target OS=linux ARCH=arm64   |
| MacOS   | amd64                 | make build-target OS=darwin ARCH=amd64  |
| MacOS   | arm64 (Apple Silicon) | make build-target OS=darwin ARCH=arm64  |

## Pull Request Guidelines
- Make sure Pull Requests have a focused scope, focus one on feature or bug fix.
- Write descriptive commit messages
- All new features must have tests
- Ensure that all CI checks pass

## Bug Reports/Feature Requests
I have included templates for these kinds of issues, when you create a new issue you should be able to select a template