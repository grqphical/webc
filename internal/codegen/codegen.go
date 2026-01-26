package codegen

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
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
	OpCodeEnd    byte = 0x0B
	OpCodeReturn byte = 0x0F

	OpCodeLocalGet byte = 0x20
	OpCodeLocalSet byte = 0x21

	OpCodeI32Const          byte = 0x41
	OpCodeI32Add            byte = 0x6A
	OpCodeI32Sub            byte = 0x6B
	OpCodeI32Mul            byte = 0x6C
	OpCodeI32SignedDivision byte = 0x6D

	OpCodeF32Const    byte = 0x43
	OpCodeF32Neg      byte = 0x8C
	OpCodeF32Add      byte = 0x92
	OpCodeF32Sub      byte = 0x93
	OpCodeF32Mul      byte = 0x94
	OpCodeF32Division byte = 0x95
)

// encodeF32 converts a float32 to its 4-byte little-endian representation
func encodeF32(f float32) []byte {
	bits := math.Float32bits(f)

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, bits)

	return buf
}

// Encodes unsigned integers (size, counts, indices)
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

// Encodes signed integers (constants)
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

func (m *WASMModule) generateTypeSection() {
	typePayload := bytes.Buffer{}
	typePayload.Write(encodeULEB128(uint32(len(m.program.Functions)))) // count of types

	for _, f := range m.program.Functions {

		typePayload.WriteByte(0x60)         // function type
		typePayload.Write(encodeULEB128(0)) // Param count: 0
		typePayload.Write(encodeULEB128(1)) // Result count: 1
		switch f.Type {
		case parser.TypeInt:
			typePayload.WriteByte(0x7F)
		case parser.TypeFloat:
			typePayload.WriteByte(0x7D)
		}
	}
	m.writeSection(SecType, typePayload.Bytes())
}

func (m *WASMModule) generateFunctionSection() {
	funcPayload := bytes.Buffer{}
	count := uint32(len(m.program.Functions))

	funcPayload.Write(encodeULEB128(count))

	for range m.program.Functions {
		funcPayload.Write(encodeULEB128(0))
	}

	m.writeSection(SecFunction, funcPayload.Bytes())
}

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

func (m *WASMModule) generateExpressionCode(value parser.Node, body *bytes.Buffer) {
	switch t := value.(type) {

	case parser.Constant:
		if t.Type == parser.TypeInt {
			body.WriteByte(OpCodeI32Const)
			intValue, _ := strconv.Atoi(t.Value)
			body.Write(encodeSLEB128(int32(intValue)))
		} else if t.Type == parser.TypeFloat {
			body.WriteByte(OpCodeF32Const)
			floatValue, _ := strconv.ParseFloat(t.Value, 32)
			body.Write(encodeF32(float32(floatValue)))
		}

	case parser.VariableAccess:
		body.WriteByte(OpCodeLocalGet)
		body.Write(encodeULEB128(uint32(t.Index)))

	case parser.UnaryExpression:
		m.generateExpressionCode(t.Value, body)

		// Check the type of the value being negated
		if t.Value.GetType() == parser.TypeFloat {
			// f32 has a direct 'neg' opcode
			body.WriteByte(OpCodeF32Neg) // f32.neg
		} else {
			// i32 doesn't have neg, so we do (0 - x)
			body.WriteByte(OpCodeI32Const)
			body.Write(encodeSLEB128(0))
			m.generateExpressionCode(t.Value, body) // Move this here for i32
			body.WriteByte(OpCodeI32Sub)
		}

	case parser.BinaryExpression:
		m.generateExpressionCode(t.A, body)
		m.generateExpressionCode(t.B, body)

		// Branching based on the type of the expression
		isFloat := t.A.GetType() == parser.TypeFloat || t.B.GetType() == parser.TypeFloat

		switch t.Operation {
		case "+":
			if isFloat {
				body.WriteByte(OpCodeF32Add)
			} else {
				body.WriteByte(OpCodeI32Add)
			}
		case "-":
			if isFloat {
				body.WriteByte(OpCodeF32Sub)
			} else {
				body.WriteByte(OpCodeI32Sub)
			}
		case "*":
			if isFloat {
				body.WriteByte(OpCodeF32Mul)
			} else {
				body.WriteByte(OpCodeI32Mul)
			}
		case "/":
			if isFloat {
				body.WriteByte(OpCodeF32Division)
			} else {
				body.WriteByte(OpCodeI32SignedDivision)
			}
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
