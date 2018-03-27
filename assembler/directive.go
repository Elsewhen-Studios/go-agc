package assembler

import (
	"bufio"
)

type directiveHandler func(a *assembler, ts *bufio.Scanner, p *instructionParams) bool

var directives = map[string]directiveHandler{
	"SETLOC": setLoc,
}

func setLoc(a *assembler, ts *bufio.Scanner, p *instructionParams) bool {
	if err := requireOperand(ts, p); err != nil {
		p.logger.LogError(err.Error())
		return false
	}

	val, err := p.ResolveOperand()
	if err != nil {
		p.logger.LogError(err.Error())
		return false
	}

	newLoc := psudoAddress(val)
	if !newLoc.isValid() {
		p.logger.LogErrorf("%v is not a valid psudo-address", p.operandToken)
		return false
	}

	action := func(a *assembler) bool {
		a.setLocation(newLoc)
		return true
	}
	action(a)
	a.queueOperation(action)

	return true
}
