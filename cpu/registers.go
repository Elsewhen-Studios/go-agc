package cpu

import (
	"math"

	"github.com/Elsewhen-Studios/go-agc/memory"
)

type register int

const (
	regA register = iota
	regL
	regQ
	regEB
	regFB
	regZ
	regBB
	regCYR   = 020
	regTIME2 = 024
	regTIME1 = iota
	regTIME3
	regTIME4
	regTIME5
	regTIME6
)

type registers [061]uint16

func (reg *registers) Set(r register, val uint16) {
	switch r {
	case regL:
		// the L register is a 15 bit register
		// but it just discards the 16th bit
		val &= 077777
	case regZ:
		// the Z register is a 12 bit register
		// so discard the rest
		val &= 07777
	case regEB:
		// mask off the EB bits and copy them
		// to the BB register
		reg[regBB] |= val & 003400 >> 8
	case regFB:
		// mask off the FB bits and copy them
		// to the BB register
		reg[regBB] |= val & 076000
	case regBB:
		// make sure to copy the EB and FB parts
		// of BB back to their respective registers
		reg[regEB] = val & 07 << 8
		reg[regFB] = val & 076000
	case regCYR:
		// do a 15-bit rotation to the right
		val = ((val << 14) | (val >> 1)) & 077777
	}
	// now that we've done all our special handling
	// we can write the value into the register
	reg[r] = val
}

// Increment increases the given register by one and returns
// whether an overflow ocurred.
func (reg *registers) Increment(r register) (overflow bool) {
	var max uint16 = math.MaxUint16 >> 1
	if r == regA || r == regL {
		// these are the 16 bit registers
		max = math.MaxUint16
	}

	val := reg[r]
	if val >= max {
		// register is going to overflow when we increment it
		overflow = true
		val = 0
	} else {
		val++
	}

	reg.Set(r, val)
	return
}

type redirectedMemory struct {
	reg *registers
	mm  *memory.Main
}

func newRedirectedMemory(r *registers, mm *memory.Main) *redirectedMemory {
	return &redirectedMemory{
		reg: r,
		mm:  mm,
	}
}

func (rm *redirectedMemory) Read(address int) (uint16, error) {
	if address < len(rm.reg) {
		return rm.reg[address], nil
	}
	return rm.mm.Read(address)
}

func (rm *redirectedMemory) Write(address int, val uint16) error {
	if address < len(rm.reg) {
		rm.reg.Set(register(address), val)
		return nil
	}
	return rm.mm.Write(address, val)
}
