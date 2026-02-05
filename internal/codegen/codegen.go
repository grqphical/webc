package codegen

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/grqphical/webc/internal/ast"
)

const magicNumberAndVersion = "\x00asm\x01\x00\x00\x00"

const (
	SecType     byte = 1
	SecFunction byte = 3
	SecExport   byte = 7
	SecCode     byte = 10
)

const (
	OpCodeEnd    byte = 0x0B
	OpCodeReturn byte = 0x0F

	OpCodeLocalGet byte = 0x20
	OpCodeLocalSet byte = 0x21

	OpCodeI32Const          byte = 0x41
	OpCodeI32Add            byte = 0x6A
	OpCodeI32Sub            byte = 0x6B
	OpCodeI32Mul            byte = 0x6C
	OpCodeI32SignedDivision byte = 0x6D
	OpCodeI32And            byte = 0x71

	OpCodeF32Const    byte = 0x43
	OpCodeF32Neg      byte = 0x8C
	OpCodeF32Add      byte = 0x92
	OpCodeF32Sub      byte = 0x93
	OpCodeF32Mul      byte = 0x94
	OpCodeF32Division byte = 0x95
)

// EncodeF32 converts a float32 to its 4-byte little-endian representation
func EncodeF32(f float32) []byte {
	bits := math.Float32bits(f)

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, bits)

	return buf
}

// Encodes unsigned integers to Little Endian Binary 128-bit format
func EncodeULEB128(n uint32) []byte {
	var res []byte
	for {
		b := byte(n & 0x7F)
		n >>= 7
		if n == 0 {
			res = append(res, b)
			break
		}
		res = append(res, b|0x80)
	}
	return res
}

// Encodes signed integers to Little Endian Binary 128-bit format
func EncodeSLEB128(n int32) []byte {
	var res []byte
	for {
		b := byte(n & 0x7F)
		n >>= 7
		if (n == 0 && (b&0x40) == 0) || (n == -1 && (b&0x40) != 0) {
			res = append(res, b)
			break
		}
		res = append(res, b|0x80)
	}
	return res
}

type WASMModule struct {
	buffer  bytes.Buffer
	program *ast.Program
}

func NewModule(program *ast.Program) *WASMModule {
	m := &WASMModule{
		program: program,
	}
	m.buffer.Write([]byte(magicNumberAndVersion))
	return m
}

func (m *WASMModule) writeSection(id byte, payload []byte) {
	m.buffer.WriteByte(id)
	m.buffer.Write(EncodeULEB128(uint32(len(payload))))
	m.buffer.Write(payload)
}

func (m *WASMModule) generateTypeSection() {
	typePayload := bytes.Buffer{}
	typePayload.Write(EncodeULEB128(uint32(len(m.program.Functions)))) // count of types

	for _, f := range m.program.Functions {

		typePayload.WriteByte(0x60)         // function type
		typePayload.Write(EncodeULEB128(0)) // Param count: 0
		typePayload.Write(EncodeULEB128(1)) // Result count: 1
		switch f.ReturnType {
		case ast.ValueTypeInt, ast.ValueTypeChar:
			typePayload.WriteByte(0x7F)
		case ast.ValueTypeFloat:
			typePayload.WriteByte(0x7D)
		}
	}
	m.writeSection(SecType, typePayload.Bytes())
}

func (m *WASMModule) generateFunctionSection() {
	funcPayload := bytes.Buffer{}
	count := uint32(len(m.program.Functions))

	funcPayload.Write(EncodeULEB128(count))

	for range m.program.Functions {
		funcPayload.Write(EncodeULEB128(0))
	}

	m.writeSection(SecFunction, funcPayload.Bytes())
}

func (m *WASMModule) generateExportSection() {
	exportPayload := bytes.Buffer{}
	exportPayload.Write(EncodeULEB128(1)) // Number of exports

	// Find the index of the "main" function
	mainIndex := 0
	foundMain := false
	for i, f := range m.program.Functions {
		if f.Name == "main" {
			mainIndex = i
			foundMain = true
			break
		}
	}

	if !foundMain {
		fmt.Println("Warning: No 'main' function found to export.")
	}

	exportPayload.Write(EncodeULEB128(4))                 // Name length
	exportPayload.WriteString("main")                     // Name
	exportPayload.WriteByte(0x00)                         // Export kind: Function
	exportPayload.Write(EncodeULEB128(uint32(mainIndex))) // Function Index

	m.writeSection(SecExport, exportPayload.Bytes())
}

func (m *WASMModule) generateCodeSection() error {
	codePayload := bytes.Buffer{}
	codePayload.Write(EncodeULEB128(uint32(len(m.program.Functions))))

	for _, function := range m.program.Functions {
		funcBody := bytes.Buffer{}

		intCount, floatCount := function.GetVariableCounts()
		numGroups := uint32(0)
		if intCount > 0 {
			numGroups++
		}
		if floatCount > 0 {
			numGroups++
		}

		funcBody.Write(EncodeULEB128(numGroups))
		if intCount > 0 {
			funcBody.Write(EncodeULEB128(uint32(intCount)))
			funcBody.WriteByte(0x7F) // i32
		}
		if floatCount > 0 {
			funcBody.Write(EncodeULEB128(uint32(floatCount)))
			funcBody.WriteByte(0x7D) // f32
		}

		// --- Instructions ---
		for _, stmt := range function.Body.Statements {
			if err := m.generateStatement(stmt, &funcBody); err != nil {
				return err
			}
		}

		funcBody.WriteByte(OpCodeEnd)

		codePayload.Write(EncodeULEB128(uint32(funcBody.Len())))
		codePayload.Write(funcBody.Bytes())
	}

	m.writeSection(SecCode, codePayload.Bytes())
	return nil
}

func (m *WASMModule) Generate() error {
	m.generateTypeSection()
	m.generateFunctionSection()
	m.generateExportSection()
	return m.generateCodeSection()
}

func (m *WASMModule) Save(filename string) error {
	return os.WriteFile(filename, m.buffer.Bytes(), 0644)
}
