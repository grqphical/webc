package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/grqphical/webc/internal/parser"
)

const magicNumberAndVersion = "\x00asm\x01\x00\x00\x00"

const (
	SecType     byte = 1
	SecFunction byte = 3
	SecExport   byte = 7
	SecCode     byte = 10
)

const (
	OpCodeEnd               byte = 0x0B
	OpCodeReturn            byte = 0x0F
	OpCodeI32Const          byte = 0x41
	OpCodeLocalGet          byte = 0x20
	OpCodeLocalSet          byte = 0x21
	OpCodeI32Add            byte = 0x6A
	OpCodeI32Sub            byte = 0x6B
	OpCodeI32Mul            byte = 0x6C
	OpCodeI32SignedDivision byte = 0x6D
)

// Helper: Encodes unsigned integers (size, counts, indices)
func encodeULEB128(n uint32) []byte {
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

// Helper: Encodes signed integers (constants)
func encodeSLEB128(n int32) []byte {
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
	program parser.Program
}

func NewModule(program parser.Program) *WASMModule {
	m := &WASMModule{
		program: program,
	}
	m.buffer.Write([]byte(magicNumberAndVersion))
	return m
}

func (m *WASMModule) writeSection(id byte, payload []byte) {
	m.buffer.WriteByte(id)
	m.buffer.Write(encodeULEB128(uint32(len(payload))))
	m.buffer.Write(payload)
}

// 1. Type Section
// Defines function signatures. For now, we assume all functions are () -> i32
func (m *WASMModule) generateTypeSection() {
	typePayload := bytes.Buffer{}

	// We only define ONE signature type: () -> i32
	// Even if we have 10 functions, if they all share this signature, we only need one type def.
	typePayload.Write(encodeULEB128(1)) // count of types

	typePayload.WriteByte(0x60)         // function type
	typePayload.Write(encodeULEB128(0)) // Param count: 0
	typePayload.Write(encodeULEB128(1)) // Result count: 1
	typePayload.WriteByte(0x7F)         // result type: i32

	m.writeSection(SecType, typePayload.Bytes())
}

// 2. Function Section
// Maps every function body (in Code section) to a Type signature (in Type section)
func (m *WASMModule) generateFunctionSection() {
	funcPayload := bytes.Buffer{}
	count := uint32(len(m.program.Functions))

	funcPayload.Write(encodeULEB128(count))

	// For every function in our program, assign it Type Index 0 (which is () -> i32)
	for range m.program.Functions {
		funcPayload.Write(encodeULEB128(0))
	}

	m.writeSection(SecFunction, funcPayload.Bytes())
}

// 3. Export Section
// Exports the function named "main" so the host (JS) can call it
func (m *WASMModule) generateExportSection() {
	exportPayload := bytes.Buffer{}
	exportPayload.Write(encodeULEB128(1)) // Number of exports

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

	exportPayload.Write(encodeULEB128(4))                 // Name length
	exportPayload.WriteString("main")                     // Name
	exportPayload.WriteByte(0x00)                         // Export kind: Function
	exportPayload.Write(encodeULEB128(uint32(mainIndex))) // Function Index

	m.writeSection(SecExport, exportPayload.Bytes())
}

// 4. Code Section
// The actual compiled machine code
func (m *WASMModule) generateCodeSection() error {
	codePayload := bytes.Buffer{}
	codePayload.Write(encodeULEB128(uint32(len(m.program.Functions))))

	for _, function := range m.program.Functions {
		funcBody := bytes.Buffer{}

		// --- Local Variable Declarations ---
		localCount := uint32(len(function.SymbolTable))
		if localCount > 0 {
			funcBody.Write(encodeULEB128(1))          // 1 group of locals
			funcBody.Write(encodeULEB128(localCount)) // count
			funcBody.WriteByte(0x7F)                  // type i32
		} else {
			funcBody.Write(encodeULEB128(0)) // 0 groups of locals
		}

		// --- Instructions ---
		for _, stmt := range function.Body.Statements {
			if err := m.generateStatement(stmt, &funcBody); err != nil {
				return err
			}
		}

		funcBody.WriteByte(OpCodeEnd)

		// Write size of this function body + content to payload
		codePayload.Write(encodeULEB128(uint32(funcBody.Len())))
		codePayload.Write(funcBody.Bytes())
	}

	m.writeSection(SecCode, codePayload.Bytes())
	return nil
}

func (m *WASMModule) generateStatement(stmt parser.Node, body *bytes.Buffer) error {
	switch s := stmt.(type) {
	case parser.ReturnStmt:
		m.generateExpressionCode(s.Value, body)
		body.WriteByte(OpCodeReturn)
		return nil

	case parser.VariableDefineStmt:
		m.generateExpressionCode(s.Value, body)
		body.WriteByte(OpCodeLocalSet)
		body.Write(encodeULEB128(uint32(s.Symbol.Index)))
		return nil

	default:
		return errors.New("unsupported statement type")
	}
}

// Recursively generates code for expressions (Post-Order Traversal)
func (m *WASMModule) generateExpressionCode(value parser.Node, body *bytes.Buffer) {
	switch t := value.(type) {

	case parser.Constant:
		body.WriteByte(OpCodeI32Const)
		intValue, _ := strconv.Atoi(t.Value)
		body.Write(encodeSLEB128(int32(intValue)))

	case parser.VariableAccess:
		body.WriteByte(OpCodeLocalGet)
		body.Write(encodeULEB128(uint32(t.Index)))

	case parser.UnaryExpression:
		// WASM doesn't have a "negate" opcode for i32, so we do (0 - value)
		body.WriteByte(OpCodeI32Const)
		body.Write(encodeSLEB128(0))
		m.generateExpressionCode(t.Value, body)
		body.WriteByte(OpCodeI32Sub)

	case parser.BinaryExpression:
		m.generateExpressionCode(t.A, body)
		m.generateExpressionCode(t.B, body)

		switch t.Operation {
		case "+":
			body.WriteByte(OpCodeI32Add)
		case "-":
			body.WriteByte(OpCodeI32Sub)
		case "*":
			body.WriteByte(OpCodeI32Mul)
		case "/":
			body.WriteByte(OpCodeI32SignedDivision)
		}
	}
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
