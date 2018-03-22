package cpu

import (
	"fmt"
)

type instruction struct {
	name        string
	code        uint16
	addressMask uint16
	timing      int
	execute     func(*CPU, *instruction, uint16) error
}

func decodeInstruction(machineCode uint16) (instruction, uint16) {
	var bestMatch *instruction
	for i, inst := range instructionSet {
		if inst.code == machineCode&^inst.addressMask && (bestMatch == nil || inst.addressMask < bestMatch.addressMask) {
			bestMatch = &instructionSet[i]
		}
	}

	if bestMatch == nil {
		panic("bad instruction")
	}

	return *bestMatch, machineCode & bestMatch.addressMask
}

const (
	maskNoAddress    = 00000
	mask10BitAddress = 01777
	mask12BitAddress = 07777
)

var instructionSet = []instruction{
	instruction{
		name:        "RELINT",
		code:        000003,
		addressMask: maskNoAddress,
		timing:      1,
		execute: func(c *CPU, i *instruction, addr uint16) error {
			c.intsOff = false
			fmt.Println("    interrupts enabled")
			return nil
		},
	},
	instruction{
		name:        "INHINT",
		code:        000004,
		timing:      1,
		addressMask: maskNoAddress,
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
		timing:      1,
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
		timing:      2,
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
		name:        "CS",
		code:        040000,
		addressMask: mask12BitAddress,
		timing:      2,
		execute: func(c *CPU, i *instruction, addr uint16) error {
			val, err := c.mm.Read(int(addr))
			if err != nil {
				return err
			}
			c.reg.Set(regA, ^val)
			// CA actually writes the original value back
			// out to memory (if the address is erasable)
			if addr&06000 == 0 {
				// if bit 11 or 12 are set then this address is in
				// fixed memory and we wouldn't want to do this write
				if err := c.mm.Write(int(addr), val); err != nil {
					return err
				}
			}
			fmt.Printf("    %05o loaded into A from %05o\n", ^val, addr)
			return nil
		},
	},
	instruction{
		name:        "TS",
		code:        054000,
		addressMask: mask10BitAddress,
		timing:      2,
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

type sequence struct {
	name    string
	timing  int
	execute func(*CPU, *sequence) *sequence
}

var (
	usPINCTime1 = sequence{
		name:   "PINC TIME1",
		timing: 1,
		execute: func(c *CPU, seq *sequence) *sequence {
			if c.reg.Increment(regTIME1) {
				return &usPINCTime2
			}
			return nil
		},
	}
	usPINCTime2 = sequence{
		name:   "PINC TIME2",
		timing: 1,
		execute: func(c *CPU, seq *sequence) *sequence {
			c.reg.Increment(regTIME2)
			return nil
		},
	}
	usPINCTime3 = sequence{
		name:   "PINC TIME3",
		timing: 1,
		execute: func(c *CPU, seq *sequence) *sequence {
			if c.reg.Increment(regTIME3) {
				c.interrupt(intT3RUPT)
			}
			return nil
		},
	}
	usPINCTime4 = sequence{
		name:   "PINC TIME4",
		timing: 1,
		execute: func(c *CPU, seq *sequence) *sequence {
			if c.reg.Increment(regTIME4) {
				c.interrupt(intT4RUPT)
			}
			return nil
		},
	}
	usPINCTime5 = sequence{
		name:   "PINC TIME5",
		timing: 1,
		execute: func(c *CPU, seq *sequence) *sequence {
			if c.reg.Increment(regTIME5) {
				c.interrupt(intT5RUPT)
			}
			return nil
		},
	}
)
