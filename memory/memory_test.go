package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErasableSequentialWrite(t *testing.T) {
	// arrange
	var mm Main

	// act
	for a := 0; a < startOfFixedMemory; a++ {
		assert.NoError(t, mm.Write(a, uint16(a)))
	}

	// assert
	// in a default setup, writing sequentially to all erasable
	// memory will fill E0, E1, E2, and E0 again
	for b := 1; b <= 2; b++ {
		for i := 0; i < erasableBankSize; i++ {
			assert.Equal(t, uint16(b*erasableBankSize+i), mm.erasable[b][i], "bank E%d[%d]", b, i)
		}
	}

	// since E0 is mapped to both 0000 - 0377 and 1400 - 1777 the
	// sequential write will result in writing it twice, so it
	// will end up with the second values
	for i := 0; i < erasableBankSize; i++ {
		assert.Equal(t, uint16(01400+i), mm.erasable[0][i], "bank E0[%d]", i)
	}
}

func TestErasableBankSelection(t *testing.T) {
	// arrange
	var mm Main

	// act
	for b := 0; b < erasableBankCount; b++ {
		// switch to bank 'b'
		mm.eb = b
		// now fill bank 'b' by writing to addresses 01400 - 01777
		// which correspond to the switchable addresses
		for i := 0; i < erasableBankSize; i++ {
			assert.NoError(t, mm.Write(01400+i, uint16(b+100)))
		}
	}

	// assert
	// now each bank should be filled with its
	// bank index + 100
	for b := 0; b < erasableBankCount; b++ {
		assertBank(t, b, mm.erasable[b][:], uint16(b+100))
	}
}

func TestSequentialRead(t *testing.T) {
	// arrange
	var mm Main
	mm.eb = 3
	for b := 0; b <= 4; b++ {
		for i := 0; i < erasableBankSize; i++ {
			mm.erasable[b][i] = uint16(b*erasableBankSize + i)
		}
	}

	mm.fb = 1
	for b := 0; b < 3; b++ {
		for i := 0; i < fixedBankSize; i++ {
			mm.fixed[b+1][i] = uint16(b*fixedBankSize + i + startOfFixedMemory)
		}
	}

	// act (and assert)
	for i := 0; i < totalMemorySize; i++ {
		val, err := mm.Read(i)
		assert.NoError(t, err)
		assert.Equal(t, uint16(i), val, "address %o", i)
	}
}

func TestRoundTripStrips16thBit(t *testing.T) {
	var mm Main

	// write 16 set bits out to memory and ensure when
	// we read it back it is only 15 bits
	assert.NoError(t, mm.Write(123, 0xFFFF), "write")
	val, err := mm.Read(123)
	assert.NoError(t, err, "read")
	assert.Equal(t, uint16(0x7FFF), val, "read")
}

func TestReadPastEndOfMemory(t *testing.T) {
	var mm Main
	_, err := mm.Read(totalMemorySize + 1)
	assert.Error(t, err)
}

func TestWritePastEndOfMemory(t *testing.T) {
	var mm Main
	err := mm.Write(totalMemorySize+1, 123)
	assert.Error(t, err)
}

func TestWriteToFixedMemory(t *testing.T) {
	var mm Main
	err := mm.Write(startOfFixedMemory+1, 123)
	assert.Error(t, err)
}

func fillBank(b []uint16, val uint16) {
	for i := 0; i < len(b); i++ {
		b[i] = val
	}
}

func assertBank(t *testing.T, bidx int, b []uint16, expected uint16) {
	for i := 0; i < erasableBankSize; i++ {
		assert.Equal(t, expected, b[i], "bank E%d[%d]", bidx, i)
	}
}
