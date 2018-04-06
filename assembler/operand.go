package assembler

import (
	"errors"
	"unicode"
)

func requireOperand(sp *scannerPeeker, p *instructionParams) error {
	t, ok := sp.Consume()
	if !ok {
		return errors.New("operand expected but none found")
	}
	p.operandToken = t
	return nil
}

type tokenKind int

const (
	octalToken = tokenKind(iota)
	decimalToken
	symbolToken
)

const (
	lastChannel = 00777

	lastUnswitchedErasableAddress = 01377
	lastErasableAddress           = 01777

	lastswitchedFixedAddress = 03777
	lastFixedAddress         = 07777
)

func clasifyToken(token string) tokenKind {
	var decNum bool

	for _, char := range token {
		if unicode.IsDigit(char) {
			if char >= '8' {
				decNum = true
			}
		} else {
			return symbolToken
		}
	}

	if decNum {
		return decimalToken
	}

	return octalToken
}

type operandValidator func(val uint16, p *instructionParams) bool

func validateTCOperand(val uint16, p *instructionParams) bool {
	if ok := requireAnyMemoryOperand(val, p); !ok {
		return ok
	} else if val == 00003 || val == 00004 || val == 00006 {
		p.logger.LogErrorf("%v is not a valid operand for %v", p.operandToken, p.instToken)
		return false
	}
	return true
}

func validateINDEXOperand(val uint16, p *instructionParams) bool {
	if ok := requireErasable(val, p); !ok {
		return ok
	} else if val == 00017 {
		p.logger.LogErrorf("%v is not a valid operand for %v", p.operandToken, p.instToken)
		return false
	}
	return true
}

func requireAnyMemoryOperand(val uint16, p *instructionParams) bool {
	if val > lastFixedAddress {
		p.logger.LogErrorf("%v is not a valid memory address", p.operandToken)
		return false
	}
	return true
}

func requireDoubleAnyMemoryOperand(val uint16, p *instructionParams) bool {
	if val > lastFixedAddress-1 {
		p.logger.LogErrorf("%v is not a valid double precision memory address", p.operandToken)
		return false
	}
	if val == lastswitchedFixedAddress {
		p.logger.LogWarningf("double precision at %v crosses the switchable/unswitchable fixed memory boundary", p.operandToken)
	}
	if val == lastErasableAddress {
		p.logger.LogWarningf("double precision at %v crosses the erasable/fixed memory boundary", p.operandToken)
	}
	if val == lastUnswitchedErasableAddress {
		p.logger.LogWarningf("double precision at %v crosses the unswitchable/switchable erasable memory boundary", p.operandToken)
	}
	return true
}

func requireErasable(val uint16, p *instructionParams) bool {
	if val > lastErasableAddress {
		p.logger.LogErrorf("%v is not a valid erasable memory address", p.operandToken)
		return false
	}
	return true
}

func requireDoubleErasable(val uint16, p *instructionParams) bool {
	if val > lastErasableAddress-1 {
		p.logger.LogErrorf("%v is not a valid double precision erasable memory address", p.operandToken)
		return false
	}
	if val == lastUnswitchedErasableAddress {
		p.logger.LogWarningf("double precision at %v crosses the unswitchable/switchable erasable memory boundary", p.operandToken)
	}
	return true
}

func requireChannel(val uint16, p *instructionParams) bool {
	if val > lastChannel {
		p.logger.LogErrorf("%v is not a valid I/O channel address", p.operandToken)
		return false
	}
	return true
}

func requireFixed(val uint16, p *instructionParams) bool {
	if val < 02000 || val > lastFixedAddress {
		p.logger.LogErrorf("%v is not a valid fixed memory address", p.operandToken)
		return false
	}
	return true
}
