package cpu

import (
	"fmt"
	"io"

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

type CPU struct {
	mm redirectedMemory

	reg        registers
	intsOff    bool
	pendingInt *interrupt
}

func NewCPU(mem io.Reader) (*CPU, error) {
	mm := new(memory.Main)
	if mem != nil {
		l := &memory.Loader{MM: mm}
		if _, err := io.Copy(l, mem); err != nil {
			return nil, err
		}
	}

	var cpu CPU
	cpu.mm.reg = &cpu.reg
	cpu.mm.mm = mm
	return &cpu, nil
}

func (c *CPU) Run() {
	c.reg.Set(regZ, 04000)
	var (
		time13Cycles     int
		time4Cycles      = -interval7_5ms
		time5Cycles      = -interval5ms
		pendingSequences []*sequence
	)
	logc := make(chan logEvent, 1000)
	go processLogEvents(logc)

	for {
		var timing int

		// check for any pending unprogrammed sequences
		if len(pendingSequences) > 0 {
			seq := pendingSequences[len(pendingSequences)-1]
			pendingSequences = pendingSequences[:len(pendingSequences)-1]

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


			// now increment the PC counter
			c.reg[regZ]++

			instr, address := decodeInstruction(val)
			logc <- logEvent{
				z:       z,
				code:    val,
				instr:   &instr,
				address: address,
			}

			if err := instr.execute(c, &instr, address); err != nil {
				panic(err)
			}

			timing = instr.timing
		}

		time13Cycles += timing
		time4Cycles += timing
		time5Cycles += timing

		if time13Cycles >= interval10ms {
			time13Cycles -= interval10ms
			pendingSequences = append(pendingSequences, &usPINCTime1)
			pendingSequences = append(pendingSequences, &usPINCTime3)
		}
		if time4Cycles >= interval10ms {
			time4Cycles -= interval10ms
			pendingSequences = append(pendingSequences, &usPINCTime4)
		}
		if time5Cycles >= interval10ms {
			time5Cycles -= interval10ms
			pendingSequences = append(pendingSequences, &usPINCTime5)
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

type logEvent struct {
	z       uint16
	code    uint16
	instr   *instruction
	address uint16
}

func processLogEvents(logc chan logEvent) {
	for e := range logc {
		fmt.Printf("%04o: %05o (%04x, %016b)", e.z, e.code, e.code<<1, e.code)
		fmt.Printf(" {%-6s %05o}\n", e.instr.name, e.address)
	}
}
