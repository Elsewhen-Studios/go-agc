package assembler

import (
	"math"
	"strconv"
	"strings"
)

type directiveHandler func(a *Assembler, sp *scannerPeeker, p *instructionParams) bool

var directives = map[string]directiveHandler{
	"SETLOC": setLoc,
	"OCT":    oct,
	"DEC":    dec,
	"2DEC":   dec2,
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

func tryParseDec(sp *scannerPeeker, p *instructionParams, dp bool) (uint16, uint16, bool) {
	if err := requireOperand(sp, p); err != nil {
		p.logger.LogError(err.Error())
		return 0, 0, false
	}

	max := 1 << 28
	var i int
	if _, ok := sp.Peek(); !ok && !strings.Contains(p.operandToken, ".") {
		//asume integer
		v, err := strconv.ParseInt(p.operandToken, 10, 64)
		if err != nil {
			p.logger.LogErrorf("unable to parse %v (%v)", p.operandToken, err.Error())
			return 0, 0, false
		}

		i = int(v)
		if !dp {
			i <<= 14
		}

		if i < 0 {
			i = -i
		}

		if i >= max {
			if dp {
				p.logger.LogErrorf("value out of range (%v)", v)
			} else {
				p.logger.LogErrorf("value out of range (%v)", v)
			}
			return 0, 0, false
		}
	} else {
		v, err := strconv.ParseFloat(p.operandToken, 64)
		if err != nil {
			p.logger.LogErrorf("unable to parse %v (%v)", p.operandToken, err.Error())
			return 0, 0, false
		}

		errorOut := false
		for {
			s, ok := sp.Consume()
			if !ok {
				break
			}

			var base float64
			if strings.HasPrefix(s, "E") {
				base = 10
			} else if strings.HasPrefix(s, "B") {
				base = 2
			} else {
				p.logger.LogErrorf("invalid exponent base specified (%v)", 2)
				errorOut = true
				continue
			}

			exp, err := strconv.ParseFloat(string([]rune(s)[1:]), 64)
			if err != nil {
				p.logger.LogErrorf("unable to parse exponent %v (%v)", s, err.Error())
				errorOut = true
				continue
			}

			v *= math.Pow(base, exp)
		}

		if errorOut {
			return 0, 0, false
		}

		v *= float64(max)
		i = int(math.Round(v))
		if i < 0 {
			i = -i
		}

		if i >= max {
			p.logger.LogErrorf("value out of range (%f)", v)
			return 0, 0, false
		}
	}

	var sign uint16
	if strings.HasPrefix(p.operandToken, "-") {
		sign = 1 << 14
	}

	mask := ((1 << 14) - 1)
	l := uint16(i&mask) | sign
	h := uint16((i>>14)&mask) | sign

	return h, l, true
}

func dec(a *Assembler, sp *scannerPeeker, p *instructionParams) bool {
	h, _, ok := tryParseDec(sp, p, false)
	if !ok {
		return false
	}

	action := func(a *Assembler) bool {
		return a.writeWordToImage(p.logger, h)
	}
	a.queueOperation(action)

	return true
}

func dec2(a *Assembler, sp *scannerPeeker, p *instructionParams) bool {
	h, l, ok := tryParseDec(sp, p, true)
	if !ok {
		return false
	}

	action := func(a *Assembler) bool {
		return a.writeWordToImage(p.logger, h) && a.writeWordToImage(p.logger, l)
	}
	a.queueOperation(action)

	return true
}
