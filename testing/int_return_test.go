package testing

import (
	"log"
	"os"
	"testing"

	"github.com/grqphical/webc/internal/codegen"
	"github.com/grqphical/webc/internal/lexer"
	"github.com/grqphical/webc/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestWASMIntegerReturn(t *testing.T) {
	source := `int main() {
		int x = 10;
		x += 20
		x -= 10
		x *= 3;
		x /= 2
		return x;
	}`

	l := lexer.New(source)
	tokens, err := l.ParseSource()
	assert.NoError(t, err)

	p := parser.New(tokens)
	program, err := p.Parse()
	assert.NoError(t, err)

	module := codegen.NewModule(program)
	err = module.Generate()
	assert.NoError(t, err)
	err = module.Save("temp/integer_return.wasm")
	assert.NoError(t, err)

	AssertWASMBinary(t, "temp/integer_return.wasm", "main", 30)
}

func init() {
	if _, err := os.Stat("./temp/"); err != nil {
		err = os.Mkdir("temp", 0700)
		if err != nil {
			log.Fatal("could not create temp directory")
		}
	}
}
