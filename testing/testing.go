package testing

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type ExpectedOutput string

const (
	ExpectedOutputI32 ExpectedOutput = "i32"
	ExpectedOutputF32 ExpectedOutput = "f32"
)

// Reads a WASM binary, calls a function from it and check it's return value against the provided expected output.
// Under the hood, it uses the testify/assert library to assert the equality check
//
// Should only be used in a testing environment (ie in a *_test.go file, ran under `go test`)
func AssertWASMBinary(t *testing.T, binaryFile string, functionName string, expectedOutput any, expectedOutputType ExpectedOutput) {
	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)

	wasmBytes, err := os.ReadFile(binaryFile)
	if err != nil {
		t.Fatalf("failed to read wasm binary: %v", err)
	}

	mod, err := runtime.InstantiateWithConfig(context.Background(), wasmBytes, wazero.NewModuleConfig())
	if err != nil {
		t.Fatalf("failed to instantiate wasm module: %v", err)
	}
	defer mod.Close(ctx)

	mainFunc := mod.ExportedFunction(functionName)
	if mainFunc == nil {
		t.Fatalf("main function not found in wasm module")
	}

	results, err := mainFunc.Call(ctx)
	if err != nil {
		t.Fatalf("failed to call main: %v", err)
	}

	var output uint64
	if len(results) == 1 {
		output = results[0]
	}

	switch expectedOutputType {
	case "i32":
		o := api.DecodeI32(output)
		assert.EqualValues(t, expectedOutput, o)
	case "f32":
		o := api.DecodeF32(output)
		assert.EqualValues(t, expectedOutput, o)
	}

}

func init() {
	if _, err := os.Stat("./temp/"); err != nil {
		err = os.Mkdir("temp", 0700)
		if err != nil {
			log.Fatal("could not create temp directory")
		}
	}
}
