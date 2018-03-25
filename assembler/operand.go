package assembler

import (
	"bufio"
	"errors"
	"fmt"
	"unicode"
)

func requireOperand(ts *bufio.Scanner, p *instructionParams) error {
	if !ts.Scan() {
		return errors.New("operand expected but none found")
	}
	p.operandToken = ts.Text()
	return nil
}

type tokenKind int

const (
	octalToken = tokenKind(iota)
	decimalToken
	symbolToken
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

type operandValidator func(val uint16, p *instructionParams) error

func validateTCOperand(val uint16, p *instructionParams) error {
	if err := requireAnyMemoryOperand(val, p); err != nil {
		return err
	} else if val == 00003 || val == 00004 || val == 00006 {
		return fmt.Errorf("%v is not a valid operand for %v", p.operandToken, p.instToken)
	}
	return nil
}

func validateINDEXOperand(val uint16, p *instructionParams) error {
	if err := requireErasable(val, p); err != nil {
		return err
	} else if val == 00017 {
		return fmt.Errorf("%v is not a valid operand for %v", p.operandToken, p.instToken)
	}
	return nil
}

func requireAnyMemoryOperand(val uint16, p *instructionParams) error {
	if val > 07777 {
		return fmt.Errorf("%v is not a valid memory address", p.operandToken)
	}
	return nil
}

func requireErasable(val uint16, p *instructionParams) error {
	if val > 01777 {
		return fmt.Errorf("%v is not a valid erasable memory address", p.operandToken)
	}
	return nil
}

func requireChannel(val uint16, p *instructionParams) error {
	if val > 00777 {
		return fmt.Errorf("%v is not a valid I/O channel address", p.operandToken)
	}
	return nil
}

func requireFixed(val uint16, p *instructionParams) error {
	if val < 02000 || val > 07777 {
		return fmt.Errorf("%v is not a valid fixed memory address", p.operandToken)
	}
	return nil
}
