package testing

import (
	"testing"

	"github.com/grqphical/webc/internal/codegen"
	"github.com/grqphical/webc/internal/lexer"
	"github.com/grqphical/webc/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestWASMIncrementDecrement(t *testing.T) {
	source := `int main() {
		int x = 20;

		x++;
		++x;
		x--;

		return x;

	}`

	l := lexer.New(source)

	p := parser.New(l)
	program := p.ParseProgram()
	assert.Empty(t, p.Errors())

	module := codegen.NewModule(program)
	err := module.Generate()
	assert.NoError(t, err)
	err = module.Save("temp/increment_decrement.wasm")
	assert.NoError(t, err)

	AssertWASMBinary(t, "temp/increment_decrement.wasm", "main", 21, ExpectedOutputI32)
}
