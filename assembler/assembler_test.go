package assembler

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildAssemblerLogger() (a *Assembler, pl problemLogger) {
	a = new(Assembler)
	pl = &assemblerLogger{asm: a, fileName: "test.asm", lineNum: 0}

	return
}

func Test_assembler_ErrorCount(t *testing.T) {
	// arrange
	a := new(Assembler)
	a.errorCount = 7

	// act
	result := a.ErrorCount()

	// assert
	assert.EqualValues(t, a.errorCount, result)
}

func Test_aliases(t *testing.T) {
	// arrange

	// act

	// assert
	for k, v := range aliases {
		match := false

		if _, ok := standardInstructions[v]; ok {
			match = true
		} else if _, ok := extendedInstructions[v]; ok {
			match = true
		}

		if _, ok := directives[v]; ok {
			assert.Falsef(t, match, "ambiguous alias definition (%v=>%v)", k, v)
			match = true
		}

		assert.Truef(t, match, "un-used alias definition (%v=>%v)", k, v)
	}
}

func Test_instructionParams_resolveOperand_validOctal(t *testing.T) {
	// arrange
	value := uint16(01357)
	p := instructionParams{operandToken: fmt.Sprintf("%#o", value)}

	// act
	result, err := p.resolveOperand()

	// assert
	assert.NoError(t, err)
	assert.EqualValues(t, value, result)
}

func Test_instructionParams_resolveOperand_invalidOctal(t *testing.T) {
	// arrange
	p := instructionParams{operandToken: "01234567"}

	// act
	_, err := p.resolveOperand()

	// assert
	assert.Error(t, err)
}

func Test_instructionParams_resolveOperand_decimal(t *testing.T) {
	// arrange
	p := instructionParams{operandToken: "2480"}

	// act
	_, err := p.resolveOperand()

	// assert
	assert.Error(t, err)
}

func Test_instructionParams_resolveOperand_validSymbol(t *testing.T) {
	// arrange
	value := uint16(01234)
	name := "TEST123"
	a := new(Assembler)
	a.symbols = map[string]uint16{name: value}
	p := instructionParams{operandToken: name}
	p.resolver = a

	// act
	result, err := p.resolveOperand()

	// assert
	assert.NoError(t, err)
	assert.EqualValues(t, value, result)
}

func Test_defineSymbolHandler_validSymbolEquals(t *testing.T) {
	// arrange
	symbol := "FOO"
	value := uint16(0777)
	a, pl := buildAssemblerLogger()

	// act
	a.parseLine(pl, fmt.Sprintf("%v = %#o", symbol, value))

	// assert
	assert.Len(t, a.Problems, 0, "problem count")

	result, ok := a.symbols[symbol]
	assert.True(t, ok, "symbol found")
	assert.EqualValues(t, value, result, "value")
}

func Test_defineSymbolHandler_validWordEquals(t *testing.T) {
	// arrange
	symbol := "FOO"
	value := uint16(0777)
	a, pl := buildAssemblerLogger()

	// act
	a.parseLine(pl, fmt.Sprintf("%v EQUALS %#o", symbol, value))

	// assert
	assert.Len(t, a.Problems, 0, "problem count")

	result, ok := a.symbols[symbol]
	assert.True(t, ok, "symbol found")
	assert.EqualValues(t, value, result, "value")
}

func Test_defineSymbolHandler_redefinedNewValue(t *testing.T) {
	// arrange
	symbol := "FOO"
	value1 := uint16(0111)
	value2 := uint16(0222)
	a, pl := buildAssemblerLogger()

	a.parseLine(pl, fmt.Sprintf("%v = %#o", symbol, value1))
	require.Len(t, a.Problems, 0, "arrange failed")

	// act
	a.parseLine(pl, fmt.Sprintf("%v = %#o", symbol, value2))

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindWarning, a.Problems[0].Kind, "problem kind")
	}

	result, ok := a.symbols[symbol]
	assert.True(t, ok, "symbol found")
	assert.EqualValues(t, value2, result, "value")
}

func Test_defineSymbolHandler_redefinedSameValue(t *testing.T) {
	// arrange
	symbol := "FOO"
	value := uint16(0333)
	a, pl := buildAssemblerLogger()

	a.parseLine(pl, fmt.Sprintf("%v = %#o", symbol, value))
	require.Len(t, a.Problems, 0, "arrange failed")

	// act
	a.parseLine(pl, fmt.Sprintf("%v = %#o", symbol, value))

	// assert
	assert.Len(t, a.Problems, 0, "problem count")

	result, ok := a.symbols[symbol]
	assert.True(t, ok, "symbol found")
	assert.EqualValues(t, value, result, "value")
}

