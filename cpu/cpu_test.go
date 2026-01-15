package cpu

import (
	"testing"
)

type StepTest struct {
	name        string
	flash       []uint16
	setup       func(cpu *CPU)
	expectCheck func(cpu *CPU, t *testing.T)
}

func TestStepInstructions(t *testing.T) {
	tests := []StepTest{
		{
			name:  "LDI R16, 0x10",
			flash: []uint16{0xE110}, // LDI R17, 0x10 (bo 0xE010 = R17)
			setup: func(cpu *CPU) {},
			expectCheck: func(cpu *CPU, t *testing.T) {
				if cpu.R[17] != 0x10 {
					t.Errorf("expected R17 = 0x10, got 0x%02X", cpu.R[17])
				}
			},
		},
		{
			name:  "ADD R16, R17 = 0x10 + 0x22",
			flash: []uint16{0x0F01}, // ADD R16, R17
			setup: func(cpu *CPU) {
				cpu.R[16] = 0x10
				cpu.R[17] = 0x22
			},
			expectCheck: func(cpu *CPU, t *testing.T) {
				if cpu.R[16] != 0x32 {
					t.Errorf("expected R16 = 0x32, got 0x%02X", cpu.R[16])
				}
				if cpu.SREG&0b00000010 != 0 { // Z == 1?
					t.Errorf("expected Z=0, got Z=1")
				}
			},
		},
		{
			name:  "ASR R18",
			flash: []uint16{0x9525}, // ASR R18
			setup: func(cpu *CPU) {
				cpu.R[18] = 0b10000101 // MSB=1, LSB=1
			},
			expectCheck: func(cpu *CPU, t *testing.T) {
				want := uint8(0b11000010)
				if cpu.R[18] != want {
					t.Errorf("expected R18 = 0x%02X, got 0x%02X", want, cpu.R[18])
				}
				if cpu.SREG&0x01 == 0 { // C flag
					t.Errorf("expected Carry to be set")
				}
			},
		},
		{
			name:  "ADC R16,R17 with Carry",
			flash: []uint16{0x1F01}, // ADC R16, R17
			setup: func(cpu *CPU) {
				cpu.R[16] = 0x10       // Rd = 0x10
				cpu.R[17] = 0x11       // Rr = 0x11
				cpu.SREG = 0x01        // Set carry flag (C = 1)
			},
			expectCheck: func(cpu *CPU, t *testing.T) {
				want := uint8(0x10 + 0x11 + 1) // = 0x22
				if cpu.R[16] != want {
					t.Errorf("expected R16 = 0x%02X, got 0x%02X", want, cpu.R[16])
				}
				if cpu.SREG&0x02 != 0 { // Z = 0 → 2nd bit (0x02) should be cleared
					t.Errorf("expected Zero flag to be cleared")
				}
				if cpu.SREG&0x01 != 0 { // C = 0 → Carry should be cleared (no overflow)
					t.Errorf("expected Carry to be cleared")
				}
			},
		},
		{
			name:  "ADD R16, R17",
			flash: []uint16{0x0F01}, // ADD R16, R17 → 0000 1111 0000 0001
			setup: func(cpu *CPU) {
				cpu.R[16] = 0x10 // Rd = 0x10
				cpu.R[17] = 0x11 // Rr = 0x11
				cpu.SREG = 0x00  // Clear all flags (no carry)
			},
			expectCheck: func(cpu *CPU, t *testing.T) {
				want := uint8(0x10 + 0x11) // 0x21
				if cpu.R[16] != want {
					t.Errorf("expected R16 = 0x%02X, got 0x%02X", want, cpu.R[16])
				}
				if cpu.SREG&0x02 != 0 { // Z (Zero) should be cleared
					t.Errorf("expected Zero flag to be cleared")
				}
				if cpu.SREG&0x01 != 0 { // C (Carry) should be cleared
					t.Errorf("expected Carry flag to be cleared")
				}
			},
		},


	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &CPU{Flash: test.flash}
			test.setup(c)
			c.Step()
			test.expectCheck(c, t)
		})
	}
}
