package testing

import (
	"testing"

	"github.com/grqphical/webc/internal/codegen"
	"github.com/grqphical/webc/internal/lexer"
	"github.com/grqphical/webc/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestForLoop(t *testing.T) {
	source := `int main() {
		int x = 0;
		for (int i = 0; i < 5; i++) {
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
	err = module.Save("temp/for_loop.wasm")
	assert.NoError(t, err)

	AssertWASMBinary(t, "temp/for_loop.wasm", "main", 5, ExpectedOutputI32)
}
