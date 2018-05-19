package cpu

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeInstruction(t *testing.T) {
	// act
	instr, address, err := decodeInstruction(030000 + 07777)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "CA", instr.name, "instr.name")
	assert.Equal(t, uint16(07777), address, "address")
}

func TestInstructionRELINT(t *testing.T) {
	runInstructionTest(t, "RELINT", "", func(t *testing.T, cpu *CPU, i *instruction) {
		// arrange
		cpu.intsOff = true

		// act
		err := i.execute(cpu, i, 0)

		// assert
		assert.NoError(t, err)
		assert.False(t, cpu.intsOff)
	})
}

func TestInstructionINHINT(t *testing.T) {
	runInstructionTest(t, "INHINT", "", func(t *testing.T, cpu *CPU, i *instruction) {
		// arrange
		cpu.intsOff = false

		// act
		err := i.execute(cpu, i, 0)

		// assert
		assert.NoError(t, err)
		assert.True(t, cpu.intsOff)
	})
}

func TestInstructionTCF(t *testing.T) {
	runInstructionTest(t, "TCF", "", func(t *testing.T, cpu *CPU, i *instruction) {
		// arrange
		cpu.reg.Set(regZ, 0123)

		// act
		err := i.execute(cpu, i, 0321)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, uint16(0321), cpu.reg[regZ])
	})
}

func TestInstructionCA(t *testing.T) {
	runInstructionTest(t, "CA", "read from memory", func(t *testing.T, cpu *CPU, i *instruction) {
		// memory cells are 15bits wide, so this verifies
		// proper handling of a 15bit word into a 16bit register

		// arrange
		cpu.mm.Write(123, 0100456)

		// act
		err := i.execute(cpu, i, 123)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, uint16(0140456), cpu.reg[regA])
	})

	runInstructionTest(t, "CA", "read from Q", func(t *testing.T, cpu *CPU, i *instruction) {
		// the Q register is also 16 bits wide, so this
		// test verifies that the 16th bit can be loaded
		// correctly

		// arrange
		cpu.reg.Set(regQ, 0100456)

		// act
		err := i.execute(cpu, i, uint16(regQ))

		// assert
		assert.NoError(t, err)
		assert.Equal(t, uint16(0100456), cpu.reg[regA])
	})

	runInstructionTest(t, "CA", "read from CYR", func(t *testing.T, cpu *CPU, i *instruction) {
		// the CA instruction rewrites the memory location after
		// reading it, so the editing registers will re-edit and
		// we want to verify that is the case

		// arrange
		cpu.reg.Set(regCYR, 000400) // this write will actually store 000200 since it gets rotated

		// act
		err := i.execute(cpu, i, uint16(regCYR))

		// assert
		assert.NoError(t, err)
		assert.Equal(t, uint16(000200), cpu.reg[regA])
		assert.Equal(t, uint16(000100), cpu.reg[regCYR])
	})

	runInstructionTest(t, "CA", "read from fixed memory", func(t *testing.T, cpu *CPU, i *instruction) {
		// the CA instruction rewrites the memory location after
		// reading it (but only if the address is in erasable
		// memory) so verify we don't try to write to fixed memory

		// arrange
		cpu.reg.Set(regA, 123)

		// act
		err := i.execute(cpu, i, uint16(04500))

		// assert
		assert.NoError(t, err)
		assert.Equal(t, uint16(0), cpu.reg[regA])
	})
}

func TestInstructionCS(t *testing.T) {
	runInstructionTest(t, "CS", "read from memory", func(t *testing.T, cpu *CPU, i *instruction) {
		// memory cells are 15bits wide, so this verifies
		// proper handling of a 15bit word into a 16bit register

		// arrange
		cpu.mm.Write(123, 0100456)

		// act
		err := i.execute(cpu, i, 123)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, uint16(037321), cpu.reg[regA])
	})

	runInstructionTest(t, "CS", "read from Q", func(t *testing.T, cpu *CPU, i *instruction) {
		// the Q register is also 16 bits wide, so this
		// test verifies that the 16th bit can be loaded
		// correctly

		// arrange
		cpu.reg.Set(regQ, 0100456)

		// act
		err := i.execute(cpu, i, uint16(regQ))

		// assert
		assert.NoError(t, err)
		assert.Equal(t, uint16(077321), cpu.reg[regA])
	})

	runInstructionTest(t, "CS", "read from CYR", func(t *testing.T, cpu *CPU, i *instruction) {
		// the CA instruction rewrites the memory location after
		// reading it, so the editing registers will re-edit and
		// we want to verify that is the case

		// arrange
		cpu.reg.Set(regCYR, 000400) // this write will actually store 000200 since it gets rotated

		// act
		err := i.execute(cpu, i, uint16(regCYR))

		// assert
		assert.NoError(t, err)
		assert.Equal(t, uint16(0177577), cpu.reg[regA])
		assert.Equal(t, uint16(0000100), cpu.reg[regCYR])
	})

	runInstructionTest(t, "CS", "read from fixed memory", func(t *testing.T, cpu *CPU, i *instruction) {
		// the CA instruction rewrites the memory location after
		// reading it (but only if the address is in erasable
		// memory) so verify we don't try to write to fixed memory

		// arrange
		cpu.reg.Set(regA, 123)

		// act
		err := i.execute(cpu, i, uint16(04500))

		// assert
		assert.NoError(t, err)
		assert.Equal(t, uint16(0177777), cpu.reg[regA])
	})
}

