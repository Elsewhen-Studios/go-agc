package cpu

import (
	"math"
	"testing"

	"github.com/Elsewhen-Studios/go-agc/memory"
	"github.com/stretchr/testify/assert"
)

func TestRegisterA(t *testing.T) {
	// register A just round trips with no modification
	var reg registers

	reg.Set(regA, 0177777)
	assert.Equal(t, uint16(0177777), reg[regA])
}

func TestRegisterL(t *testing.T) {
	// register L truncates to 15 bits
	var reg registers

	reg.Set(regL, 0177777)
	assert.Equal(t, uint16(0077777), reg[regL])
}

func TestRegisterQ(t *testing.T) {
	// register Q just round trips with no modification
	var reg registers

	reg.Set(regQ, 0177777)
	assert.Equal(t, uint16(0177777), reg[regQ])
}

func TestRegisterEB(t *testing.T) {
	// ensure writes to EB are reflected in BB
	var reg registers
	// setup BB to have values from FB to ensure
	// that writes to EB are not destroying FB
	reg[regBB] = 052000

	reg.Set(regEB, 003000)
	assert.Equal(t, uint16(003000), reg[regEB])
	assert.Equal(t, uint16(052006), reg[regBB])
}

func TestRegisterFB(t *testing.T) {
	// ensure writes to FB are reflected in BB
	var reg registers
	// setup BB to have values from EB to ensure
	// that writes to FB are not destroying EB
	reg[regBB] = 000006

	reg.Set(regFB, 052000)
	assert.Equal(t, uint16(052000), reg[regFB])
	assert.Equal(t, uint16(052006), reg[regBB])
}

func TestRegisterZ(t *testing.T) {
	// register Z truncates to 12 bits
	var reg registers

	reg.Set(regZ, 0177777)
	assert.Equal(t, uint16(0007777), reg[regZ])
}

func TestRegisterBB(t *testing.T) {
	// ensure writes to BB are reflected in EB and FB
	var reg registers

	reg.Set(regBB, 052006)
	assert.Equal(t, uint16(052006), reg[regBB])
	assert.Equal(t, uint16(052000), reg[regFB])
	assert.Equal(t, uint16(003000), reg[regEB])
}

func TestRegisterCYR(t *testing.T) {
	// register CYR does a right rotation by 1 bit
	var reg registers

	reg.Set(regCYR, 052525) // lsb is 1
	assert.Equal(t, uint16(065252), reg[regCYR])
	reg.Set(regCYR, 025252) // lsb is 0
	assert.Equal(t, uint16(012525), reg[regCYR])
}

func TestRegisterIncrement(t *testing.T) {
	scenarios := []struct {
		name       string
		r          register
		start, end uint16
		overflow   bool
	}{
		{"A - no overflow", regA, 0x7FFF, 0x8000, false},
		{"A - overflow", regA, math.MaxUint16, 0, true},
		{"TIME1 - no overflow", regTIME1, 0x7FFE, 0x7FFF, false},
		{"TIME1 - overflow", regTIME1, 0x7FFF, 0, true},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// arrange
			var reg registers
			reg[scenario.r] = scenario.start

			// act
			overflow := reg.Increment(scenario.r)

			// assert
			assert.Equal(t, scenario.overflow, overflow, "overflow")
			assert.Equal(t, scenario.end, reg[scenario.r], "end value")
		})
	}
}

func TestRedirectedMemory_ReadRegister(t *testing.T) {
	var (
		reg registers
		mm  memory.Main
	)
	rm := newRedirectedMemory(&reg, &mm)

	reg.Set(regZ, 0xA)
	mm.Write(05, 0xB)

	val, err := rm.Read(05)

	assert.NoError(t, err)
	assert.Equal(t, uint16(0xA), val)

}

func TestRedirectedMemory_ReadMemory(t *testing.T) {
	var (
		reg registers
		mm  memory.Main
	)
	rm := newRedirectedMemory(&reg, &mm)

	mm.Write(123, 0xABC)

	val, err := rm.Read(123)

	assert.NoError(t, err)
	assert.Equal(t, uint16(0xABC), val)
}

func TestRedirectedMemory_WriteRegister(t *testing.T) {
	var (
		reg registers
		mm  memory.Main
	)
	rm := newRedirectedMemory(&reg, &mm)

	err := rm.Write(05, 0xABC)

	assert.NoError(t, err)
	assert.Equal(t, uint16(0xABC), reg[regZ])
	val, err := mm.Read(05)
	assert.NoError(t, err)
	assert.Equal(t, uint16(0), val)
}

func TestRedirectedMemory_WriteMemory(t *testing.T) {
	var (
		reg registers
		mm  memory.Main
	)
	rm := newRedirectedMemory(&reg, &mm)

	err := rm.Write(123, 0xABC)

	assert.NoError(t, err)
	assert.Equal(t, uint16(0), reg[regZ])

	val, err := mm.Read(123)
	assert.NoError(t, err)
	assert.Equal(t, uint16(0xABC), val)
}