func Test_defineSymbolHandler_noOperand(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()

	// act
	a.parseLine(pl, "FOO = ")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}

	assert.Len(t, a.symbols, 0, "symbols length")
}

func Test_defineSymbolHandler_resolvable(t *testing.T) {
	// arrange
	symbol1 := "FOO"
	value1 := uint16(0111)
	symbol2 := "BAR"
	a, pl := buildAssemblerLogger()
	a.parseLine(pl, fmt.Sprintf("%v = %#o", symbol1, value1))

	// act
	a.parseLine(pl, fmt.Sprintf("%v = %v", symbol2, symbol1))

	// assert
	assert.Len(t, a.Problems, 0, "problem count")

	result, ok := a.symbols[symbol2]
	assert.True(t, ok, "symbol found")
	assert.EqualValues(t, value1, result, "value")
}

func Test_defineSymbolHandler_unresolved(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()

	// act
	a.parseLine(pl, "FOO = BAR")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}

	assert.Len(t, a.symbols, 0, "symbols length")
}

func Test_intentionalBankTransition(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()

	a.parseLine(pl, "SETLOC 032000") //Begining of F9
	require.Len(t, a.Problems, 0, "arrange failed")

	// act
	a.parseLine(pl, "CA 0123")

	// assert
	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_autoBankTransition(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()

	a.parseLine(pl, "SETLOC 031777") //End of F8
	a.parseLine(pl, "CA 0123")
	require.Len(t, a.Problems, 0, "arrange failed")

	// act
	a.parseLine(pl, "CA 0123")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindWarning, a.Problems[0].Kind, "problem kind")
	}
}

func Test_endOfErasableTransition(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()

	a.parseLine(pl, "SETLOC 0003777") //End of E7
	a.parseLine(pl, "CA 0123")
	require.Len(t, a.Problems, 0, "arrange failed")

	// act
	a.parseLine(pl, "CA 0123")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_endOfFixedTransition(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()

	a.parseLine(pl, "SETLOC 0127777") //End of F31+SB
	a.parseLine(pl, "CA 0123")
	require.Len(t, a.Problems, 0, "arrange failed")

	// act
	a.parseLine(pl, "CA 0123")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_bankTransition(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()

	a.parseLine(pl, "SETLOC 031777") //End of F8
	a.parseLine(pl, "CA 0123")
	require.Len(t, a.Problems, 0, "arrange failed")

	// act
	a.parseLine(pl, "CA 0123")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindWarning, a.Problems[0].Kind, "problem kind")
	}
}

func Test_findInstruction_standard(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl, extended: false}

	// act
	inst := findInstruction(p, "CA")

	// assert
	assert.NotNil(t, inst, "return value")
	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_findInstruction_standardWhileExtended(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl, extended: true}

	// act
	inst := findInstruction(p, "CA")

	// assert
	assert.NotNil(t, inst, "return value")
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_findInstruction_extended(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl, extended: true}

	// act
	inst := findInstruction(p, "DCA")

	// assert
	assert.NotNil(t, inst, "return value")
	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_findInstruction_extendedWhileStandard(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl, extended: false}

	// act
	inst := findInstruction(p, "DCA")

	// assert
	assert.NotNil(t, inst, "return value")
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_findInstruction_notFound(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl, extended: false}

	// act
	inst := findInstruction(p, "FOO")

	// assert
	assert.Nil(t, inst, "return value")
	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_trailingCommentWithSpace(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()

	// act
	a.parseLine(pl, "CA 0123 #Comment Goes Here")

	// assert
	assert.Len(t, a.Problems, 0, "problem count")
	assert.Len(t, a.operations, 1, "operation count")
}

func Test_trailingCommentWithOutSpace(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()

	// act
	a.parseLine(pl, "CA 0123#Comment Goes Here")

	// assert
	assert.Len(t, a.Problems, 0, "problem count")
	assert.Len(t, a.operations, 1, "operation count")
}

func Test_fullLineComment(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()

	// act
	a.parseLine(pl, "\t#Comment Goes Here")

	// assert
	assert.Len(t, a.Problems, 0, "problem count")
	assert.Len(t, a.operations, 0, "operation count")
}

