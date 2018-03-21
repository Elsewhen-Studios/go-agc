package assembler

import "errors"

type psudoAddress uint16

const (
	erasableBankSize = 00400 //256
	erasableBankMask = erasableBankSize - 1
	erasableBankBits = 8

	fixedBankSize = 02000 //1024
	fixedBankMask = fixedBankSize - 1
	fixedBankBits = 10

	startOfUnE       = 0000000
	startOfSwE       = 0001400
	startOfFixed     = 0004000
	startOfGap       = 0014000
	startOfFixedCont = 0020000
	startOfSuperBit  = 0110000
	startOfExtraSB   = 0120000
	startOfUnused    = 0130000
)

func (p psudoAddress) isValid() bool {
	return p < startOfGap || (p >= startOfFixedCont && p < startOfUnused)
}

func (p psudoAddress) isErasable() bool {
	return p < startOfFixed
}

func (p psudoAddress) isFixed() bool {
	return (p >= startOfFixed && p < startOfGap) || (p >= startOfFixedCont && p < startOfUnused)
}

func (p psudoAddress) isSuperBit() bool {
	return p >= startOfSuperBit && p < startOfUnused
}

func (p psudoAddress) bankAndOffset() (erasable bool, bank uint16, offset uint16) {
	if p.isErasable() {
		erasable = true
		bank = uint16(p) >> erasableBankBits
		offset = uint16(p) & erasableBankMask
		return
	}

	erasable = false
	bank = uint16(p) >> fixedBankBits
	if bank >= 4 {
		bank -= 4
	}
	offset = uint16(p) & fixedBankMask
	return
}

func (p psudoAddress) asOperand() uint16 {
	erasable, bank, offset := p.bankAndOffset()

	if erasable {
		if bank >= 4 {
			bank = 3
		}
		return (bank << erasableBankBits) | offset
	}

	if bank < 2 || bank >= 4 {
		bank = 1
	}
	return (bank << fixedBankBits) | offset
}

func (p psudoAddress) asBankedOperand() uint16 {
	if p.isErasable() {
		return (3 << erasableBankBits) | (uint16(p) & erasableBankMask)
	}

	return (1 << fixedBankBits) | (uint16(p) & fixedBankMask)
}

func (p psudoAddress) nextValid() (psudoAddress, error) {
	p++

	switch p {
	//erasable
	case startOfFixed:
		return 0xFFFF, errors.New("End of Erasable Memory")
		//fixed
	case startOfGap:
		return startOfFixedCont, nil
	case startOfUnused:
		return 0xFFFF, errors.New("End of Fixed Memory")
	default:
		return p, nil
	}
}

func (p psudoAddress) isBeginingOfSwitchableBank() bool {
	erasable, bank, offset := p.bankAndOffset()
	if erasable {
		return (bank > 2) && (offset == 0)
	}

	return (bank < 2 || bank > 3) && (offset == 0)
}
