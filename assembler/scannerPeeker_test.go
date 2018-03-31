package assembler

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newScannerPeeker(t *testing.T) {
	// arrange
	s := bufio.NewScanner(strings.NewReader("test string"))

	// act
	sp := newScannerPeeker(s)

	// assert
	assert.Equal(t, s, sp.scanner, "scanner")
	assert.False(t, sp.primed, "primed")
}

func Test_scannerPeeker_firstPeek(t *testing.T) {
	// arrange
	s := bufio.NewScanner(strings.NewReader("first second"))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	token, ok := sp.Peek()

	// assert
	assert.True(t, ok, "ok")
	assert.EqualValues(t, token, "first")
}

func Test_scannerPeeker_secondPeek(t *testing.T) {
	// arrange
	s := bufio.NewScanner(strings.NewReader("first second"))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)
	token, ok := sp.Peek()

	// act
	token, ok = sp.Peek()

	// assert
	assert.True(t, ok, "ok")
	assert.EqualValues(t, token, "first")
}

func Test_scannerPeeker_firstConsume(t *testing.T) {
	// arrange
	s := bufio.NewScanner(strings.NewReader("first second"))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	// act
	token, ok := sp.Consume()

	// assert
	assert.True(t, ok, "ok")
	assert.EqualValues(t, token, "first")
}

func Test_scannerPeeker_secondConsume(t *testing.T) {
	// arrange
	s := bufio.NewScanner(strings.NewReader("first second"))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)
	token, ok := sp.Consume()

	// act
	token, ok = sp.Consume()

	// assert
	assert.True(t, ok, "ok")
	assert.EqualValues(t, token, "second")
}
