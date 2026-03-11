package testing

import (
	"testing"

	"github.com/grqphical/webc/internal/codegen"
	"github.com/grqphical/webc/internal/lexer"
	"github.com/grqphical/webc/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestWhileLoop(t *testing.T) {
	source := `int main() {
		int x = 0;

		while (x < 5) {
			x += 1;
		}

		return x;
	}`

	l := lexer.New(source)

	p := parser.New(l)
	program := p.ParseProgram()
	assert.Empty(t, p.Errors())

	module := codegen.NewModule(program)
	err := module.Generate()
	assert.NoError(t, err)
	err = module.Save("temp/while_loop.wasm")
	assert.NoError(t, err)

	AssertWASMBinary(t, "temp/while_loop.wasm", "main", 5, ExpectedOutputI32)
}
