package memory

import (
	"github.com/pkg/errors"
)

const (
	erasableBankSize   = 0400
	fixedBankSize      = 02000
	erasableBankCount  = 8
	fixedBankCount     = 36
	startOfFixedMemory = 02000
	totalMemorySize    = 010000
	wordMask           = 077777
)

// test

type ebank [erasableBankSize]uint16
type fbank [fixedBankSize]uint16
type bank []uint16

// Main represents the full addressable main memory of the AGC.
type Main struct {
	erasable [erasableBankCount]ebank
	fixed    [fixedBankCount]fbank
	eb       int
	fb       int
	sb       bool
}

// Read gets a 15 bit word from a specified address in main memory.
func (mm *Main) Read(address int) (uint16, error) {
	b, err := mm.selectBank(address)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get bank for address %o", address)
	}

	return b[address%len(b)], nil
}

// Write stores a 15 bit word to a specified address in main memory.
func (mm *Main) Write(address int, val uint16) error {
	b, err := mm.selectBank(address)
	if err != nil {
		return errors.Wrapf(err, "failed to get bank for address %o", address)
	}

	if len(b) == fixedBankSize {
		return errors.Errorf("address %o is fixed and cannot be written", address)
	}

	// only store 15 bits because the 16th bit is used for parity
	// in the hardware, here we just discard it
	b[address%len(b)] = val & wordMask
	return nil
}

func (mm *Main) selectBank(address int) (bank, error) {
	if address >= totalMemorySize {
		return nil, errors.Errorf("address %o is out of range", address)
	}

	if address < startOfFixedMemory {
		// erasable memory
		// 00000 - 00377 -> erasable[0]
		// 00400 - 00777 -> erasable[1]
		// 01000 - 01377 -> erasable[2]
		// 01400 - 01777 -> erasable[eb]
		idx := address / erasableBankSize
		if idx == 3 {
			idx = mm.eb
		}
		return mm.erasable[idx][:], nil
	}

	// fixed memory
	// 02000 - 03777 -> fixed[fb,sb]
	// 04000 - 05777 -> fixed[2]
	// 06000 - 07777 -> fixed[3]
	idx := (address-startOfFixedMemory)/fixedBankSize + 1
	if idx == 1 {
		idx = mm.fb
		if mm.fb >= 030 && mm.sb {
			// the "super-bit" is set so fb 030 - 037 are
			// actually 040 - 047, meaning we need to add 010
			idx += 010
		}
	}
	return mm.fixed[idx][:], nil
}