func Test_aliasResolving(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()

	// act
	a.parseLine(pl, "TCR 1122") //alias for TC

	// assert
	assert.Len(t, a.Problems, 0, "problem count")
	assert.Len(t, a.operations, 1, "operation count")
}

func Test_missingInst(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()

	// act
	a.parseLine(pl, "TCR")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_demandEndOfLine_unexpectedToken(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()

	// act
	a.parseLine(pl, "CA 0123 CA 0123")

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
	assert.Len(t, a.operations, 1, "operation count")
}

func Test_defineLabel_valid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	loc := 04462
	a.parseLine(pl, fmt.Sprintf("SETLOC %#o", loc))
	label := "FOOBAR"

	// act
	a.parseLine(pl, label)

	// assert
	assert.Len(t, a.Problems, 0, "problem count")
	if assert.Len(t, a.symbols, 1, "symbol count") {
		v, ok := a.symbols[label]
		if assert.True(t, ok, "symbol key exists") {
			assert.EqualValues(t, loc, v)
		}
	}
}

func Test_defineLabel_validWithInst(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	loc := 04473
	a.parseLine(pl, fmt.Sprintf("SETLOC %#o", loc))
	label := "FOOBAR"
	line := label + " CA 0123"

	// act
	a.parseLine(pl, line)

	// assert
	assert.Len(t, a.Problems, 0, "problem count")
	if assert.Len(t, a.symbols, 1, "symbol count") {
		v, ok := a.symbols[label]
		if assert.True(t, ok, "symbol key exists") {
			assert.EqualValues(t, loc, v)
		}
	}
	assert.Len(t, a.operations, 2, "operation count")
}

func Test_defineLabel_undefined(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	a.setLocation(0xFFFF)
	label := "FOOBAR"

	// act
	a.parseLine(pl, label)

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
	assert.Len(t, a.symbols, 0, "symbol count")
}

func Test_defineLabel_unknownCommand(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	loc := 04520
	a.parseLine(pl, fmt.Sprintf("SETLOC %#o", loc))
	label := "FOOBAR"
	line := label + " BAZ 0123"

	// act
	a.parseLine(pl, line)

	// assert
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
	if assert.Len(t, a.symbols, 1, "symbol count") {
		v, ok := a.symbols[label]
		if assert.True(t, ok, "symbol key exists") {
			assert.EqualValues(t, loc, v)
		}
	}
	assert.Len(t, a.operations, 1, "operation count")
}

func Test_EmptyProgram(t *testing.T) {
	// arrange
	a := new(Assembler)
	r := strings.NewReader("")

	// act
	a.assembleReader(r, "test_file.asm")

	// assert
	assert.Len(t, a.Problems, 0, "problem count")
}

func writeToSlice(a *Assembler) ([]byte, error) {
	w := new(bytes.Buffer)
	err := a.WriteOut(w)
	return w.Bytes(), err
}

func Test_WriteOutBeforeAssemble(t *testing.T) {
	// arrange
	a := new(Assembler)

	// act
	img, err := writeToSlice(a)

	// assert
	assert.Error(t, err, "return error")
	assert.Len(t, a.Problems, 0, "problem count")
	assert.Len(t, img, 0, "image length")
}

func Test_ValidImageSize(t *testing.T) {
	// arrange
	a := new(Assembler)
	ok := a.assembleReader(strings.NewReader(""), "fake_file.asm")
	require.True(t, ok, "arrange failed")

	// act
	img, err := writeToSlice(a)

	// assert
	assert.NoError(t, err, "return error")
	assert.Len(t, a.Problems, 0, "problem count")
	assert.Len(t, img, 2*40*1024, "image length")
}

type errorWriter struct {
	delay int
}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	l := len(p)
	if e.delay < l {
		return e.delay, errors.New("errorWriter eventually returns an error")
	}

	e.delay -= l
	return l, nil
}

