package testing

import (
	"testing"

	"github.com/grqphical/webc/internal/codegen"
	"github.com/grqphical/webc/internal/lexer"
	"github.com/grqphical/webc/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestIfStatements(t *testing.T) {
	source := `int main() {
		int x = 0;

		if (5 > 4) {
			x = 6;
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
	err = module.Save("temp/if_statement.wasm")
	assert.NoError(t, err)

	AssertWASMBinary(t, "temp/if_statement.wasm", "main", 6, ExpectedOutputI32)
}
