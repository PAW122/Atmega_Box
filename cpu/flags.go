package cpu

func (cpu *CPU) setFlag(bit uint8, val bool) {
	if val {
		cpu.SREG |= (1 << bit)
	} else {
		cpu.SREG &^= (1 << bit)
	}
}

func (cpu *CPU) setFlagsAdd(rd, rr, r uint8) {
	// Bit 3 i 7 oryginalnych wartości i wyniku
	rd3, rr3, r3 := (rd>>3)&1, (rr>>3)&1, (r>>3)&1
	rd7, rr7, r7 := (rd>>7)&1, (rr>>7)&1, (r>>7)&1

	// H - Half carry z bitu 3 do 4
	h := (rd3 & rr3) | (rr3 & ^r3) | (^r3 & rd3)
	// V - Overflow (przepełnienie przy liczbach ze znakiem)
	v := (rd7 & rr7 & ^r7) | (^rd7 & ^rr7 & r7)
	// N - Negative (czy bit 7 wyniku = 1)
	n := r7
	// Z - Zero
	z := r == 0
	// C - Carry (przeniesienie z bitu 7)
	c := (rd7 & rr7) | (rr7 & ^r7) | (^r7 & rd7)
	// S - Sign = N ⊕ V
	s := n ^ v

	cpu.setFlag(5, h != 0) // H
	cpu.setFlag(3, v != 0) // V
	cpu.setFlag(2, n != 0) // N
	cpu.setFlag(1, z)      // Z
	cpu.setFlag(0, c != 0) // C
	cpu.setFlag(4, s != 0) // S
}

func (cpu *CPU) setFlagsLogic(r uint8) {
	n := (r >> 7) & 1
	z := r == 0
	v := false
	s := (n ^ boolToBit(v)) != 0

	cpu.setFlag(1, z)         // Z
	cpu.setFlag(2, n != 0)    // N
	cpu.setFlag(3, v)         // V
	cpu.setFlag(4, s)         // S
}

func boolToBit(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}