package codegen

import (
	"bytes"
	"errors"
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
	OpCodeReturn            byte = 0x0B
	OpCodeI32Const          byte = 0x41
	OpCodeLocalGet          byte = 0x20
	OpCodeLocalSet          byte = 0x21
	OpCodeI32Add            byte = 0x6A
	OpCodeI32Sub            byte = 0x6B
	OpCodeI32Mul            byte = 0x6C
	OpCodeI32SignedDivision byte = 0x6D
)

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
	typePayload.Write(encodeULEB128(uint32(len(m.program.Functions)))) // num of types (functions)
	typePayload.WriteByte(0x60)                                        // identifies it as a function
	typePayload.Write(encodeULEB128(0))                                // Param count: 0
	typePayload.Write(encodeULEB128(1))                                // Result count: 1
	typePayload.WriteByte(0x7F)                                        // type: i32

	m.writeSection(SecType, typePayload.Bytes())

}

func (m *WASMModule) generateFunctionSection() {
	funcPayload := bytes.Buffer{}
	funcPayload.Write(encodeULEB128(1)) // Number of functions
	funcPayload.Write(encodeULEB128(0)) // Index of the type defined above
	m.writeSection(SecFunction, funcPayload.Bytes())
}

func (m *WASMModule) generateExportSection() {
	exportPayload := bytes.Buffer{}
	exportPayload.Write(encodeULEB128(1)) // Number of exports
	exportPayload.Write(encodeULEB128(4)) // String length
	exportPayload.WriteString("main")
	exportPayload.WriteByte(0x00)         // Export kind: Function
	exportPayload.Write(encodeULEB128(0)) // Function index: 0
	m.writeSection(SecExport, exportPayload.Bytes())
}

func (m *WASMModule) generateExpressionCode(value parser.Node, body *bytes.Buffer) {
	if constant, ok := value.(parser.Constant); ok {
		body.WriteByte(OpCodeI32Const)
		intValue, _ := strconv.Atoi(constant.Value)

		body.Write(encodeSLEB128(int32(intValue)))
	} else if varAccess, ok := value.(parser.VariableAccess); ok {
		body.WriteByte(OpCodeLocalGet)
		body.Write(encodeULEB128(uint32(varAccess.Index)))

	} else if unary, ok := value.(parser.UnaryExpression); ok {
		// negate valuing by generating code for 0 - value
		body.WriteByte(OpCodeI32Const)
		body.Write(encodeSLEB128(0))
		m.generateExpressionCode(unary.Value, body)
		body.WriteByte(OpCodeI32Sub)
	} else if binaryExpr, ok := value.(parser.BinaryExpression); ok {
		m.generateExpressionCode(binaryExpr.A, body)
		m.generateExpressionCode(binaryExpr.B, body)
		switch binaryExpr.Operation {
		case "-":
			body.WriteByte(OpCodeI32Sub)
		case "+":
			body.WriteByte(OpCodeI32Add)
		case "/":
			body.WriteByte(OpCodeI32SignedDivision)
		case "*":
			body.WriteByte(OpCodeI32Mul)
		}
	}

}

func (m *WASMModule) generateCodeSection() error {
	codePayload := bytes.Buffer{}
	codePayload.Write(encodeULEB128(1)) // number of function bodies

	body := bytes.Buffer{}
	body.Write(encodeULEB128(1))                                               // 1 group of locals
	body.Write(encodeULEB128(uint32(len(m.program.Functions[0].SymbolTable)))) // local variable count
	body.WriteByte(0x7F)                                                       // Type i32

	mainFunc := m.program.Functions[0]
	for _, stmt := range mainFunc.Body.Statements {
		if retStmt, ok := stmt.(parser.ReturnStmt); ok {
			m.generateExpressionCode(retStmt.Value, &body)
			body.WriteByte(OpCodeReturn)

		} else if varDefineStmt, ok := stmt.(parser.VariableDefineStmt); ok {
			m.generateExpressionCode(varDefineStmt.Value, &body)
			variableIndex := varDefineStmt.Symbol.Index

			body.WriteByte(OpCodeLocalSet)
			body.Write(encodeULEB128(uint32(variableIndex)))

		} else {
			return errors.New("unsupported statement")
		}
	}

	codePayload.Write(encodeULEB128(uint32(body.Len())))
	codePayload.Write(body.Bytes())
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
