package assembler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_noopEncoder_validErasable(t *testing.T) {
	// arrange
	a, pl := buildAssemblerLogger()
	p := &instructionParams{logger: pl, location: psudoAddress(0456)} //bank E1

	// act
	mc, ok := noopEncoder(p)

	// assert
	assert.True(t, ok, "result")
	assert.Zero(t, len(a.problems), "problem count")

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
	assert.Zero(t, len(a.problems), "problem count")

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
	if assert.EqualValues(t, 1, len(a.problems), "problem count") {
		assert.EqualValues(t, problemKindError, a.problems[0].Kind, "problem kind")
	}
}
