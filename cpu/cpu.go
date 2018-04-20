package cpu

import (
	"fmt"

	"github.com/Elsewhen-Studios/go-agc/memory"
)

const (
	interval10ms  = 893
	interval7_5ms = interval10ms * 3 / 4
	interval5ms   = interval10ms / 2
)

type interrupt int

const (
	intBOOT interrupt = iota
	intT6RUPT
	intT5RUPT
	intT3RUPT
	intT4RUPT
)

// CPU simulates the core logic of the AGC.
type CPU struct {
	mm redirectedMemory

	reg        registers
	intsOff    bool
	pendingInt *interrupt

	Debugger Debugger
}

// NewCPU creates a new CPU using the given main memory.
func NewCPU(mem *memory.Main) *CPU {
	if mem == nil {
		mem = new(memory.Main)
	}

	var cpu CPU
	cpu.mm.reg = &cpu.reg
	cpu.mm.mm = mem
	cpu.Debugger = new(noDebugger)
	return &cpu
}

// Run executes instructions from main memory.
func (c *CPU) Run() {
	c.reg.Set(regZ, 04000)
	var (
		pendingSequences []*sequence
		timers           = map[*sequence]*timer{
			&usPINCTime1: newTimer("TIME1", interval10ms, 0),
			&usPINCTime3: newTimer("TIME3", interval10ms, 0),
			&usPINCTime4: newTimer("TIME4", interval10ms, -interval7_5ms),
			&usPINCTime5: newTimer("TIME5", interval10ms, -interval5ms),
		}
	)
	log := newLogger(1000)
	go log.process()
	defer log.stop()

	for {
		var timing int

		// check for any pending unprogrammed sequences
		if len(pendingSequences) > 0 {
			seq := pendingSequences[len(pendingSequences)-1]
			pendingSequences = pendingSequences[:len(pendingSequences)-1]

			log.log(uSequenceEvent{seq: seq})
			if subSeq := seq.execute(c, seq); subSeq != nil {
				pendingSequences = append(pendingSequences, subSeq)
			}

			timing = seq.timing
		} else {
			z := c.reg[regZ]
			val, err := c.mm.Read(int(z))
			if err != nil {
				panic(err)
			}

			instr, address, err := decodeInstruction(val)
			if err != nil {
				panic(fmt.Sprintf("failed to decode instruction at %05o: %v", z, err))
			}
			c.Debugger.Debug(DebugEvent{
				z:       z,
				code:    val,
				instr:   &instr,
				address: address,
			})

			// now increment the PC counter
			c.reg[regZ]++

			if err := instr.execute(c, &instr, address); err != nil {
				panic(err)
			}

			timing = instr.timing
		}

		// increment our timers by the amount of cycles
		for useq, tmr := range timers {
			if tmr.Inc(timing) {
				// timer rolled over, queue up
				// the unprogrammed sequence
				pendingSequences = append(pendingSequences, useq)
				log.log(timerEvent{name: tmr.n})
			}
		}

		if c.pendingInt != nil {
			z := c.reg[regZ]
			c.reg.Set(regZRUPT, z)
			val, err := c.mm.Read(int(z))
			if err != nil {
				panic(err)
			}
			c.reg.Set(regBRUPT, val)
			c.reg.Set(regZ, 04000+uint16(*c.pendingInt)*4)
			fmt.Printf("INT! %04o - ZRUPT:%05o BRUPT:%05o\n", *c.pendingInt, z, val)
			c.intsOff = true
			c.pendingInt = nil
		}
	}
}

// overflow returns +1 if a positive overflow has ocurred, -1 if a negative overflow
// has ocurred, and zero if there has been no overflow.
func (c *CPU) overflow() int {
	// there has been an overflow if bits 16 and 15 of
	// register A differ
	if c.reg[regA]&0100000 != c.reg[regA]&0040000 {
		// there has been an overflow, now determine which kind
		if c.reg[regA]&0100000 == 0 {
			return +1
		}
		return -1
	}
	return 0
}

func (c *CPU) interrupt(i interrupt) {
	c.pendingInt = &i
}
