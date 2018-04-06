package cpu

import "fmt"

type LogEventType int

const (
	LogInstruction LogEventType = iota
	LogTimer
	LogUSequence
)

type LogEvent interface {
	fmt.Stringer
	Type() LogEventType
}

type Logger struct {
	logc         chan LogEvent
	enabledTypes map[LogEventType]bool
}

func NewLogger(buffer int) *Logger {
	return &Logger{
		logc:         make(chan LogEvent, buffer),
		enabledTypes: make(map[LogEventType]bool),
	}
}

func (l *Logger) Process() {
	for e := range l.logc {
		if l.enabledTypes[e.Type()] {
			fmt.Println(e.String())
		}
	}
}

func (l *Logger) Stop() {
	close(l.logc)
}

func (l *Logger) Log(e LogEvent) {
	l.logc <- e
}

type InstructionEvent struct {
	z       uint16
	code    uint16
	instr   *instruction
	address uint16
}

func (e InstructionEvent) Type() LogEventType { return LogInstruction }

func (e InstructionEvent) String() string {
	return fmt.Sprintf("%04o: %05o (%04x) {%-6s %05o}", e.z, e.code, e.code, e.instr.name, e.address)
}

type USequenceEvent struct {
	seq *sequence
}

func (e USequenceEvent) Type() LogEventType { return LogUSequence }

func (e USequenceEvent) String() string {
	return fmt.Sprintf("----: %s", e.seq.name)
}

type TimerEvent struct {
	name string
}

func (e TimerEvent) Type() LogEventType { return LogTimer }

func (e TimerEvent) String() string {
	return fmt.Sprintf("Timer %s fired!", e.name)
}
