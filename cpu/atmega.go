package cpu

import "fmt"

type CPU struct {
	R     [32]uint8   // Rejestry R0–R31
	SREG  uint8       // Status register
	PC    uint16      // Program Counter
	SRAM  [2048]uint8 // Pamięć danych
	Flash []uint16    // Pamięć programu (wczytywana z pliku .hex)
	Halt  bool        // Flaga zatrzymania
	// Arduino
	PortB, PortC, PortD uint8
	DDRB, DDRC, DDRD    uint8
	SerialOutput        string    // bufor do COM TX
	IO                  [64]uint8 // uproszczona przestrzeń I/O

}

func LoadHex(path string) ([]uint16, error) {
	// Parsowanie Intel HEX — można użyć biblioteki lub napisać własny parser
	return []uint16{0x940C, 0x0000 /* ... */}, nil
}

func (cpu *CPU) Step() {
	instr := cpu.Flash[cpu.PC]
	cpu.PC++

	switch {
	case instr&0xF000 == 0xE000: // LDI Rd, K
		rd := 16 + ((instr >> 4) & 0xF)
		k := ((uint8(instr>>4) & 0xF0) | uint8(instr&0x0F))
		cpu.R[rd] = k

		// fmt.Printf("[LDI] R%d ← 0x%02X ('%c')\n", rd, k, k)

	case instr&0xFC00 == 0x0C00: // ADD Rd,Rr
		rd := (instr >> 4) & 0x1F
		rr := (instr & 0x0F) | ((instr >> 5) & 0x10)
		result := cpu.R[rd] + cpu.R[rr]
		cpu.setFlagsAdd(cpu.R[rd], cpu.R[rr], result)
		cpu.R[rd] = result
	case instr&0xFF00 == 0x9600: // ADIW
		rd := 24 + ((instr >> 4) & 0x6) // 24, 26, 28, 30
		k := uint16((instr & 0xF) | ((instr >> 2) & 0x30))
		low := uint16(cpu.R[rd])
		high := uint16(cpu.R[rd+1])
		val := (high << 8) | low
		res := val + k

		cpu.R[rd] = uint8(res & 0xFF)
		cpu.R[rd+1] = uint8((res >> 8) & 0xFF)

		// Flagi
		r15 := (res >> 15) & 1
		rd7 := (val >> 15) & 1
		cpu.setFlag(1, res == 0)                                 // Z
		cpu.setFlag(0, (^val>>15)&1 != 0 && r15 != 0)            // C
		cpu.setFlag(2, r15 != 0)                                 // N
		cpu.setFlag(3, (rd7 == 1 && r15 == 0))                   // V
		cpu.setFlag(4, ((cpu.SREG>>2)&1)^((cpu.SREG>>3)&1) != 0) // S = N ^ V

	case instr&0xF800 == 0xB800: // OUT (1011 1AAr rrrr AAAA)
		port := uint8(instr & 0x3F) // 6-bit port
		rr := (instr >> 4) & 0x1F
		fmt.Printf("[DEBUG] OUT → port: 0x%02X  rr: R%d = 0x%02X\n", port, rr, cpu.R[rr])

		val := cpu.R[rr]

		// Zapisz do przestrzeni I/O
		cpu.IO[port] = val

		// Specjalne akcje
		switch port {
		case 0x0C:
			fmt.Printf("[UART] Write char='%c' hex=0x%02X\n", val, val)

			cpu.WriteSerial(val)
		}

	/**
		5.1 ADC – Add with Carry

		Opcode 16-bit:
		0001 11rd dddd rrrr

		example: ADC R16,R17
	**/
	case instr&0xFC00 == 0x1C00:
		rd := ((instr >> 4) & 0x1F) | ((instr >> 8) & 0x20)
		rr := (instr & 0x0F) | ((instr >> 4) & 0x10)

		Rd := cpu.R[rd]
		Rr := cpu.R[rr]
		C := cpu.SREG & 0x01 // Carry flag


		// fmt.Printf("[ADC] R%d ← R%d + R%d + C\n", rd, rd, rr)
		result := Rd + Rr + C
		cpu.R[rd] = result

		// Bity pomocnicze
		Rd3 := (Rd >> 3) & 1
		Rr3 := (Rr >> 3) & 1
		R3  := (result >> 3) & 1

		Rd7 := (Rd >> 7) & 1
		Rr7 := (Rr >> 7) & 1
		R7  := (result >> 7) & 1

		// Flagi
		H := (Rd3 & Rr3) | (Rr3 & R3) | (R3 & Rd3)
		V := (Rd7 & Rr7 & (^R7 & 1)) | ((^Rd7 & 1) & (^Rr7 & 1) & R7)
		N := R7
		Z := result == 0
		Cf := (Rd7 & Rr7) | (Rr7 & R7) | (R7 & Rd7)
		S := N ^ V

		// Ustawienie flag
		cpu.setFlag(0, Cf != 0)    // C
		cpu.setFlag(1, Z)          // Z
		cpu.setFlag(2, N != 0)     // N
		cpu.setFlag(3, V != 0)     // V
		cpu.setFlag(4, S != 0)     // S
		cpu.setFlag(5, H != 0)     // H

	/**
		5.2 ADD – Add without Carry

		oppcode:
		0000 11rd dddd rrrr

	**/
	case instr&0xFC00 == 0xC00:
		rd := (instr >> 4 & 0xF) | (instr >> 4) & 0x20
		rr := (instr & 0xF) | (instr >> 4) & 0x20
		
		Rd := cpu.R[rd]
		Rr := cpu.R[rr]

		result := Rd + Rr
		cpu.R[rd] = result

		// Set flags

		Rd3 := (Rd >> 3) & 1
		Rr3 := (Rr >> 3) & 1
		R3  := (result >> 3) & 1

		Rd7 := (Rd >> 7) & 1
		Rr7 := (Rr >> 7) & 1
		R7 := (result >> 7) & 1

		// (^R3 & 1)
		// bierzemy bit R3 wykonujemy negacje i bierzemy bit 0
		H := (Rd3 & Rr3) | (Rr3 & (^R3 & 1)) | ((^R3 & 1) & Rd3)
		V := (Rd7 & Rr7 & (^R7 & 1)) | ((^Rd7 & 1) & (^Rr7 & 1) & R7)
		N := R7
		Z := result == 0
		Cf := (Rd7 & Rr7) | (Rr7 & R7) | (R7 & Rd7)
		S := N ^ V

		cpu.setFlag(0, Cf != 0)   // C
		cpu.setFlag(1, Z)         // Z
		cpu.setFlag(2, N != 0)    // N
		cpu.setFlag(3, V != 0)    // V
		cpu.setFlag(4, S != 0)    // S
		cpu.setFlag(5, H != 0)    // H

	/**
		5.4 AND – Logical AND
		decoder OK

		16-bit Opcode:
		0010 00rd dddd rrrr

		0010 00 - instrukcja AND 

		max adres instrukcji
		1111 1100 0000 0000 = 0xFC00
		adres instrikcji AND
		0010 0000 0000 0000 = 0x2000

		operacja:
		Rd ← Rd ∧ Rr
		Rd = rd + rr

		Syntax:
		AND Rd,Rr
	**/
	case instr&0xFC00 == 0x2000: 
		// odpowiada dddd z opcode (4 boty)
		rd := (instr >> 4) & 0x1F

		// odpowiada rrrr z opcode (4 bity)
		rr := (instr & 0x0F) | ((instr >> 5) & 0x10)
		
		// wynik operacji AND
		result := cpu.R[rd] + cpu.R[rr]

		// ustawienie flag
		cpu.setFlagsLogic(result)
		cpu.R[rd] = result

	/**
		5.5 ANDI – Logical AND with Immediate
		decoder OK

		16-bit Opcode:
		0111 KKKK dddd KKKK

		operation:
		Rd ← Rd ∧ K

		wyjaśnienie:
		0x7000 = pierwsze 4 bity
		dddd = rejestr R16-R31

		bit:   15 14 13 12 | 11 10 9 8 | 7 6 5 4 | 3 2 1 0
		value:  0  1  1  1 |  K  K K K | d d d d | K K K K

	**/
	case instr&0xFC00 == 0x7000:
		// 16 + aby przesunąć się do dddd (offset)
		// wyciągnięcie wartości dddd z opcode || ((instr >> 4) & 0xF)
		rd := 16 + ((instr >> 4) & 0xF) // R20
		k := uint8(((instr >> 4) & 0xF0) | (instr & 0xF0))

		// Rd ← Rd ∧ K
		result := cpu.R[rd] & k
		cpu.R[rd] = result

		// ustawienie flag
		cpu.setFlagsLogic(result)

	/**
	=================================================================================
	EXAMPLE INSTRUCTION =============================================================
	=================================================================================
		5.6 ASR – Arithmetic Shift Right

		16-bit Opcode:
		1001 010d dddd 0101

		example ddddd:
		bin: 10010
		hex: 0x9525

		opis:
		przemieszcza ostatni bit na pozycję 0

		* wszystkie bity z opcoce które są = 1
		opcode = 1001 010d dddd 0101
		adres = 1001 0100 0000 0101 = 0x9405

		wybieranie bitów:
		0xFE0F - oznacza które bity mamy porównać dla adresu instrukcji
		0xFE0F = 1111 1110 0000 1111
		opcode = 1001 010d dddd 0101
		czyli porównujemy wszystkie bity oprócz tych oznaczonych przez litere d


	**/
	case instr&0xFE0F == 0x9405: // ASR

	// Przesuwamy instrukcję o 4 bity w prawo, aby wyrównać bity dddd,
    // następnie maskujemy 0x1F, aby wyciągnąć 5-bitowy numer rejestru (0–31).
	rd := (instr >> 4) & 0x1F


	old := cpu.R[rd]
	lsb := old & 0x01 // bit 0 (przed przesunięciem)
	msb := old & 0x80 // bit 7 (niezmieniony)

	// >> = przesunięcie w prawo (bitshift) + bit 7 bez zmiany
	result := (old >> 1) | msb
	cpu.R[rd] = result

	n := (result >> 7) & 1
	z := result == 0
	c := lsb != 0
	v := (n != 0) != c           // V = N ⊕ C
	s := (n != 0) != v           // S = N ⊕ V

	cpu.setFlag(0, c)      // C <- bit 0
	cpu.setFlag(1, z)      // Z
	cpu.setFlag(2, n != 0) // N
	cpu.setFlag(3, v)      // V
	cpu.setFlag(4, s)      // S

	case instr == 0x0000: // NOP
		// do nothing
	default:
		fmt.Printf("Nieznana instrukcja: 0x%04X\n", instr)
		cpu.Halt = true
	}

}

func (cpu *CPU) WriteSerial(b byte) {
	cpu.SerialOutput += string(b)
}