func TestInstructionDXCH(t *testing.T) {
	runInstructionTest(t, "DXCH", "simple exchange", func(t *testing.T, cpu *CPU, i *instruction) {
		// arrange
		cpu.reg.Set(regA, 0xA)
		cpu.reg.Set(regL, 0xB)
		cpu.mm.Write(0400, 0xC)
		cpu.mm.Write(0401, 0xD)

		// act
		err := i.execute(cpu, i, 0400)

		// assert
		require.NoError(t, err)
		assert.Equal(t, uint16(0xC), cpu.reg[regA], "register A")
		assert.Equal(t, uint16(0xD), cpu.reg[regL], "register L")
		memK1, _ := cpu.mm.Read(0400)
		memK2, _ := cpu.mm.Read(0401)
		assert.Equal(t, uint16(0xA), memK1, "memory @ K")
		assert.Equal(t, uint16(0xB), memK2, "memory @ K+1")
	})

	runInstructionTest(t, "DXCH", "overlap exchange", func(t *testing.T, cpu *CPU, i *instruction) {
		// arrange
		cpu.reg.Set(regA, 0xA)
		cpu.reg.Set(regL, 0xB)
		cpu.reg.Set(regQ, 0xC)

		// act
		err := i.execute(cpu, i, uint16(regL))

		// assert
		require.NoError(t, err)
		assert.Equal(t, uint16(0xC), cpu.reg[regA], "register A")
		assert.Equal(t, uint16(0xA), cpu.reg[regL], "register L")
		assert.Equal(t, uint16(0xB), cpu.reg[regQ], "register Q")
	})

	runInstructionTest(t, "DXCH", "exchange fixed memory", func(t *testing.T, cpu *CPU, i *instruction) {
		// act
		err := i.execute(cpu, i, 02100)

		// assert
		assert.Error(t, err)
	})

	runInstructionTest(t, "DXCH", "exchange K+1 fixed memory", func(t *testing.T, cpu *CPU, i *instruction) {
		// act
		err := i.execute(cpu, i, 01777)

		// assert
		assert.Error(t, err)
	})
}

func TestInstructionTS(t *testing.T) {
	runInstructionTest(t, "TS", "positive overflow", func(t *testing.T, cpu *CPU, i *instruction) {
		// arrange
		cpu.reg.Set(regA, 040123)
		cpu.reg.Set(regZ, 100)

		// act
		err := i.execute(cpu, i, 123)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, uint16(1), cpu.reg[regA])
		assert.Equal(t, uint16(101), cpu.reg[regZ])
		val, err := cpu.mm.Read(123)
		assert.NoError(t, err)
		assert.Equal(t, uint16(0123), val)
	})

	runInstructionTest(t, "TS", "negative overflow", func(t *testing.T, cpu *CPU, i *instruction) {
		// arrange
		cpu.reg.Set(regA, 0100123)
		cpu.reg.Set(regZ, 100)

		// act
		err := i.execute(cpu, i, 123)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, uint16(0177776), cpu.reg[regA])
		assert.Equal(t, uint16(101), cpu.reg[regZ])
		val, err := cpu.mm.Read(123)
		assert.NoError(t, err)
		assert.Equal(t, uint16(0140123), val)
	})

	runInstructionTest(t, "TS", "no overflow", func(t *testing.T, cpu *CPU, i *instruction) {
		// arrange
		cpu.reg.Set(regA, 0123)
		cpu.reg.Set(regZ, 100)

		// act
		err := i.execute(cpu, i, 123)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, uint16(0123), cpu.reg[regA])
		assert.Equal(t, uint16(100), cpu.reg[regZ])
		val, err := cpu.mm.Read(123)
		assert.NoError(t, err)
		assert.Equal(t, uint16(0123), val)
	})
}

func runInstructionTest(t *testing.T, name, scenario string, f func(*testing.T, *CPU, *instruction)) {
	subTestName := "instruction " + name
	if len(scenario) > 0 {
		subTestName += " - " + scenario
	}
	t.Run(subTestName, func(t *testing.T) {
		cpu := NewCPU(nil)
		i := getInstruction(name)

		f(t, cpu, &i)
	})
}

func getInstruction(name string) instruction {
	for _, i := range instructionSet {
		if i.name == name {
			return i
		}
	}
	panic(fmt.Sprintf("instruction %s not found!", name))
}
