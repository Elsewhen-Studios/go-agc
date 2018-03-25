package assembler

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testLastChannel  = 0x01FF
	testLastErasable = 0x03FF
	testLastFixed    = 0x0FFF
)

func Test_requireOperand_valid(t *testing.T) {
	// arrange
	op := "SYMBOL"
	s := " \t" + op
	ts := bufio.NewScanner(strings.NewReader(s))
	ts.Split(bufio.ScanWords)
	p := new(instructionParams)

	// act
	err := requireOperand(ts, p)

	// assert
	assert.NoError(t, err)
	assert.EqualValues(t, op, p.operandToken, "operand token")
}

func Test_requireOperand_noOperand(t *testing.T) {
	// arrange
	ts := bufio.NewScanner(strings.NewReader(""))
	ts.Split(bufio.ScanWords)
	p := new(instructionParams)

	// act
	err := requireOperand(ts, p)

	// assert
	assert.Error(t, err)
}

func Test_clasifyToken_octal(t *testing.T) {
	assert.EqualValues(t, octalToken, clasifyToken("01234567"))
}

func Test_clasifyToken_decimal(t *testing.T) {
	assert.EqualValues(t, decimalToken, clasifyToken("0123456789"))
}

func Test_clasifyToken_symbol(t *testing.T) {
	assert.EqualValues(t, symbolToken, clasifyToken("012value89"))
}

func Test_validateTCOperand_valid(t *testing.T) {
	// arrange
	values := []uint16{0x0000, 02, 05, 07, testLastFixed}
	for _, v := range values {
		p := new(instructionParams)
		// act
		result := validateTCOperand(v, p)
		// assert
		assert.NoErrorf(t, result, "operand %#o", v)
	}
}

func Test_validateTCOperand_invalid(t *testing.T) {
	// arrange
	values := []uint16{03, 04, 06, testLastFixed + 1, 0xFFFF}
	for _, v := range values {
		p := new(instructionParams)
		// act
		result := validateTCOperand(v, p)
		// assert
		assert.Errorf(t, result, "operand %#o", v)
	}
}

func Test_validateINDEXOperand_valid(t *testing.T) {
	// arrange
	values := []uint16{0x0000, 016, 020, testLastErasable}
	for _, v := range values {
		p := new(instructionParams)
		// act
		result := validateINDEXOperand(v, p)
		// assert
		assert.NoErrorf(t, result, "operand %#o", v)
	}
}

func Test_validateINDEXOperand_invalid(t *testing.T) {
	// arrange
	values := []uint16{017, testLastErasable + 1, 0xFFFF}
	for _, v := range values {
		p := new(instructionParams)
		// act
		result := validateINDEXOperand(v, p)
		// assert
		assert.Errorf(t, result, "operand %#o", v)
	}
}

func Test_requireAnyMemoryOperand_valid(t *testing.T) {
	// arrange
	values := []uint16{0x0000, testLastFixed}
	for _, v := range values {
		p := new(instructionParams)
		// act
		result := requireAnyMemoryOperand(v, p)
		// assert
		assert.NoErrorf(t, result, "operand %#o", v)
	}
}

func Test_requireAnyMemoryOperand_invalid(t *testing.T) {
	// arrange
	values := []uint16{testLastFixed + 1, 0xFFFF}
	for _, v := range values {
		p := new(instructionParams)
		// act
		result := requireAnyMemoryOperand(v, p)
		// assert
		assert.Errorf(t, result, "operand %#o", v)
	}
}

func Test_requireErasable_valid(t *testing.T) {
	// arrange
	values := []uint16{0x0000, testLastErasable}
	for _, v := range values {
		p := new(instructionParams)
		// act
		result := requireErasable(v, p)
		// assert
		assert.NoErrorf(t, result, "operand %#o", v)
	}
}

func Test_requireErasable_invalid(t *testing.T) {
	// arrange
	values := []uint16{testLastErasable + 1, 0xFFFF}
	for _, v := range values {
		p := new(instructionParams)
		// act
		result := requireErasable(v, p)
		// assert
		assert.Errorf(t, result, "operand %#o", v)
	}
}

func Test_requireChannel_valid(t *testing.T) {
	// arrange
	values := []uint16{0x0000, testLastChannel}
	for _, v := range values {
		p := new(instructionParams)
		// act
		result := requireChannel(v, p)
		// assert
		assert.NoErrorf(t, result, "operand %#o", v)
	}
}

func Test_requireChannel_invalid(t *testing.T) {
	// arrange
	values := []uint16{testLastChannel + 1, 0xFFFF}
	for _, v := range values {
		p := new(instructionParams)
		// act
		result := requireChannel(v, p)
		// assert
		assert.Errorf(t, result, "operand %#o", v)
	}
}

func Test_requireFixed_valid(t *testing.T) {
	// arrange
	values := []uint16{testLastErasable + 1, testLastFixed}
	for _, v := range values {
		p := new(instructionParams)
		// act
		result := requireFixed(v, p)
		// assert
		assert.NoErrorf(t, result, "operand %#o", v)
	}
}

func Test_requireFixed_invalid(t *testing.T) {
	// arrange
	values := []uint16{0x0000, testLastErasable, testLastFixed + 1, 0xFFFF}
	for _, v := range values {
		p := new(instructionParams)
		// act
		result := requireFixed(v, p)
		// assert
		assert.Errorf(t, result, "operand %#o", v)
	}
}
