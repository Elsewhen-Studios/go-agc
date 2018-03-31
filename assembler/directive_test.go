package assembler

import (
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
