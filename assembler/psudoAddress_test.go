package assembler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_psudoAddress_isValid(t *testing.T) {
	// arrange

	// act

	// assert
	for i := 0x0000; i < 0x1800; i++ {
		assert.Truef(t, psudoAddress(i).isValid(), "psudo-address %o", i)
	}
	for i := 0x1800; i < 0x2000; i++ {
		assert.Falsef(t, psudoAddress(i).isValid(), "psudo-address %o", i)
	}
	for i := 0x2000; i < 0xB000; i++ {
		assert.Truef(t, psudoAddress(i).isValid(), "psudo-address %o", i)
	}
	for i := 0xB000; i <= 0xFFFF; i++ {
		assert.Falsef(t, psudoAddress(i).isValid(), "psudo-address %o", i)
	}
}

func Test_psudoAddress_isErasable(t *testing.T) {
	// arrange

	// act

	// assert
	for i := 0x0000; i < 0x0800; i++ {
		assert.Truef(t, psudoAddress(i).isErasable(), "psudo-address %o", i)
	}
	for i := 0x0800; i < 0xFFFF; i++ {
		assert.Falsef(t, psudoAddress(i).isErasable(), "psudo-address %o", i)
	}
}

func Test_psudoAddress_isFixed(t *testing.T) {
	// arrange

	// act

	// assert
	for i := 0x0000; i < 0x0800; i++ {
		assert.Falsef(t, psudoAddress(i).isFixed(), "psudo-address %o", i)
	}
	for i := 0x0800; i < 0x1800; i++ {
		assert.Truef(t, psudoAddress(i).isFixed(), "psudo-address %o", i)
	}
	for i := 0x1800; i < 0x2000; i++ {
		assert.Falsef(t, psudoAddress(i).isFixed(), "psudo-address %o", i)
	}
	for i := 0x2000; i < 0xB000; i++ {
		assert.Truef(t, psudoAddress(i).isFixed(), "psudo-address %o", i)
	}
	for i := 0xB000; i < 0xFFFF; i++ {
		assert.Falsef(t, psudoAddress(i).isFixed(), "psudo-address %o", i)
	}
}

func Test_psudoAddress_isSuperBit(t *testing.T) {
	// arrange

	// act

	// assert
	for i := 0x0000; i < 0x9000; i++ {
		assert.Falsef(t, psudoAddress(i).isSuperBit(), "psudo-address %o", i)
	}
	for i := 0x9000; i < 0xB000; i++ {
		assert.Truef(t, psudoAddress(i).isSuperBit(), "psudo-address %o", i)
	}
	for i := 0xB000; i < 0xFFFF; i++ {
		assert.Falsef(t, psudoAddress(i).isSuperBit(), "psudo-address %o", i)
	}
}

func Test_psudoAddress_bankAndOffset(t *testing.T) {
	// arrange
	bankSets := []struct {
		paAddr    uint16
		erasable  bool
		startBank int
		bankCount int
		len       int
	}{
		{paAddr: 0x0000, erasable: true, startBank: 0, bankCount: 8, len: 256},
		{paAddr: 0x1000, erasable: false, startBank: 0, bankCount: 2, len: 1024},
		{paAddr: 0x0800, erasable: false, startBank: 2, bankCount: 2, len: 1024},
		{paAddr: 0x2000, erasable: false, startBank: 4, bankCount: 36, len: 1024},
	}

	// act

	// assert
	for _, bankSet := range bankSets {
		pa := psudoAddress(bankSet.paAddr)
		for bank := bankSet.startBank; bank < bankSet.startBank+bankSet.bankCount; bank++ {
			for i := 0; i < bankSet.len; i++ {
				e, b, o := pa.bankAndOffset()

				assert.EqualValuesf(t, bankSet.erasable, e, "psudo-address %o (erasable)", int(pa))
				assert.EqualValuesf(t, bank, b, "psudo-address %o (bank)", int(pa))
				assert.EqualValuesf(t, i, o, "psudo-address %o (offset)", int(pa))
				pa++
			}
		}
	}
}

