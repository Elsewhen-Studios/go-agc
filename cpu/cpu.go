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

type CPU struct {
	mm redirectedMemory

	reg     registers
	intsOff bool
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

	for {
		var timing int

		// check for any pending unprogrammed sequences
		if len(pendingSequences) > 0 {
			seq := pendingSequences[len(pendingSequences)-1]
			pendingSequences = pendingSequences[:len(pendingSequences)-1]

			fmt.Printf("----: %s\n", seq.name)
			if subSeq := seq.execute(c, seq); subSeq != nil {
				pendingSequences = append(pendingSequences, subSeq)
			}

			timing = seq.timing
		} else {
			val, err := c.mm.Read(int(c.reg[regZ]))
			if err != nil {
				panic(err)
			}

			fmt.Printf("%04o: %05o (%04x, %016b)", c.reg[regZ], val, val<<1, val)

			// now increment the PC counter
			c.reg[regZ]++

			instr, address := decodeInstruction(val)
			fmt.Printf(" {%-6s %05o} T1/3:%3d  T4:%3d  T5:%3d\n", instr.name, address, time13Cycles, time4Cycles, time5Cycles)

			if err := instr.execute(c, &instr, address); err != nil {
				panic(err)
			}

			timing = instr.timing
		}

		time13Cycles += timing
		time4Cycles += timing
		time5Cycles += timing

		if time13Cycles >= interval10ms {
			// time to increment TIME1 & TIME3
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
