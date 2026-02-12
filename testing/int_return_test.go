package testing

import (
	"testing"

	"github.com/grqphical/webc/internal/codegen"
	"github.com/grqphical/webc/internal/lexer"
	"github.com/grqphical/webc/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestWASMIntegerReturn(t *testing.T) {
	source := `int main() {
		int x = 10;

		x += 10;

		return x;
	}`

	l := lexer.New(source)

	p := parser.New(l)
	program := p.ParseProgram()
	assert.Empty(t, p.Errors())

	module := codegen.NewModule(program)
	err := module.Generate()
	assert.NoError(t, err)
	err = module.Save("temp/integer_return.wasm")
	assert.NoError(t, err)

	AssertWASMBinary(t, "temp/integer_return.wasm", "main", 20, ExpectedOutputI32)
}
