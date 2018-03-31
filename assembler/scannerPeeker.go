package assembler

import "bufio"

type scannerPeeker struct {
	scanner *bufio.Scanner
	primed  bool
	result  bool
}

func newScannerPeeker(s *bufio.Scanner) *scannerPeeker {
	return &scannerPeeker{scanner: s}
}

func (sp *scannerPeeker) prime() {
	if !sp.primed {
		sp.result = sp.scanner.Scan()
		sp.primed = true
	}
}

func (sp *scannerPeeker) Peek() (string, bool) {
	sp.prime()
	return sp.scanner.Text(), sp.result
}

func (sp *scannerPeeker) Consume() (string, bool) {
	sp.prime()
	sp.primed = false
	return sp.scanner.Text(), sp.result
}
