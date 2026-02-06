package codegen

import (
	"bytes"
	"encoding/binary"
	"errors"
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
	LocalTypeI32 byte = 0x7F
	LocalTypeF32 byte = 0x7D
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

func checkCompatibleTypes(left, right ast.ValueType) bool {
	if left == right {
		return true
	}

	switch left {
	case ast.ValueTypeInt:
		return right == ast.ValueTypeChar
	case ast.ValueTypeChar:
		return right == ast.ValueTypeInt
	default:
		return false
	}
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

		funcBody.Write(EncodeULEB128(uint32(len(function.Symbols))))

		for _, sym := range function.Symbols {
			funcBody.Write(EncodeULEB128(1))

			switch sym.Type {
			case ast.ValueTypeInt, ast.ValueTypeChar:
				funcBody.WriteByte(LocalTypeI32)
			case ast.ValueTypeFloat:
				funcBody.WriteByte(LocalTypeF32)
			}
		}

		// --- Instructions ---
		for _, stmt := range function.Statements {
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

func (m *WASMModule) generateExpressionCode(exp ast.Expression, funcBody *bytes.Buffer) error {
	switch e := exp.(type) {
	case *ast.IntegerLiteral:
		funcBody.WriteByte(OpCodeI32Const)
		funcBody.Write(EncodeSLEB128(int32(e.Value)))
		return nil
	case *ast.FloatLiteral:
		funcBody.WriteByte(OpCodeF32Const)
		funcBody.Write(EncodeF32(float32(e.Value)))
		return nil
	case *ast.CharLiteral:
		funcBody.WriteByte(OpCodeI32Const)
		funcBody.Write(EncodeSLEB128(int32(e.Value)))
		return nil
	case *ast.Identifier:
		index := e.Symbol.Index
		funcBody.WriteByte(OpCodeLocalGet)
		funcBody.Write(EncodeULEB128(uint32(index)))
		return nil
	case *ast.PrefixExpression:
		switch e.Operator {
		case "-":
			if e.Right.ValueType() == ast.ValueTypeInt || e.Right.ValueType() == ast.ValueTypeChar {
				funcBody.WriteByte(OpCodeI32Const)
				funcBody.Write(EncodeSLEB128(0))
				m.generateExpressionCode(e.Right, funcBody)
				funcBody.WriteByte(OpCodeI32Sub)
			} else if e.Right.ValueType() == ast.ValueTypeFloat {
				m.generateExpressionCode(e.Right, funcBody)
				funcBody.WriteByte(OpCodeF32Neg)
			}
		default:
			return errors.ErrUnsupported
		}

	case *ast.InfixExpression:
		if !checkCompatibleTypes(e.Left.ValueType(), e.Right.ValueType()) {
			return errors.New("incompatible types for infix operation`")
		}

		m.generateExpressionCode(e.Left, funcBody)
		m.generateExpressionCode(e.Right, funcBody)
		switch e.Operator {
		case "+":
			if e.Left.ValueType() == ast.ValueTypeInt || e.Left.ValueType() == ast.ValueTypeChar {
				funcBody.WriteByte(OpCodeI32Add)
			} else if e.Left.ValueType() == ast.ValueTypeFloat {
				funcBody.WriteByte(OpCodeF32Add)
			}
		case "-":
			if e.Left.ValueType() == ast.ValueTypeInt || e.Left.ValueType() == ast.ValueTypeChar {
				funcBody.WriteByte(OpCodeI32Sub)
			} else if e.Left.ValueType() == ast.ValueTypeFloat {
				funcBody.WriteByte(OpCodeF32Sub)
			}
		case "*":
			if e.Left.ValueType() == ast.ValueTypeInt || e.Left.ValueType() == ast.ValueTypeChar {
				funcBody.WriteByte(OpCodeI32Mul)
			} else if e.Left.ValueType() == ast.ValueTypeFloat {
				funcBody.WriteByte(OpCodeF32Mul)
			}
		case "/":
			if e.Left.ValueType() == ast.ValueTypeInt || e.Left.ValueType() == ast.ValueTypeChar {
				funcBody.WriteByte(OpCodeI32SignedDivision)
			} else if e.Left.ValueType() == ast.ValueTypeFloat {
				funcBody.WriteByte(OpCodeF32Division)
			}
		}

		if e.Left.ValueType() == ast.ValueTypeChar {
			// make sure characters wrap around after 255
			funcBody.WriteByte(OpCodeI32Const)
			funcBody.Write(EncodeSLEB128(255))
			funcBody.WriteByte(OpCodeI32And)
		}

		return nil
	default:
		return errors.New("unsupported expression")
	}

	return nil
}

func (m *WASMModule) generateVariableDefinition(stmt *ast.VariableDefineStatement, funcBody *bytes.Buffer) error {
	if stmt.Value == nil {
		// variable is jsut defined, has not been set yet
		return nil
	}

	m.generateExpressionCode(stmt.Value, funcBody)
	index := stmt.Name.Symbol.Index
	funcBody.WriteByte(OpCodeLocalSet)
	funcBody.Write(EncodeULEB128(uint32(index)))

	return nil
}

func (m *WASMModule) generateReturnStatement(stmt *ast.ReturnStatement, funcBody *bytes.Buffer) error {
	m.generateExpressionCode(stmt.ReturnValue, funcBody)
	funcBody.WriteByte(OpCodeReturn)
	return nil
}

func (m *WASMModule) generateVariableUpdate(stmt *ast.VariableUpdateStatement, funcBody *bytes.Buffer) error {
	m.generateExpressionCode(stmt.NewValue, funcBody)
	funcBody.WriteByte(OpCodeLocalSet)
	funcBody.Write(EncodeULEB128(uint32(stmt.Name.Symbol.Index)))
	return nil
}

func (m *WASMModule) generateStatement(stmt ast.Statement, funcBody *bytes.Buffer) error {
	switch s := stmt.(type) {
	case *ast.VariableDefineStatement:
		return m.generateVariableDefinition(s, funcBody)
	case *ast.ReturnStatement:
		return m.generateReturnStatement(s, funcBody)
	case *ast.VariableUpdateStatement:
		return m.generateVariableUpdate(s, funcBody)
	default:
		return errors.New("unknown statement type")
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
