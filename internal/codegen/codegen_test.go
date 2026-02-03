package codegen_test

import (
	"testing"

	"github.com/grqphical/webc/internal/codegen"
	"github.com/stretchr/testify/assert"
)

func TestEncodingFunctions(t *testing.T) {
	unsignedIntegerTest := 643892
	unsignedIntegerExpectedOutput := []byte{0xB4, 0xA6, 0x27}

	assert.Equal(t, unsignedIntegerExpectedOutput, codegen.EncodeULEB128(uint32(unsignedIntegerTest)))

	signedIntegerTest := -643892
	signedIntegerExpectedOutput := []byte{0xCC, 0xD9, 0x58}

	assert.Equal(t, signedIntegerExpectedOutput, codegen.EncodeSLEB128(int32(signedIntegerTest)))

	floatTest := -3.14
	floatExpectedOutput := []byte{0xC3, 0xF5, 0x48, 0xc0}

	assert.Equal(t, floatExpectedOutput, codegen.EncodeF32(float32(floatTest)))
}
