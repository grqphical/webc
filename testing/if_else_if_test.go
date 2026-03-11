package testing

import (
	"testing"

	"github.com/grqphical/webc/internal/codegen"
	"github.com/grqphical/webc/internal/lexer"
	"github.com/grqphical/webc/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestIfElseIfStatements(t *testing.T) {
	source := `int main() {
		int x = 0;

		if (5 < 4) {
			x = 6;
		} else if (4 < 5) {
			x = 5; 
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
	err = module.Save("temp/if_else_statement.wasm")
	assert.NoError(t, err)

	AssertWASMBinary(t, "temp/if_else_statement.wasm", "main", 5, ExpectedOutputI32)
}

func TestIfElseIfWithLessOrEqualStatements(t *testing.T) {
	source := `int main() {
		int x = 5;

		if (x <= 5) {
			x = 6;
		} else if (4 < 5) {
			x = 4; 
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
	err = module.Save("temp/if_else_lt_eq_statement.wasm")
	assert.NoError(t, err)

	AssertWASMBinary(t, "temp/if_else_lt_eq_statement.wasm", "main", 6, ExpectedOutputI32)
}
