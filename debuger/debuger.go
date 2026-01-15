// Package debugger: umożliwia debugowanie emulatora AVR
package debugger

import (
	"atmega/cpu"
	"fmt"
)

// PrintRegisters wypisuje zawartość rejestrów R0–R31 i SREG
type Debugger struct {
	CPU *cpu.CPU
}

func (d *Debugger) PrintRegisters() {
	fmt.Println("==== Rejestry ====")
	for i := 0; i < 32; i++ {
		fmt.Printf("R%-2d: 0x%02X\t", i, d.CPU.R[i])
		if (i+1)%4 == 0 {
			fmt.Println()
		}
	}
	fmt.Printf("\nPC:  0x%04X\n", d.CPU.PC)
	fmt.Printf("SREG: %08b [", d.CPU.SREG)
	flags := []string{"C", "Z", "N", "V", "S", "H", "T", "I"}
	for i := 7; i >= 0; i-- {
		if (d.CPU.SREG>>i)&1 == 1 {
			fmt.Printf("%s ", flags[7-i])
		}
	}
	fmt.Println("]")
	fmt.Println("==============")
}

// PrintInstruction wypisuje kolejną instrukcję do wykonania
func (d *Debugger) PrintInstruction() {
	if int(d.CPU.PC) >= len(d.CPU.Flash) {
		fmt.Println("[EOF] Brak instrukcji")
		return
	}
	instr := d.CPU.Flash[d.CPU.PC]
	fmt.Printf("Instrukcja @PC=0x%04X: 0x%04X\n", d.CPU.PC, instr)
}

// ======= Poniżej prosty program testowy =======

// TestowyProgram zwraca prosty program AVR jako []uint16
func TestowyProgram() []uint16 {
	return []uint16{
		0x1C00,
		0xE100, // LDI R16, 0x10
		0xE111, // LDI R17, 0x11
		0x1F01, // ADC R16, R17
		0x0000, // NOP
	}
}

