package testing

import (
	"testing"

	"github.com/grqphical/webc/internal/codegen"
	"github.com/grqphical/webc/internal/lexer"
	"github.com/grqphical/webc/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestLogicalExpressionReturn(t *testing.T) {
	source := `int main() {
		return 10 < 15;
	}`

	l := lexer.New(source)

	p := parser.New(l)
	program := p.ParseProgram()
	assert.Empty(t, p.Errors())

	module := codegen.NewModule(program)
	err := module.Generate()
	assert.NoError(t, err)
	err = module.Save("temp/logical_expression.wasm")
	assert.NoError(t, err)

	AssertWASMBinary(t, "temp/logical_expression.wasm", "main", 1, ExpectedOutputI32)
}

func TestNotExpressionReturn(t *testing.T) {
	source := `int main() {
		return !(10 < 15);
	}`

	l := lexer.New(source)

	p := parser.New(l)
	program := p.ParseProgram()
	assert.Empty(t, p.Errors())

	module := codegen.NewModule(program)
	err := module.Generate()
	assert.NoError(t, err)
	err = module.Save("temp/not_expression.wasm")
	assert.NoError(t, err)

	AssertWASMBinary(t, "temp/not_expression.wasm", "main", 0, ExpectedOutputI32)
}

func TestFloatLogicalExpressionReturn(t *testing.T) {
	source := `int main() {
		return 10.0 < 15.0;
	}`

	l := lexer.New(source)

	p := parser.New(l)
	program := p.ParseProgram()
	assert.Empty(t, p.Errors())

	module := codegen.NewModule(program)
	err := module.Generate()
	assert.NoError(t, err)
	err = module.Save("temp/logical_float_expression.wasm")
	assert.NoError(t, err)

	AssertWASMBinary(t, "temp/logical_float_expression.wasm", "main", 1, ExpectedOutputI32)
}

func TestFloatNotExpressionReturn(t *testing.T) {
	source := `int main() {
		return !(10.0 < 15.0);
	}`

	l := lexer.New(source)

	p := parser.New(l)
	program := p.ParseProgram()
	assert.Empty(t, p.Errors())

	module := codegen.NewModule(program)
	err := module.Generate()
	assert.NoError(t, err)
	err = module.Save("temp/not_float_expression.wasm")
	assert.NoError(t, err)

	AssertWASMBinary(t, "temp/not_float_expression.wasm", "main", 0, ExpectedOutputI32)
}
