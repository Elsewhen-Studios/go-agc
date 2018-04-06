package assembler

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_instAliases(t *testing.T) {
	// arrange

	// act

	// assert
	for k, v := range instAliases {
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

func Test_findInstruction_standard(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl, extended: false, instToken: "CA"}

	// act
	inst := findInstruction(p)

	// assert
	assert.NotNil(t, inst, "return value")
	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_findInstruction_standardWhileExtended(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl, extended: true, instToken: "CA"}

	// act
	inst := findInstruction(p)

	// assert
	assert.NotNil(t, inst, "return value")
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_findInstruction_extended(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl, extended: true, instToken: "DCA"}

	// act
	inst := findInstruction(p)

	// assert
	assert.NotNil(t, inst, "return value")
	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_findInstruction_extendedWhileStandard(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl, extended: false, instToken: "DCA"}

	// act
	inst := findInstruction(p)

	// assert
	assert.NotNil(t, inst, "return value")
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_findInstruction_notFound(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl, extended: false, instToken: "FOO"}

	// act
	inst := findInstruction(p)

	// assert
	assert.Nil(t, inst, "return value")
	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_noopEncoder_validErasable(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl, location: psudoAddress(0456)} //bank E1

	// act
	mc, ok := noopEncoder(p)

	// assert
	assert.True(t, ok, "result")
	assert.Len(t, a.Problems, 0, "problem count")

	assert.EqualValues(t, 030000, mc, "machine code")
}

func Test_noopEncoder_validFixed(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl, location: psudoAddress(022022)} //bank F5

	// act
	mc, ok := noopEncoder(p)

	// assert
	assert.True(t, ok, "result")
	assert.Len(t, a.Problems, 0, "problem count")

	assert.EqualValues(t, 012023, mc, "machine code")
}

func Test_noopEncoder_invalidEndofMemory(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl, location: psudoAddress(0xAFFF)} //bank F31+SB

	// act
	_, ok := noopEncoder(p)

	// assert
	assert.False(t, ok, "result")
	if assert.Len(t, a.Problems, 1, "problem count") {
		assert.EqualValues(t, ProblemKindError, a.Problems[0].Kind, "problem kind")
	}
}

func Test_DXCH_encoding(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	op := 00060
	p := &instructionParams{logger: pl, resolver: a, extended: false, instToken: "DXCH", operandToken: fmt.Sprintf("%#o", op)}
	i := findInstruction(p)
	require.NotNil(t, i, "instruction lookup")
	require.Len(t, a.Problems, 0, "lookup problem count")

	// act
	mc, ok := i.encode(p)

	// assert
	assert.True(t, ok, "result")
	assert.EqualValues(t, 052001+op, mc, "machine code")
	assert.Len(t, a.Problems, 0, "problem count")
}

