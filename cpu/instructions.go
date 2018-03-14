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
		name:        "RELINT",
		code:        000003,
		addressMask: 00000,
		execute: func(c *CPU, i *instruction, addr uint16) error {
			c.intsOff = false
			fmt.Println("    interrupts enabled")
			return nil
		},
	},
	instruction{
		name:        "INHINT",
		code:        000004,
		addressMask: 00000,
		execute: func(c *CPU, i *instruction, addr uint16) error {
			c.intsOff = true
			fmt.Println("    interrupts disabled")
			return nil
		},
	},
	instruction{
		name:        "TCF",
		code:        010000,
		addressMask: mask12BitAddress,
		execute: func(c *CPU, i *instruction, addr uint16) error {
			fmt.Printf("    jumping to %05o\n", addr)
			c.reg.Set(regZ, addr)
			return nil
		},
	},
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
	instruction{
		name:        "TS",
		code:        054000,
		addressMask: mask10BitAddress,
		execute: func(c *CPU, i *instruction, addr uint16) error {
			fmt.Printf("    Wrote %05o from A to %05o\n", c.reg[regA], addr)
			if err := c.mm.Write(int(addr), c.reg[regA]); err != nil {
				return err
			}
			switch c.overflow() {
			case 0:
				// no overflow, no special behavior
			case -1:
				// negative overflow, set A to -1 and increment Z (to skip the next instruction)
				c.reg.Set(regA, 0177776)
				c.reg[regZ]++
				fmt.Println("    A set to -1 and skipping next instruction")
			case +1:
				// positive overflow, set A to +1 and increment Z (to skip the next instruction)
				c.reg.Set(regA, 1)
				c.reg[regZ]++
				fmt.Println("    A set to +1 and skipping next instruction")
			}
			return nil
		},
	},
}
