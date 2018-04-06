package assembler

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var aliases = map[string]string{
	"TCR": "TC",
	"CAF": "CA",
	"CAE": "CA",
	"NDX": "INDEX",
	"MSK": "MASK",
}

type symbolResolver interface {
	resolveSymbol(s string) (uint16, bool)
}

type instructionParams struct {
	logger   problemLogger
	resolver symbolResolver

	location psudoAddress
	extended bool

	instToken    string
	operandToken string
}

func (p *instructionParams) resolveOperand() (uint16, error) {
	kind := clasifyToken(p.operandToken)
	switch kind {
	case octalToken:
		v, err := strconv.ParseUint(p.operandToken, 8, 16)
		if err != nil {
			return 0, fmt.Errorf("unable to parse %v (%v)", p.operandToken, err.Error())
		}
		return uint16(v), nil
	case decimalToken:
		return 0, errors.New("decimal operands are not valid, use octal")
	case symbolToken:
		v, ok := p.resolver.resolveSymbol(p.operandToken)
		if !ok {
			return 0, fmt.Errorf("symbol %v is undefined", p.operandToken)
		}
		return v, nil
	default:
		panic(fmt.Sprintf("unknown token kind (%X)", kind))
	}
}

type operationFinalizer func(a *Assembler) bool

//Assembler is used to for assembling source files into AGC images.
type Assembler struct {
	sourceLineNum int
	location      psudoAddress
	locMsgHandler func(pl problemLogger)

	extended bool

	symbols    map[string]uint16
	Problems   []Problem
	errorCount int

	operations []operationFinalizer
	image      map[uint16]*[1024]uint16
}

//ErrorCount reports the number of errors encountered.
func (a *Assembler) ErrorCount() int {
	return a.errorCount
}

func (a *Assembler) resolveSymbol(s string) (uint16, bool) {
	if a.symbols == nil {
		return 0, false
	}

	v, ok := a.symbols[s]
	return v, ok
}

func (a *Assembler) defineSymbol(pl problemLogger, s string, v uint16) {
	if a.symbols == nil {
		a.symbols = make(map[string]uint16)
	} else if oldVal, ok := a.symbols[s]; ok {
		if oldVal == v {
			return
		}
		pl.LogWarningf("symbol %v is being re-defined", s)
	}

	a.symbols[s] = v
}

func (a *Assembler) setLocation(newLoc psudoAddress) {
	a.location = newLoc
	a.locMsgHandler = nil
}

func (a *Assembler) incLocation() {
	if a.location.isValid() {
		newLoc, err := a.location.nextValid()
		if err != nil {
			if a.location.isErasable() {
				a.locMsgHandler = logErasableEndError
			} else {
				a.locMsgHandler = logFixedEndError
			}
		} else {
			if newLoc.isBeginingOfSwitchableBank() {
				a.locMsgHandler = logBankWarning
			} else {
				a.locMsgHandler = nil
			}
		}
		a.location = newLoc
	}
}

func logErasableEndError(pl problemLogger) {
	pl.LogError("end of erasable memory")
}

func logFixedEndError(pl problemLogger) {
	pl.LogError("end of fixed memory")
}

func logBankWarning(pl problemLogger) {
	pl.LogWarning("address tansitioned to new switchable bank")
}

func (a *Assembler) requireLocation(logger problemLogger) psudoAddress {
	if a.locMsgHandler != nil {
		a.locMsgHandler(logger)
		a.locMsgHandler = nil
	}
	return a.location
}

func (a *Assembler) queueOperation(of operationFinalizer) {
	a.operations = append(a.operations, of)
}

//Assemble processes a file at the given path
func (a *Assembler) Assemble(filePath string) bool {
	r, err := os.Open(filePath)
	if err != nil {
		l := assemblerLogger{asm: a, fileName: filePath, lineNum: 0}
		l.LogErrorf("could not open file (%v)", err.Error())
		return false
	}
	defer func() { r.Close() }()

	return a.assembleReader(r, filePath)
}

func (a *Assembler) assembleReader(r io.Reader, filePath string) bool {
	lineScanner := bufio.NewScanner(r)

	a.location = psudoAddress(0xFFFF)
	a.extended = false

	line := 0
	for lineScanner.Scan() {
		line++
		pl := &assemblerLogger{asm: a, fileName: filePath, lineNum: line}

		a.parseLine(pl, lineScanner.Text())

		if a.errorCount >= 10 {
			pl.LogError("assembler stopped due to too many errors")
			return false
		}
	}

	if a.errorCount > 0 {
		return false
	}

	a.location = psudoAddress(0xFFFF)
	a.image = make(map[uint16]*[1024]uint16)
	for _, of := range a.operations {
		if !of(a) || a.errorCount > 0 {
			a.image = nil
			return false
		}
	}

	return true
}

