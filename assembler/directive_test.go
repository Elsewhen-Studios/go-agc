package assembler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_setLoc_valid(t *testing.T) {
	// arrange
	a := new(assembler)
	a.SetLocation(psudoAddress(04000))

	// act
	ok := a.parseLine("SETLOC 2000")

	// assert
	assert.True(t, ok, "result")
	assert.Zero(t, len(a.problems), "error count")

	assert.NotNil(t, a.location, "location")
	assert.EqualValues(t, 02000, int(*a.location), "location")
}

func Test_setLoc_invalid(t *testing.T) {
	// arrange
	a := new(assembler)
	a.SetLocation(psudoAddress(04000))

	// act
	ok := a.parseLine("SETLOC 14000")

	// assert
	assert.True(t, ok, "result")
	if assert.EqualValues(t, 1, len(a.problems), "problem count") {
		assert.EqualValues(t, problemKindError, a.problems[0].Kind, "problem kind")
	}

	assert.NotNil(t, a.location, "location")
	assert.EqualValues(t, 04000, int(*a.location), "location")
}

func Test_setLoc_noOperand(t *testing.T) {
	// arrange
	a := new(assembler)
	a.SetLocation(psudoAddress(04000))

	// act
	ok := a.parseLine("SETLOC ")

	// assert
	assert.True(t, ok, "result")
	if assert.EqualValues(t, 1, len(a.problems), "problem count") {
		assert.EqualValues(t, problemKindError, a.problems[0].Kind, "problem kind")
	}

	assert.NotNil(t, a.location, "location")
	assert.EqualValues(t, 04000, int(*a.location), "location")
}

func Test_setLoc_unresolvable(t *testing.T) {
	// arrange
	a := new(assembler)
	a.Init()
	a.SetLocation(psudoAddress(04000))

	// act
	ok := a.parseLine("SETLOC FOO")

	// assert
	assert.True(t, ok, "result")
	if assert.EqualValues(t, 1, len(a.problems), "problem count") {
		assert.EqualValues(t, problemKindError, a.problems[0].Kind, "problem kind")
	}

	assert.NotNil(t, a.location, "location")
	assert.EqualValues(t, 04000, int(*a.location), "location")
}
