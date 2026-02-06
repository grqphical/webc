package testing

import (
	"testing"

	"github.com/grqphical/webc/internal/codegen"
	"github.com/grqphical/webc/internal/lexer"
	"github.com/grqphical/webc/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestWASMCharReturn(t *testing.T) {
	source := `char main() {
		char x = 'a';

		x += 4;

		return x;
	}`

	l := lexer.New(source)

	p := parser.New(l)
	program := p.ParseProgram()

	module := codegen.NewModule(program)
	err := module.Generate()
	assert.NoError(t, err)
	err = module.Save("temp/integer_return.wasm")
	assert.NoError(t, err)

	AssertWASMBinary(t, "temp/integer_return.wasm", "main", 101, "i32")
}