func Test_WriteOutWithError(t *testing.T) {
	// arrange
	a := new(Assembler)
	ok := a.assembleReader(strings.NewReader(""), "fake_file.asm") //program must be valid
	require.True(t, ok, "arrange failed")
	w := &errorWriter{delay: 1024}

	// act
	err := a.WriteOut(w)

	// assert
	assert.Error(t, err, "return error")
	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_assembler_writeWordToImage_valid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	a.setLocation(psudoAddress(04617))
	a.image = make(map[uint16]*[1024]uint16)

	// act
	ok := a.writeWordToImage(pl, 01234)

	// assert
	assert.True(t, ok, "result")
	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_assembler_writeWordToImage_invalid(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	a.setLocation(psudoAddress(0130000))
	a.image = make(map[uint16]*[1024]uint16)

	// act
	ok := a.writeWordToImage(pl, 01234)

	// assert
	assert.False(t, ok, "result")
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_assembler_writeWordToImage_erasable(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	a.setLocation(psudoAddress(0647))
	a.image = make(map[uint16]*[1024]uint16)

	// act
	ok := a.writeWordToImage(pl, 01234)

	// assert
	assert.False(t, ok, "result")
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_assembler_Assemble_singleInstructionValid(t *testing.T) {
	// arrange
	a := new(Assembler)
	prog := "SETLOC 04000\nSTART\tTCF START\n"

	// act
	ok := a.assembleReader(strings.NewReader(prog), "fake_file.asm")

	// assert
	assert.True(t, ok, "result")
	assert.Len(t, a.Problems, 0, "problem count")
	assertImageValue(t, uint16(014000), a, 2, 0)
	assertImageValue(t, uint16(000000), a, 2, 1)
}

func Test_assembler_Assemble_stopParsingAfter10Errors(t *testing.T) {
	// arrange
	a := new(Assembler)
	prog := "CA\nCA\nCA\nCA\nCA\nCA\nCA\nCA\nCA\nCA\nCA\nCA\nCA\nCA\nCA\nCA\nCA\nCA\nCA\nCA\n" //this would generate 20 operand missing errors

	// act
	ok := a.assembleReader(strings.NewReader(prog), "fake_file.asm")

	// assert
	assert.False(t, ok, "result")
	if assert.Len(t, a.Problems, 11, "problem count") {
		for i := 0; i < len(a.Problems); i++ {
			//one extra error created to inform user of early stop
			assert.EqualValuesf(t, ProblemKindError, a.Problems[i].Kind, "problem kind (index %v)", i)
		}
	}
}

func Test_assembler_Assemble_continueParsingAfter10Warnings(t *testing.T) {
	// arrange
	a := new(Assembler)
	prog := ""
	for i := 0; i < 20; i++ {
		prog = fmt.Sprintf("%vFOO = %#o\n", prog, 0400+i) //generate 19 warnings
	}

	// act
	ok := a.assembleReader(strings.NewReader(prog), "fake_file.asm")

	// assert
	assert.True(t, ok, "result")
	if assert.Len(t, a.Problems, 19, "problem count") {
		for i := 0; i < len(a.Problems); i++ {
			//one extra error created to inform user of early stop
			assert.EqualValuesf(t, ProblemKindWarning, a.Problems[i].Kind, "problem kind (index %v)", i)
		}
	}
}

func Test_assembler_Assemble_unresolvedOperandAbortsBuild(t *testing.T) {
	// arrange
	a := new(Assembler)
	a.setLocation(04000)
	prog := "CA FOOBAR"

	// act
	ok := a.assembleReader(strings.NewReader(prog), "fake_file.asm")

	// assert
	assert.False(t, ok, "result")
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValuesf(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_assembler_Assemble_invalidOperandError(t *testing.T) {
	// arrange
	a := new(Assembler)
	a.setLocation(04000)
	prog := "CA 077777"

	// act
	ok := a.assembleReader(strings.NewReader(prog), "fake_file.asm")

	// assert
	assert.False(t, ok, "result")
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValuesf(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_assembler_Assemble_customCommandEncoder(t *testing.T) {
	// arrange
	a := new(Assembler)
	loc := 04760
	prog := fmt.Sprintf("SETLOC %#o\nNOOP\n", loc)

	// act
	ok := a.assembleReader(strings.NewReader(prog), "fake_file.asm")

	// assert
	assert.True(t, ok, "result")
	assert.Len(t, a.Problems, 0, "problem count")
	assertImageValue(t, uint16(010000|(loc+1)), a, 2, uint16(loc&01777))
}

func assertImageValue(t *testing.T, expected uint16, a *Assembler, bank uint16, offset uint16) {
	if assert.NotNil(t, a.image, "image exists") {
		if b, ok := a.image[bank]; assert.True(t, ok, "bank exists") {
			if assert.Len(t, *b, 1024, "bank lnegth") {
				assert.EqualValues(t, expected, b[offset], "instruction value")
			}
		}
	}
}
