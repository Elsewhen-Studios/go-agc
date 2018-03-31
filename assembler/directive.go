package assembler

import (
	"strconv"
)

type directiveHandler func(a *Assembler, sp *scannerPeeker, p *instructionParams) bool

var directives = map[string]directiveHandler{
	"SETLOC": setLoc,
	//	"DEC":    dec,
	"OCT": oct,
}

func setLoc(a *Assembler, sp *scannerPeeker, p *instructionParams) bool {
	if err := requireOperand(sp, p); err != nil {
		p.logger.LogError(err.Error())
		return false
	}

	val, err := p.resolveOperand()
	if err != nil {
		p.logger.LogError(err.Error())
		return false
	}

	newLoc := psudoAddress(val)
	if !newLoc.isValid() {
		p.logger.LogErrorf("%v is not a valid psudo-address", p.operandToken)
		return false
	}

	action := func(a *Assembler) bool {
		a.setLocation(newLoc)
		return true
	}
	action(a)
	a.queueOperation(action)

	return true
}

func oct(a *Assembler, sp *scannerPeeker, p *instructionParams) bool {
	if err := requireOperand(sp, p); err != nil {
		p.logger.LogError(err.Error())
		return false
	}

	v, err := strconv.ParseUint(p.operandToken, 8, 16)
	if err != nil {
		p.logger.LogErrorf("unable to parse %v (%v)", p.operandToken, err.Error())
		return false
	}

	if v > 077777 {
		p.logger.LogErrorf("%v is out of range", p.operandToken)
		return false
	}

	a.requireLocation(p.logger)
	a.incLocation()

	action := func(a *Assembler) bool {
		return a.writeWordToImage(p.logger, uint16(v))
	}
	a.queueOperation(action)

	return true
}
