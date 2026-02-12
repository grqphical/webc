package testing

import (
	"testing"

	"github.com/grqphical/webc/internal/codegen"
	"github.com/grqphical/webc/internal/lexer"
	"github.com/grqphical/webc/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestFunctionCall(t *testing.T) {
	source := `
float timestwo(float x) {
  return x * 2.0;
}

float main()
{
  float x = timestwo(2.0);

  return x;
}
`

	l := lexer.New(source)

	p := parser.New(l)
	program := p.ParseProgram()
	assert.Empty(t, p.Errors())

	module := codegen.NewModule(program)
	err := module.Generate()
	assert.NoError(t, err)
	err = module.Save("temp/float_return.wasm")
	assert.NoError(t, err)

	AssertWASMBinary(t, "temp/float_return.wasm", "main", float32(4.0), "f32")
}