func Test_psudoAddress_asOperand(t *testing.T) {
	// arrange
	bankSets := []struct {
		paAddr uint16
		opAddr uint16
		len    int
		reps   int
	}{
		//Erasable
		{paAddr: 0x0000, opAddr: 00000, len: 256, reps: 1},
		{paAddr: 0x0100, opAddr: 00400, len: 256, reps: 1},
		{paAddr: 0x0200, opAddr: 01000, len: 256, reps: 1},
		{paAddr: 0x0300, opAddr: 01400, len: 256, reps: 5},
		//Fixed
		{paAddr: 0x0800, opAddr: 04000, len: 1024, reps: 1},
		{paAddr: 0x0C00, opAddr: 06000, len: 1024, reps: 1},
		{paAddr: 0x1000, opAddr: 02000, len: 1024, reps: 2},
		{paAddr: 0x2000, opAddr: 02000, len: 1024, reps: 36},
	}

	// act

	// assert
	for _, bankSet := range bankSets {
		pa := psudoAddress(bankSet.paAddr)
		for rep := 0; rep < bankSet.reps; rep++ {
			for i := 0; i < bankSet.len; i++ {
				op := pa.asOperand()

				assert.EqualValuesf(t, int(bankSet.opAddr)+i, op, "psudo-address %o", int(pa))
				pa++
			}
		}
	}
}

func Test_psudoAddress_asBankedOperand(t *testing.T) {
	// arrange
	bankSets := []struct {
		paAddr uint16
		opAddr uint16
		len    int
		reps   int
	}{
		//Erasable
		{paAddr: 0x0000, opAddr: 01400, len: 256, reps: 8},
		//Fixed
		{paAddr: 0x0800, opAddr: 02000, len: 1024, reps: 4},
		{paAddr: 0x2000, opAddr: 02000, len: 1024, reps: 36},
	}

	// act

	// assert
	for _, bankSet := range bankSets {
		pa := psudoAddress(bankSet.paAddr)
		for rep := 0; rep < bankSet.reps; rep++ {
			for i := 0; i < bankSet.len; i++ {
				op := pa.asBankedOperand()

				assert.EqualValuesf(t, int(bankSet.opAddr)+i, op, "psudo-address %o", int(pa))
				pa++
			}
		}
	}
}

func Test_psudoAddress_nextValid(t *testing.T) {
	//erasable
	pa := testPsudoAddressSequence(t, psudoAddress(0x0000), 8*256-1)
	_, err := pa.nextValid()
	assert.Errorf(t, err, "psudo-address %o", int(pa))

	//fixed
	pa = testPsudoAddressSequence(t, psudoAddress(0x0800), 4*1024-1)
	pa, err = pa.nextValid()
	assert.EqualValuesf(t, 0x2000, int(pa), "psudo-address %o", int(pa))

	pa = testPsudoAddressSequence(t, pa, 36*1024-1)
	pa, err = pa.nextValid()
	assert.Errorf(t, err, "psudo-address %o", int(pa))
}

func testPsudoAddressSequence(t *testing.T, pa psudoAddress, count int) psudoAddress {
	for i := 0x0000; i < count; i++ {
		// act
		next, err := pa.nextValid()
		// assert
		assert.NoErrorf(t, err, "psudo-address %o", int(pa))
		assert.EqualValuesf(t, int(pa)+1, int(next), "psudo-address %o", int(pa))

		pa = next
	}
	return pa
}

func Test_psudoAddress_isBeginingOfSwitchableBank(t *testing.T) {
	// arrange
	swbanks := []uint16{
		//Erasable
		0x0300, 0x0400, 0x0500, 0x0600, 0x0700,
		//Fixed
		0x1000, 0x1400, 0x2000, 0x2400, 0x2800,
		0x2C00, 0x3000, 0x3400, 0x3800, 0x3C00,
		0x4000, 0x4400, 0x4800, 0x4C00, 0x5000,
		0x5400, 0x5800, 0x5C00, 0x6000, 0x6400,
		0x6800, 0x6C00, 0x7000, 0x7400, 0x7800,
		0x7C00, 0x8000, 0x8400, 0x8800, 0x8C00,
		0x9000, 0x9400, 0x9800, 0x9C00, 0xA000,
		0xA400, 0xA800, 0xAC00,
	}

loop:
	for i := 0x0000; i < 0xFFFF; i++ {
		pa := psudoAddress(i)
		if pa.isValid() {
			// act
			result := pa.isBeginingOfSwitchableBank()

			// assert
			for _, addr := range swbanks {
				if uint16(i) == addr {
					assert.Truef(t, result, "psudo-address %o", int(pa))
					continue loop
				}
			}
			assert.Falsef(t, result, "psudo-address %o", int(pa))
		}
	}
}
