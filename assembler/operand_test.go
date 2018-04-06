package assembler

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testLastChannel            = 0x01FF
	testLastUnswitchedErasable = 0x02FF
	testLastErasable           = 0x03FF
	testLastSwitchedFixed      = 0x07FF
	testLastFixed              = 0x0FFF
)

func Test_requireOperand_valid(t *testing.T) {
	// arrange
	op := "SYMBOL"
	s := " \t" + op
	ts := bufio.NewScanner(strings.NewReader(s))
	ts.Split(bufio.ScanWords)
	sp := newScannerPeeker(ts)
	p := new(instructionParams)

	// act
	err := requireOperand(sp, p)

	// assert
	assert.NoError(t, err)
	assert.EqualValues(t, op, p.operandToken, "operand token")
}

func Test_requireOperand_noOperand(t *testing.T) {
	// arrange
	ts := bufio.NewScanner(strings.NewReader(""))
	ts.Split(bufio.ScanWords)
	sp := newScannerPeeker(ts)
	p := new(instructionParams)

	// act
	err := requireOperand(sp, p)

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
	a, pl := buildAssemblerLogger()
	values := []uint16{0x0000, 02, 05, 07, testLastFixed}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := validateTCOperand(v, p)

		// assert
		assert.Truef(t, result, "result (%#o)", v)
		assert.Len(t, a.Problems, probCount, "problem count (%#o)", v)
	}
}

func Test_validateTCOperand_invalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{03, 04, 06, testLastFixed + 1, 0xFFFF}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := validateTCOperand(v, p)

		// assert
		assert.Falsef(t, result, "result (%#o)", v)
		if assert.Len(t, a.Problems, probCount+1, "problem count (%#o)", v) {
			assert.EqualValues(t, ProblemKindError, a.Problems[probCount].Kind, "problem kind (%#o)", v)
		}
	}
}

func Test_validateINDEXOperand_valid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{0x0000, 016, 020, testLastErasable}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := validateINDEXOperand(v, p)

		// assert
		assert.Truef(t, result, "result (%#o)", v)
		assert.Len(t, a.Problems, probCount, "problem count (%#o)", v)
	}
}

func Test_validateINDEXOperand_invalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{017, testLastErasable + 1, 0xFFFF}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := validateINDEXOperand(v, p)

		// assert
		assert.Falsef(t, result, "result (%#o)", v)
		if assert.Len(t, a.Problems, probCount+1, "problem count (%#o)", v) {
			assert.EqualValues(t, ProblemKindError, a.Problems[probCount].Kind, "problem kind (%#o)", v)
		}
	}
}

func Test_requireAnyMemoryOperand_valid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{0x0000, testLastFixed}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := requireAnyMemoryOperand(v, p)

		// assert
		assert.Truef(t, result, "result (%#o)", v)
		assert.Len(t, a.Problems, probCount, "problem count (%#o)", v)
	}
}

func Test_requireAnyMemoryOperand_invalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{testLastFixed + 1, 0xFFFF}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := requireAnyMemoryOperand(v, p)

		// assert
		assert.Falsef(t, result, "result (%#o)", v)
		if assert.Len(t, a.Problems, probCount+1, "problem count (%#o)", v) {
			assert.EqualValues(t, ProblemKindError, a.Problems[probCount].Kind, "problem kind (%#o)", v)
		}
	}
}

func Test_requireDoubleAnyMemoryOperand_valid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{0x0000, testLastFixed - 1}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := requireDoubleAnyMemoryOperand(v, p)

		// assert
		assert.Truef(t, result, "result (%#o)", v)
		assert.Len(t, a.Problems, probCount, "problem count (%#o)", v)
	}
}

func Test_requireDoubleAnyMemoryOperand_warning(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{testLastUnswitchedErasable, testLastErasable, testLastSwitchedFixed}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := requireDoubleAnyMemoryOperand(v, p)

		// assert
		assert.Truef(t, result, "result (%#o)", v)
		if assert.Len(t, a.Problems, probCount+1, "problem count (%#o)", v) {
			assert.EqualValues(t, ProblemKindWarning, a.Problems[probCount].Kind, "problem kind (%#o)", v)
		}
	}
}

