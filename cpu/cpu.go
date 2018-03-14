package cpu

import (
	"fmt"
	"io"

	"github.com/Elsewhen-Studios/go-agc/memory"
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

	for {
		val, err := c.mm.Read(int(c.reg[regZ]))
		if err != nil {
			panic(err)
		}

		fmt.Printf("%04o: %05o (%04x, %016b)", c.reg[regZ], val, val<<1, val)

		// now increment the PC counter
		c.reg[regZ]++

		instr, address := decodeInstruction(val)
		fmt.Printf(" {%-6s %05o}\n", instr.name, address)

		if err := instr.execute(c, &instr, address); err != nil {
			panic(err)
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
