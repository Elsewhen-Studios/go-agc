package assembler

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_setLoc_valid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	a.setLocation(psudoAddress(0xFFFF))

	// act
	a.parseLine(pl, "SETLOC 2000")

	// assert
	assert.Len(t, a.Problems, 0, "error count")

	assert.NotNil(t, a.location, "location")
	assert.EqualValues(t, 02000, int(a.location), "location")
	assert.Len(t, a.operations, 1, "operation count")
}

func Test_setLoc_invalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	a.setLocation(psudoAddress(04000))

	// act
	a.parseLine(pl, "SETLOC 14000")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}

	assert.NotNil(t, a.location, "location")
	assert.EqualValues(t, 04000, int(a.location), "location")
	assert.Len(t, a.operations, 0, "operation count")
}

func Test_setLoc_noOperand(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	a.setLocation(psudoAddress(04000))

	// act
	a.parseLine(pl, "SETLOC ")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}

	assert.NotNil(t, a.location, "location")
	assert.EqualValues(t, 04000, int(a.location), "location")
	assert.Len(t, a.operations, 0, "operation count")
}

func Test_setLoc_unresolvable(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	a.setLocation(psudoAddress(04000))

	// act
	a.parseLine(pl, "SETLOC FOO")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}

	assert.NotNil(t, a.location, "location")
	assert.EqualValues(t, 04000, int(a.location), "location")
	assert.Len(t, a.operations, 0, "operation count")
}

func Test_oct_validExecute(t *testing.T) {
	// arrange
	a := new(Assembler)
	loc := 04075
	v := uint16(076)
	prog := fmt.Sprintf("SETLOC %o\nOCT %o\n", loc, v)

	// act
	ok := a.assembleReader(strings.NewReader(prog), "fake_file.asm")

	// assert
	assert.True(t, ok, "result")
	assert.Len(t, a.Problems, 0, "problem count")
	assertImageValue(t, v, a, 2, uint16(loc&01777))
}

func Test_oct_decimal(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	a.setLocation(psudoAddress(04000))

	// act
	a.parseLine(pl, "OCT 1289")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
	assert.Len(t, a.operations, 0, "operation count")
}

func Test_oct_toLarge(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	a.setLocation(psudoAddress(04000))

	// act
	a.parseLine(pl, "OCT 100000")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
	assert.Len(t, a.operations, 0, "operation count")
}

func Test_oct_veryLarge(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	a.setLocation(psudoAddress(04000))

	// act
	a.parseLine(pl, "OCT 7654321")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
	assert.Len(t, a.operations, 0, "operation count")
}

func Test_oct_noOperand(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	a.setLocation(psudoAddress(04000))

	// act
	a.parseLine(pl, "OCT ")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
	assert.Len(t, a.operations, 0, "operation count")
}

const (
	negZero = 1 << 14

	max14Bit = (1 << 14) - 1
	max28Bit = (1 << 28) - 1
)

func convertTo15BitOnes(v int) uint16 {
	var neg bool
	if v < 0 {
		neg = true
		v = -v
	}

	if v > max14Bit {
		panic("value out of range")
	}

	if neg {
		v |= negZero
	}

	return uint16(v)
}

func convertTo30BitOnes(v int) (h uint16, l uint16) {
	var neg bool
	if v < 0 {
		neg = true
		v = -v
	}

	if v > max28Bit {
		panic("value out of range")
	}

	l = uint16(v & max14Bit)
	h = uint16(v >> 14)

	if neg {
		l |= negZero
		h |= negZero
	}

	return
}