func Test_requireDoubleAnyMemoryOperand_invalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{testLastFixed, 0xFFFF}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := requireDoubleAnyMemoryOperand(v, p)

		// assert
		assert.Falsef(t, result, "result (%#o)", v)
		if assert.Len(t, a.Problems, probCount+1, "problem count (%#o)", v) {
			assert.EqualValues(t, ProblemKindError, a.Problems[probCount].Kind, "problem kind (%#o)", v)
		}
	}
}

func Test_requireErasable_valid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{0x0000, testLastErasable}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := requireErasable(v, p)

		// assert
		assert.Truef(t, result, "result (%#o)", v)
		assert.Len(t, a.Problems, probCount, "problem count (%#o)", v)
	}
}

func Test_requireErasable_invalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{testLastErasable + 1, 0xFFFF}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := requireErasable(v, p)

		// assert
		assert.Falsef(t, result, "result (%#o)", v)
		if assert.Len(t, a.Problems, probCount+1, "problem count (%#o)", v) {
			assert.EqualValues(t, ProblemKindError, a.Problems[probCount].Kind, "problem kind (%#o)", v)
		}
	}
}

func Test_requireDoubleErasable_valid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{0x0000, testLastErasable - 1}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := requireDoubleErasable(v, p)

		// assert
		assert.Truef(t, result, "result (%#o)", v)
		assert.Len(t, a.Problems, probCount, "problem count (%#o)", v)
	}
}

func Test_requireDoubleErasable_warning(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{testLastUnswitchedErasable}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := requireDoubleErasable(v, p)

		// assert
		assert.Truef(t, result, "result (%#o)", v)
		if assert.Len(t, a.Problems, probCount+1, "problem count (%#o)", v) {
			assert.EqualValues(t, ProblemKindWarning, a.Problems[probCount].Kind, "problem kind (%#o)", v)
		}
	}
}

func Test_requireDoubleErasable_invalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{testLastErasable, 0xFFFF}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := requireDoubleErasable(v, p)

		// assert
		assert.Falsef(t, result, "result (%#o)", v)
		if assert.Len(t, a.Problems, probCount+1, "problem count (%#o)", v) {
			assert.EqualValues(t, ProblemKindError, a.Problems[probCount].Kind, "problem kind (%#o)", v)
		}
	}
}

func Test_requireChannel_valid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{0x0000, testLastChannel}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := requireChannel(v, p)

		// assert
		assert.Truef(t, result, "result (%#o)", v)
		assert.Len(t, a.Problems, probCount, "problem count (%#o)", v)
	}
}

func Test_requireChannel_invalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{testLastChannel + 1, 0xFFFF}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := requireChannel(v, p)

		// assert
		assert.Falsef(t, result, "result (%#o)", v)
		if assert.Len(t, a.Problems, probCount+1, "problem count (%#o)", v) {
			assert.EqualValues(t, ProblemKindError, a.Problems[probCount].Kind, "problem kind (%#o)", v)
		}
	}
}

func Test_requireFixed_valid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{testLastErasable + 1, testLastFixed}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := requireFixed(v, p)

		// assert
		assert.Truef(t, result, "result (%#o)", v)
		assert.Len(t, a.Problems, probCount, "problem count (%#o)", v)
	}
}

func Test_requireFixed_invalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	values := []uint16{0x0000, testLastErasable, testLastFixed + 1, 0xFFFF}

	for _, v := range values {
		p := &instructionParams{logger: pl}
		probCount := len(a.Problems)

		// act
		result := requireFixed(v, p)

		// assert
		assert.Falsef(t, result, "result (%#o)", v)
		if assert.Len(t, a.Problems, probCount+1, "problem count (%#o)", v) {
			assert.EqualValues(t, ProblemKindError, a.Problems[probCount].Kind, "problem kind (%#o)", v)
		}
	}
}
