package cpu

import "fmt"

type logEventType int

const (
	logInstruction logEventType = iota
	logTimer
	logUSequence
)

type logEvent interface {
	fmt.Stringer
	Type() logEventType
}

type logger struct {
	logc         chan logEvent
	enabledTypes map[logEventType]bool
}

func newLogger(buffer int) *logger {
	return &logger{
		logc:         make(chan logEvent, buffer),
		enabledTypes: make(map[logEventType]bool),
	}
}

func (l *logger) process() {
	for e := range l.logc {
		if l.enabledTypes[e.Type()] {
			fmt.Println(e.String())
		}
	}
}

func (l *logger) stop() {
	close(l.logc)
}

func (l *logger) log(e logEvent) {
	l.logc <- e
}

type instructionEvent struct {
	z       uint16
	code    uint16
	instr   *instruction
	address uint16
}

func (e instructionEvent) Type() logEventType { return logInstruction }

func (e instructionEvent) String() string {
	return fmt.Sprintf("%04o: %05o (%04x) {%-6s %05o}", e.z, e.code, e.code, e.instr.name, e.address)
}

type uSequenceEvent struct {
	seq *sequence
}

func (e uSequenceEvent) Type() logEventType { return logUSequence }

func (e uSequenceEvent) String() string {
	return fmt.Sprintf("----: %s", e.seq.name)
}

type timerEvent struct {
	name string
}

func (e timerEvent) Type() logEventType { return logTimer }

func (e timerEvent) String() string {
	return fmt.Sprintf("Timer %s fired!", e.name)
}