func Test_tryParseDec_posIntegerSingle(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	value := 012345
	line := fmt.Sprintf("%d", value)
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, convertTo15BitOnes(value), h, "high word")
	assert.EqualValues(t, 0, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_posIntegerDouble(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	value := 01234567012
	vh, vl := convertTo30BitOnes(value)
	line := fmt.Sprintf("%d", value)
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, true)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, vh, h, "high word")
	assert.EqualValues(t, vl, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_negIntegerSingle(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	value := -012345
	line := fmt.Sprintf("%d", value)
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, convertTo15BitOnes(value), h, "high word")
	assert.EqualValues(t, negZero, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_negIntegerDouble(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	value := -01234567012
	vh, vl := convertTo30BitOnes(value)
	line := fmt.Sprintf("%d", value)
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, true)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, vh, h, "high word")
	assert.EqualValues(t, vl, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_negIntegerSingleZero(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "-0"
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, negZero, h, "high word")
	assert.EqualValues(t, negZero, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_negIntegerDoubleZero(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "-0"
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, true)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, negZero, h, "high word")
	assert.EqualValues(t, negZero, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_posIntegerSingleLarge(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := fmt.Sprintf("%d", max14Bit+1)
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	_, _, ok := tryParseDec(sp, p, false)

	// assert
	assert.False(t, ok, "result")

	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_tryParseDec_posIntegerDoubleLarge(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := fmt.Sprintf("%d", max28Bit+1)
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	_, _, ok := tryParseDec(sp, p, true)

	// assert
	assert.False(t, ok, "result")

	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_tryParseDec_negIntegerSingleLarge(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := fmt.Sprintf("%d", -(max14Bit + 1))
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	_, _, ok := tryParseDec(sp, p, false)

	// assert
	assert.False(t, ok, "result")

	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_tryParseDec_negIntegerDoubleLarge(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := fmt.Sprintf("%d", -(max28Bit + 1))
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	_, _, ok := tryParseDec(sp, p, true)

	// assert
	assert.False(t, ok, "result")

	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_tryParseDec_IntegerInvalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "123FooBar678"
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	_, _, ok := tryParseDec(sp, p, false)

	// assert
	assert.False(t, ok, "result")

	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_tryParseDec_posFloatNoExponent(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "0.75"
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, 030000, h, "high word")
	assert.EqualValues(t, 0, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_negFloatNoExponent(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "-0.75"
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, 030000|negZero, h, "high word")
	assert.EqualValues(t, negZero, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_posFloat0Exponent(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "0.75 E0 B0"
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, 030000, h, "high word")
	assert.EqualValues(t, 0, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_negFloat0Exponent(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "-0.75 E0 B0"
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, 030000|negZero, h, "high word")
	assert.EqualValues(t, negZero, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_posFloatDecExponent(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "75 E-2" //0.75
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, 030000, h, "high word")
	assert.EqualValues(t, 0, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_negFloatDecExponent(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "-75 E-2" //-0.75
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, 030000|negZero, h, "high word")
	assert.EqualValues(t, negZero, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_posFloatBinExponent(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "3 B-2" //0.75
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, 030000, h, "high word")
	assert.EqualValues(t, 0, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_negFloatBinExponent(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "-3 B-2" //-0.75
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, 030000|negZero, h, "high word")
	assert.EqualValues(t, negZero, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_posFloatDecAndBinExponent(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "300 E-2 B-2" //0.75
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, 030000, h, "high word")
	assert.EqualValues(t, 0, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_negFloatDecAndBinExponent(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "-300 E-2 B-2" //-0.75
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, 030000|negZero, h, "high word")
	assert.EqualValues(t, negZero, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_posFloatTiny(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "1 B-28" //1.0
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, 0, h, "high word")
	assert.EqualValues(t, 1, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_negFloatTiny(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "-1 B-28"
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, negZero, h, "high word")
	assert.EqualValues(t, 1|negZero, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_negFloatZero(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "-0."
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	h, l, ok := tryParseDec(sp, p, false)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, negZero, h, "high word")
	assert.EqualValues(t, negZero, l, "low word")

	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_tryParseDec_posFloatLarge(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "50.0 E-2 B1" //1.0
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	_, _, ok := tryParseDec(sp, p, false)

	// assert
	assert.False(t, ok, "result")

	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_tryParseDec_negFloatLarge(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "-1.0"
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	_, _, ok := tryParseDec(sp, p, false)

	// assert
	assert.False(t, ok, "result")

	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_tryParseDec_FloatInvalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "1.0FooBar123"
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	_, _, ok := tryParseDec(sp, p, false)

	// assert
	assert.False(t, ok, "result")

	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_tryParseDec_FloatExponentInvalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "0.75 Q2"
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	_, _, ok := tryParseDec(sp, p, false)

	// assert
	assert.False(t, ok, "result")

	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_tryParseDec_FloatDecExponentInvalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "0.75 E2Foo"
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	_, _, ok := tryParseDec(sp, p, false)

	// assert
	assert.False(t, ok, "result")

	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_tryParseDec_FloatBinExponentInvalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl}

	line := "0.75 B2Foo"
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	_, _, ok := tryParseDec(sp, p, false)

	// assert
	assert.False(t, ok, "result")

	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_dec_valid(t *testing.T) {
	// arrange
	a := new(Assembler)

	value := 012345
	prog := fmt.Sprintf("SETLOC 4000\nDEC %d\nOCT 77777", value)

	// act
	ok := a.assembleReader(strings.NewReader(prog), "fake_file.asm")

	// assert
	assert.True(t, ok, "result")
	assert.Len(t, a.Problems, 0, "problem count")
	assertImageValue(t, convertTo15BitOnes(value), a, 2, 0)
	assertImageValue(t, 077777, a, 2, 1)
}

func Test_dec_noOperand(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	a.setLocation(psudoAddress(04000))

	// act
	a.parseLine(pl, "DEC ")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}

	assert.NotNil(t, a.location, "location")
	assert.EqualValues(t, 04000, int(a.location), "location")
	assert.Len(t, a.operations, 0, "operation count")
}

func Test_dec2_valid(t *testing.T) {
	// arrange
	a := new(Assembler)

	value := 01234567012
	vh, vl := convertTo30BitOnes(value)
	prog := fmt.Sprintf("SETLOC 4000\n2DEC %d\nOCT 77777", value)

	// act
	ok := a.assembleReader(strings.NewReader(prog), "fake_file.asm")

	// assert
	assert.True(t, ok, "result")
	assert.Len(t, a.Problems, 0, "problem count")
	assertImageValue(t, vh, a, 2, 0)
	assertImageValue(t, vl, a, 2, 1)
	assertImageValue(t, 077777, a, 2, 2)
}

func Test_dec2_noOperand(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	a.setLocation(psudoAddress(04000))

	// act
	a.parseLine(pl, "2DEC ")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}

	assert.NotNil(t, a.location, "location")
	assert.EqualValues(t, 04000, int(a.location), "location")
	assert.Len(t, a.operations, 0, "operation count")
}