func Test_encode(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	tests := []struct {
		inst string
		ext  bool
		op   int
		exp  uint16
	}{
		//00
		{inst: "TC", ext: false, op: 02345, exp: 002345},
		{inst: "XXALQ", ext: false, exp: 000000},
		{inst: "XLQ", ext: false, exp: 000001},
		{inst: "RETURN", ext: false, exp: 000002},
		{inst: "RELINT", ext: false, exp: 000003},
		{inst: "INHINT", ext: false, exp: 000004},
		{inst: "EXTEND", ext: false, exp: 000006},
		//01
		{inst: "CCS", ext: false, op: 01234, exp: 011234},
		{inst: "TCF", ext: false, op: 02345, exp: 012345},
		//02
		{inst: "DAS", ext: false, op: 01234, exp: 021235},
		{inst: "DDOUBL", ext: false, exp: 020001},
		{inst: "LXCH", ext: false, op: 01234, exp: 023234},
		{inst: "ZL", ext: false, exp: 022007},
		{inst: "INCR", ext: false, op: 01234, exp: 025234},
		{inst: "ADS", ext: false, op: 01234, exp: 027234},
		//03
		{inst: "CA", ext: false, op: 02345, exp: 032345},
		//04
		{inst: "CS", ext: false, op: 02345, exp: 042345},
		//05
		{inst: "INDEX", ext: false, op: 01234, exp: 051234},
		{inst: "RESUME", ext: false, exp: 050017},
		{inst: "DXCH", ext: false, op: 01234, exp: 053235},
		{inst: "DTCF", ext: false, exp: 052005},
		{inst: "DTCB", ext: false, exp: 052006},
		{inst: "TS", ext: false, op: 01234, exp: 055234},
		{inst: "OVSK", ext: false, exp: 054000},
		{inst: "TCAA", ext: false, exp: 054005},
		{inst: "XCH", ext: false, op: 01234, exp: 057234},
		//06
		{inst: "AD", ext: false, op: 02345, exp: 062345},
		{inst: "DOUBLE", ext: false, exp: 060000},
		//07
		{inst: "MASK", ext: false, op: 02345, exp: 072345},
		//10
		{inst: "READ", ext: true, op: 0123, exp: 000123},
		{inst: "WRITE", ext: true, op: 0123, exp: 001123},
		{inst: "RAND", ext: true, op: 0123, exp: 002123},
		{inst: "WAND", ext: true, op: 0123, exp: 003123},
		{inst: "ROR", ext: true, op: 0123, exp: 004123},
		{inst: "WOR", ext: true, op: 0123, exp: 005123},
		{inst: "RXOR", ext: true, op: 0123, exp: 006123},
		{inst: "EDRUPT", ext: true, op: 0123, exp: 007123},
		//11
		{inst: "DV", ext: true, op: 01234, exp: 011234},
		{inst: "BZF", ext: true, op: 02345, exp: 012345},
		//12
		{inst: "MSU", ext: true, op: 01234, exp: 021234},
		{inst: "QXCH", ext: true, op: 01234, exp: 023234},
		{inst: "ZQ", ext: true, exp: 022007},
		{inst: "AUG", ext: true, op: 01234, exp: 025234},
		{inst: "DIM", ext: true, op: 01234, exp: 027234},
		//13
		{inst: "DCA", ext: true, op: 02345, exp: 032346},
		//14
		{inst: "DCS", ext: true, op: 02345, exp: 042346},
		{inst: "DCOM", ext: true, exp: 040001},
		//15
		{inst: "INDEX", ext: true, op: 02345, exp: 052345},
		//16
		{inst: "SU", ext: true, op: 01234, exp: 061234},
		{inst: "BZMF", ext: true, op: 02345, exp: 062345},
		//17
		{inst: "MP", ext: true, op: 02345, exp: 072345},
		{inst: "SQUARE", ext: true, exp: 070000},
	}

	for _, test := range tests {
		p := &instructionParams{logger: pl, resolver: a, extended: test.ext, instToken: test.inst, operandToken: fmt.Sprintf("%#o", test.op)}
		i := findInstruction(p)
		require.NotNilf(t, i, "instruction lookup (%v)", test.inst)
		probCount := len(a.Problems)

		// act
		mc, ok := i.encode(p)

		// assert
		assert.Truef(t, ok, "result (%v)", test.inst)
		assert.EqualValuesf(t, test.exp, mc, "machine code (%v)", test.inst)
		assert.Lenf(t, a.Problems, probCount, "problem count (%v)", test.inst)
	}
}

func Test_encode_noop(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	tests := []struct {
		loc psudoAddress
		exp uint16
	}{
		{loc: 01234, exp: 030000},
		{loc: 04567, exp: 014570},
	}

	for _, test := range tests {
		p := &instructionParams{logger: pl, resolver: a, extended: false, instToken: "NOOP", location: test.loc}
		i := findInstruction(p)
		require.NotNilf(t, i, "instruction lookup (%#o)", test.loc)
		probCount := len(a.Problems)

		// act
		mc, ok := i.encode(p)

		// assert
		assert.Truef(t, ok, "result (%#o)", test.loc)
		assert.EqualValuesf(t, test.exp, mc, "machine code (%#o)", test.loc)
		assert.Lenf(t, a.Problems, probCount, "problem count (%#o)", test.loc)
	}
}
