package cpu

import "fmt"

type instruction struct {
	name        string
	code        uint16
	addressMask uint16
	execute     func(*CPU, *instruction, uint16) error
}

func decodeInstruction(instr uint16) (instruction, uint16) {
	for _, i := range instructionSet {
		addr := instr & i.addressMask
		if i.code == instr^addr {
			return i, addr
		}
	}

	panic("bad instruction")
}

const (
	mask12BitAddress = 07777
	mask10BitAddress = 01777
)

var instructionSet = []instruction{
	instruction{
		name:        "CA",
		code:        030000,
		addressMask: mask12BitAddress,
		execute: func(c *CPU, i *instruction, addr uint16) error {
			val, err := c.mm.Read(int(addr))
			if err != nil {
				return err
			}
			c.reg.Set(regA, val)
			// CA actually writes the original value back
			// out to memory (if the address is erasable)
			if addr&06000 == 0 {
				// if bit 11 or 12 are set then this address is in
				// fixed memory and we wouldn't want to do this write
				if err := c.mm.Write(int(addr), val); err != nil {
					return err
				}
			}
			fmt.Printf("    %05o loaded into A from %05o\n", val, addr)
			return nil
		},
	},
}
