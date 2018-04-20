package cpu

import (
	"bufio"
	"fmt"
	"os"
)

type DebugEvent struct {
	z       uint16
	code    uint16
	instr   *instruction
	address uint16
}

type Debugger interface {
	Debug(e DebugEvent)
}

type noDebugger int

func (d *noDebugger) Debug(e DebugEvent) {}

type InteractiveDebugger struct {
	dbgevtc  chan DebugEvent
	dbgctlc  chan struct{}
	outc     chan string
	commandc chan interface{}
}

func NewInteractiveDebugger() *InteractiveDebugger {
	return &InteractiveDebugger{
		dbgevtc:  make(chan DebugEvent),
		dbgctlc:  make(chan struct{}),
		outc:     make(chan string),
		commandc: make(chan interface{}),
	}
}
func (d *InteractiveDebugger) Run() {
	var (
		bp    = map[uint16]bool{04000: true}
		steps int
		stdin = bufio.NewScanner(os.Stdin)
	)

debugLoop:
	for {
		e := <-d.dbgevtc

		// check to see if we should break
		var doBreak bool
		if steps > 0 {
			steps--
			if steps == 0 {
				doBreak = true
			}
		}
		if bp[e.z] {
			doBreak = true
		}

		if doBreak {
			// time to take a break, output the
			// event and start a prompt
			fmt.Printf("%04o: %05o (%04x) {%-6s %05o}\n", e.z, e.code, e.code, e.instr.name, e.address)

		promptLoop:
			for {
				// get a command from the user
				fmt.Printf("> ")
				if !stdin.Scan() {
					break debugLoop
				}
				input := stdin.Text()

				// process input
				var cmd string
				fmt.Sscan(input, &cmd)
				switch cmd {
				case "step", "s":
					steps = 1
					break promptLoop
				case "stepi", "si":
					fmt.Sscanf(input, fmt.Sprintf("%s %%d", cmd), &steps)
					break promptLoop
				case "run", "r":
					break promptLoop
				case "breakpoint", "bp":
					var addr uint16
					fmt.Sscanf(input, fmt.Sprintf("%s %%o", cmd), &addr)
					bp[addr] = !bp[addr]
					break
				case "":
					break
				default:
					fmt.Println("unrecognized command", cmd)
					break
				}
			}
		}

		// now let the CPU keep going
		d.dbgctlc <- struct{}{}
	}

	if err := stdin.Err(); err != nil {
		panic(err)
	}
}

func (d *InteractiveDebugger) Debug(e DebugEvent) {
	d.dbgevtc <- e
	<-d.dbgctlc
}

type breakpointCommand uint16
type stepiCommand int
type runCommand struct{}
