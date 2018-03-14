package memory

import (
	"bytes"
	"encoding/binary"
	"io"
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
		assertErasableBank(t, b, mm.erasable[b], uint16(b*erasableBankSize))
	}

	// since E0 is mapped to both 0000 - 0377 and 1400 - 1777 the
	// sequential write will result in writing it twice, so it
	// will end up with the second values
	assertErasableBank(t, 0, mm.erasable[0], uint16(3*erasableBankSize))
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
			assert.NoError(t, mm.Write(01400+i, uint16(b+100+i)))
		}
	}

	// assert
	// now each bank should be filled with its
	// bank index + 100
	for b := 0; b < erasableBankCount; b++ {
		assertErasableBank(t, b, mm.erasable[b], uint16(b+100))
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
	//Note: This test bypasses the erasable bank 0 address exceptions which are only applied through Write()
	for i := 0; i < totalMemorySize; i++ {
		val, err := mm.Read(i)
		assert.NoError(t, err)
		assert.Equal(t, uint16(i), val, "address %o", i)
	}
}

func TestFixedBankSelection(t *testing.T) {
	//arrange
	var mm Main
	for b := 0; b < fixedBankCount+fixedSBBankCount; b++ {
		for i := 0; i < fixedBankSize; i++ {
			mm.fixed[b][i] = uint16(b + 100 + i)
		}
	}

	// act (and assert)
	mm.sb = false
	for b := 0; b < fixedBankCount; b++ {
		mm.fb = b
		for i := 0; i < fixedBankSize; i++ {
			val, err := mm.Read(startOfFixedMemory + i)
			assert.NoError(t, err)
			assert.Equal(t, uint16(b+100+i), val, "address %o", i)
		}
	}
	mm.sb = true
	for b := 0; b < fixedBankCount; b++ {
		mm.fb = b
		for i := 0; i < fixedBankSize; i++ {
			val, err := mm.Read(startOfFixedMemory + i)
			assert.NoError(t, err)
			expected := b + 100 + i
			if b >= fixedBankCount-fixedSBBankCount {
				expected += fixedSBBankCount
			}
			assert.Equal(t, uint16(expected), val, "address %o", i)
		}
	}
}

func TestRoundTripOverflowCorrection(t *testing.T) {
	var mm Main

	t.Run("positive overflow", func(t *testing.T) {
		// Write an overflowed word (16th bit != 15th bit)
		// and verify that it is overflow corrected.
		assert.NoError(t, mm.Write(123, 17318), "write")
		val, err := mm.Read(123)
		assert.NoError(t, err, "read")
		assert.Equal(t, uint16(934), val, "read")
	})

	t.Run("negative overflow", func(t *testing.T) {
		// Now test a negative overflow (an underflow)
		// Note: the AGC uses 1's complement so the MSB
		// is actually a sign bit (48217 = -17318)
		assert.NoError(t, mm.Write(124, 48217), "write")
		val, err := mm.Read(124)
		assert.NoError(t, err, "read")
		assert.Equal(t, uint16(64601), val, "read")
	})
}

func TestReadOutOfRange(t *testing.T) {
	var (
		mm  Main
		err error
	)

	_, err = mm.Read(totalMemorySize + 1)
	assert.Error(t, err)

	_, err = mm.Read(-1)
	assert.Error(t, err)
}

func TestWriteOutOfRange(t *testing.T) {
	var (
		mm  Main
		err error
	)

	err = mm.Write(totalMemorySize, 123)
	assert.Error(t, err)

	err = mm.Write(-1, 123)
	assert.Error(t, err)
}

func TestWriteToFixedMemory(t *testing.T) {
	var mm Main
	err := mm.Write(startOfFixedMemory, 123)
	assert.Error(t, err)
}

func TestBankOutOfRange(t *testing.T) {
	var (
		mm  Main
		err error
	)

	mm.eb = -1
	_, err = mm.Read(erasableBankSize * 3)
	assert.Error(t, err)

	mm.eb = erasableBankCount + 1
	_, err = mm.Read(erasableBankSize * 3)
	assert.Error(t, err)

	mm.fb = -1
	_, err = mm.Read(startOfFixedMemory)
	assert.Error(t, err)

	mm.fb = fixedBankCount + 1
	_, err = mm.Read(startOfFixedMemory)
	assert.Error(t, err)
}

/*
func fillBank(b []uint16, val uint16) {
	for i := 0; i < len(b); i++ {
		b[i] = val
	}
}
*/

func assertErasableBank(t *testing.T, bidx int, b ebank, baseValue uint16) {
	for i := 0; i < erasableBankSize; i++ {
		assert.Equal(t, baseValue+uint16(i), b[i], "bank E%d[%d]", bidx, i)
	}
}

func TestLoader(t *testing.T) {
	// arrange
	var mm Main
	l := &Loader{MM: &mm}

	buf := new(bytes.Buffer)
	for b := 0; b < fixedBankCount+fixedSBBankCount; b++ {
		for i := 0; i < fixedBankSize; i++ {
			buf.Write([]byte{0x1E, byte(b) << 1})
		}
	}
	expectedLen := int64(buf.Len())

	// act
	n, err := io.Copy(l, buf)

	// assert
	assert.Equal(t, expectedLen, n, "bytes copied")
	assert.NoError(t, err, "io.Copy error")

	for b := 0; b < fixedBankCount+fixedSBBankCount; b++ {
		for i := 0; i < fixedBankSize; i++ {
			assert.Equal(t, uint16(0xF00+b), mm.fixed[b][i], "F%d[%d]", b, i)
		}
	}
}

func TestLoader_SplitWrites(t *testing.T) {
	// arrange
	var mm Main
	l := &Loader{MM: &mm}
	raw := make([]byte, fixedBankSize*2)
	for i := 0; i < fixedBankSize; i++ {
		binary.BigEndian.PutUint16(raw[i*2:], uint16(i)<<1)
	}

	// act
	for i := 0; i < len(raw); {
		for j := 1; j <= 3 && i+j <= len(raw); j++ {
			n, err := l.Write(raw[i : i+j])
			assert.Equal(t, j, n, "i:%d j:%d bytes copied", i, j)
			assert.NoError(t, err, "i:%d j:%d error", i, j)
			i += j
		}
	}

	// assert
	for i := 0; i < fixedBankSize; i++ {
		assert.Equal(t, uint16(i), mm.fixed[0][i], "F0[%d]", i)
	}
}
