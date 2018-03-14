package memory

import (
	"encoding/binary"

	"github.com/pkg/errors"
)

const (
	erasableBankSize   = 0400
	fixedBankSize      = 02000
	erasableBankCount  = 8
	fixedBankCount     = 32
	fixedSBBankCount   = 8
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
	fixed    [fixedBankCount + fixedSBBankCount]fbank
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
	if address < 0 || address >= totalMemorySize {
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
			if mm.eb < 0 || mm.eb >= erasableBankCount {
				return nil, errors.Errorf("erasable bank %o is out of range", mm.eb)
			}

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
		if mm.fb < 0 || mm.fb >= fixedBankCount {
			return nil, errors.Errorf("fixed bank %o is out of range", mm.eb)
		}

		idx = mm.fb
		if mm.fb >= 030 && mm.sb {
			// the "super-bit" is set so fb 030 - 037 are
			// actually 040 - 047, meaning we need to add 010
			idx += 010
		}
	}
	return mm.fixed[idx][:], nil
}

// Loader is used to stream data into a Main instance.
type Loader struct {
	MM *Main

	leftOver *byte

	pos int
	cb  []uint16
	cbi int
}

func (l *Loader) Write(p []byte) (n int, err error) {
	if l.cb == nil {
		l.cb = l.MM.fixed[0][:]
	}

	pLen := len(p)
	if l.leftOver != nil {
		// we have a leftOver byte from a previous write so go
		// ahead and tack it onto the front of p
		p = append([]byte{*l.leftOver}, p...)
		l.leftOver = nil
	}

	for i := 0; i < len(p); i += 2 {
		if i == len(p)-1 {
			// we need two bytes to make a uint16 but we only
			// have one left, so store it away and hope the
			// next write gives us our second byte
			tmp := p[i]
			l.leftOver = &tmp
			return pLen, nil
		}

		// first check for end of bounds before we write
		if l.pos >= len(l.cb) {
			// we're done with this bank, time to
			// move onto the next one
			l.cbi++
			if l.cbi >= len(l.MM.fixed) {
				// we've reached the end and we have no room
				// left to write in, so we have to error
				return i, errors.Errorf("reached end of memory")
			}

			// now that we've handled bounds check, grab the
			// next bank and rest our pos
			l.cb = l.MM.fixed[l.cbi][:]
			l.pos = 0
		}

		val := binary.BigEndian.Uint16(p[i:])
		l.cb[l.pos] = val >> 1
		l.pos++
	}

	// we wrote all the data without trouble
	return pLen, nil
}