func (a *Assembler) parseLine(pl problemLogger, line string) {
	//trim comments
	if c := strings.IndexRune(line, '#'); c >= 0 {
		line = line[:c]
	}

	line = strings.ToUpper(line)
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	sp := newScannerPeeker(s)

	token, ok := a.tryParseCommand(pl, sp)
	if ok {
		return
	}

	a.addLabel(pl, token)

	token, ok = a.tryParseCommand(pl, sp)
	if ok {
		return
	}

	pl.LogErrorf("unknown instruction %v", token)
}

func (a *Assembler) tryParseCommand(pl problemLogger, sp *scannerPeeker) (string, bool) {
	token, ok := sp.Consume()
	if !ok {
		return "", true
	}

	if a.tryToken(pl, token, sp) {
		demandEndOfLine(pl, sp)
		return "", true
	}

	return token, false
}

func (a *Assembler) tryToken(pl problemLogger, token string, sp *scannerPeeker) bool {
	p := &instructionParams{
		logger:    pl,
		resolver:  a,
		extended:  a.extended,
		instToken: token,
	}
	instName := p.instToken

	//Check for alias
	if alias, ok := aliases[instName]; ok {
		instName = alias
	}

	//Directive
	if dh, ok := directives[instName]; ok {
		dh(a, sp, p)
		return true
	}

	//Instruction
	if inst := findInstruction(p, instName); inst != nil {
		p.location = a.requireLocation(p.logger)

		a.queueOperation(getInstOperation(sp, p, inst))

		//prepare for next instruction
		a.extended = inst.setExtend
		a.incLocation()

		return true
	}

	//Symbol
	return a.tryParseSymbolDef(sp, p)
}

func demandEndOfLine(pl problemLogger, sp *scannerPeeker) {
	if t, ok := sp.Peek(); ok {
		pl.LogErrorf("expected end of line but token found (%v)", t)
	}
}

func getInstOperation(sp *scannerPeeker, p *instructionParams, i *instruction) operationFinalizer {
	if i.validateOperand != nil {
		if err := requireOperand(sp, p); err != nil {
			p.logger.LogError(err.Error())
			return nil
		}
	}

	return func(a *Assembler) bool {
		mc, ok := i.encode(p)
		if !ok {
			return false
		}

		return a.writeWordToImage(p.logger, mc)
	}
}

func (a *Assembler) tryParseSymbolDef(sp *scannerPeeker, p *instructionParams) bool {
	if t, ok := sp.Peek(); !ok {
		return false
	} else if t != "=" && t != "EQUALS" {
		return false
	}
	sp.Consume()

	if err := requireOperand(sp, p); err != nil {
		p.logger.LogError(err.Error())
		return true
	}

	val, err := p.resolveOperand()
	if err != nil {
		p.logger.LogError(err.Error())
		return true
	}

	a.defineSymbol(p.logger, p.instToken, val)
	return true
}

func (a *Assembler) addLabel(pl problemLogger, s string) {
	pa := a.requireLocation(pl)
	var val uint16
	if !pa.isValid() {
		pl.LogErrorf("location for label %v is undefined", s)
		return
	}

	val = pa.asOperand()
	a.defineSymbol(pl, s, val)
}

func (a *Assembler) writeWordToImage(pl problemLogger, v uint16) bool {
	loc := a.requireLocation(pl)
	if !loc.isValid() {
		pl.LogError("writing to invalid address")
		return false
	}

	e, b, o := loc.bankAndOffset()
	if e {
		pl.LogError("writing to erasable memory")
		return false
	}

	bank, ok := a.image[b]
	if !ok {
		bank = new([1024]uint16)
		a.image[b] = bank
	}

	bank[o] = v

	a.incLocation()
	return true
}

//ImageBuilt indicates that an AGC image has been successfully built and can be extracted with WriteOut.
func (a *Assembler) ImageBuilt() bool {
	return a.image != nil
}

//WriteOut writes the AGC image to the specified Writer.
func (a *Assembler) WriteOut(w io.Writer) error {
	if !a.ImageBuilt() {
		return errors.New("a valid image has not been assembled")
	}

	for i := uint16(0); i < 32+8; i++ {
		bank, ok := a.image[i]
		if !ok {
			bank = new([1024]uint16)
		}

		for _, mc := range bank {
			if err := binary.Write(w, binary.BigEndian, mc); err != nil {
				return err
			}
		}
	}

	return nil
}
