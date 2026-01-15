package main

import (
	"fmt"

	"atmega/cpu"
	debugger "atmega/debuger"

	"github.com/rivo/tview"
)

func main() {
	// txtDebuger()
	uiDebuger()
}

func txtDebuger() {
	cpu := cpu.CPU{Flash: debugger.TestowyProgram()}
	dbg := debugger.Debugger{CPU: &cpu}

	for !cpu.Halt {
		dbg.PrintInstruction()
		cpu.Step()
		dbg.PrintRegisters()
	}

	fmt.Println("CPU halted.")
}

func uiDebuger() {
	app := tview.NewApplication()
	cpuInst := cpu.CPU{Flash: debugger.TestowyProgram()}
	// debug := debugger.Debugger{CPU: &cpuInst}

	runUiDebuger(app, cpuInst)
}

// func main() {
// 	app := tview.NewApplication()
// 	cpuInst := cpu.CPU{Flash: debugger.TestowyProgram()}
// 	// debug := debugger.Debugger{CPU: &cpuInst}

// 	runUiDebuger(app, cpuInst)
// }

func runUiDebuger(app *tview.Application, cpuInst cpu.CPU) {
	instrView := tview.NewList()
	statusView := tview.NewTextView().SetDynamicColors(true)

	var pinView *tview.TextView = tview.NewTextView()
	pinView.SetDynamicColors(true)
	pinView.SetTitle("Arduino Pins")
	pinView.SetBorder(true)

	var serialView *tview.TextView = tview.NewTextView()
	serialView.SetDynamicColors(true)
	serialView.SetTitle("serialView")
	serialView.SetBorder(true)

	// Z góry wykonaj wszystkie instrukcje do cache
	type CPUState struct {
		R            [32]uint8
		SREG         uint8
		PC           uint16
		IO           [64]uint8
		SerialOutput string
	}

	var states []CPUState
	cpuCopy := cpuInst
	for !cpuCopy.Halt && int(cpuCopy.PC) < len(cpuCopy.Flash) {
		instr := cpuCopy.Flash[cpuCopy.PC]
		addr := fmt.Sprintf("%04X", cpuCopy.PC)
		decoded := decodeInstr(instr)
	instrView.AddItem(fmt.Sprintf("%s: 0x%04X [%s]", addr, instr, decoded), "", 0, nil)


		cpuCopy.Step()
		states = append(states, CPUState{
			R:            cpuCopy.R,
			SREG:         cpuCopy.SREG,
			PC:           cpuCopy.PC,
			IO:           cpuCopy.IO,
			SerialOutput: cpuCopy.SerialOutput,
		})

	}

	instrView.SetChangedFunc(func(ix int, mainText, _ string, _ rune) {
		if ix >= len(states) {
			statusView.SetText("Brak danych.")
			return
		}
		s := states[ix]
		status := fmt.Sprintf("[yellow]PC: 0x%04X\nSREG: %08b\n", s.PC, s.SREG)
		for i := 0; i < 32; i++ {
			status += fmt.Sprintf("R%-2d: 0x%02X\t", i, s.R[i])
			if (i+1)%4 == 0 {
				status += "\n"
			}
		}
		statusView.SetText(status)

		// ==== Arduino PINs ====
		portB := s.IO[0x05] // PORTB
		ddrB := s.IO[0x04]  // DDRB
		portD := s.IO[0x0B] // PORTD
		ddrD := s.IO[0x0A]  // DDRD

		pinStatus := "[green]Digital Pins:\n"
		for i := 0; i < 8; i++ {
			state := "IN"
			if (ddrD>>i)&1 == 1 {
				if (portD>>i)&1 == 1 {
					state = "HIGH"
				} else {
					state = "LOW"
				}
			}
			pinStatus += fmt.Sprintf("D%-2d: %-5s  ", i, state)
		}
		pinStatus += "\n"
		for i := 0; i < 6; i++ {
			state := "IN"
			if (ddrB>>i)&1 == 1 {
				if (portB>>i)&1 == 1 {
					state = "HIGH"
				} else {
					state = "LOW"
				}
			}
			pinStatus += fmt.Sprintf("D%-2d: %-5s  ", i+8, state)
		}
		pinView.SetText(pinStatus)

		// ==== Serial COM ====
		serialView.SetText("[white]" + s.SerialOutput)

	})

	mainFlex := tview.NewFlex().
		AddItem(instrView, 40, 1, true).
		AddItem(statusView, 0, 2, false)

	bottomFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(pinView, 5, 1, false).
		AddItem(serialView, 4, 1, false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainFlex, 0, 3, true).   // 2/3 ekranu
		AddItem(bottomFlex, 0, 2, false) // 1/3 ekranu

	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}

func decodeInstr(instr uint16) string {
	switch {
	case instr&0xF000 == 0xE000: // LDI
		rd := 16 + ((instr >> 4) & 0xF)
		k := uint8((instr & 0x0F) | ((instr >> 4) & 0xF0))
		return fmt.Sprintf("LDI R%d, 0x%02X", rd, k)
	case instr&0xFC00 == 0x0C00: // ADD
		rd := (instr >> 4) & 0x1F
		rr := (instr & 0x0F) | ((instr >> 5) & 0x10)
		return fmt.Sprintf("ADD R%d, R%d", rd, rr)
	case instr&0xF800 == 0xB800: // OUT
		port := instr & 0x3F
		rr := (instr >> 4) & 0x1F
		return fmt.Sprintf("OUT 0x%02X, R%d", port, rr)
	case instr&0xFF00 == 0x9600: // ADIW
		rd := 24 + ((instr >> 4) & 0x6)
		k := uint16((instr & 0xF) | ((instr >> 2) & 0x30))
		return fmt.Sprintf("ADIW R%d:%d, %d", rd+1, rd, k)
	case instr&0xFC00 == 0x7000: // ANDI
		rd := 16 + ((instr >> 4) & 0xF) // R20
		k := uint8(((instr >> 4) & 0xF0) | (instr & 0xF0))
		return fmt.Sprintf("ANDI R%d, 0x%02X", rd, k)
	case instr&0xFC00 == 0x2000: // AND
		rd := (instr >> 4) & 0x1F
		rr := (instr & 0x0F) | ((instr >> 5) & 0x10)
		return fmt.Sprintf("AND R%d, R%d", rd, rr)
	case instr&0xFE0F == 0x9405: // ASR
		rd := (instr >> 4) & 0x1F
		return fmt.Sprintf("ASR R%d", rd)
	case instr&0xFC00 == 0x1C00: // ADC
		rd := ((instr >> 4) & 0x1F) | ((instr >> 8) & 0x20)
		rr := (instr & 0x0F) | ((instr >> 4) & 0x10)
		return fmt.Sprintf("[ADC] R%d ← R%d + R%d + C\n", rd, rd, rr)
	case instr == 0x0000:
		return "NOP"
	default:
		return "UNKNOWN"
	}
}

// ==============HELPER TOOLS===================================
// EncodeStringToLdiOut generuje instrukcje LDI/OUT zapisujące znaki na UART (port 0x0C)

/*
	Text to instructions
	usage:
	// prog := EncodeStringToLdiOut("Hello\n", 20)
	// DisasmProgram(prog)
*/
func EncodeStringToLdiOut(s string, reg int) []uint16 {
	if reg < 16 || reg > 31 {
		panic("LDI możliwe tylko dla R16–R31")
	}

	var program []uint16
	rdBits := uint16(reg - 16) // tylko 4 bity

	for _, ch := range s {
		k := uint8(ch)
		ldi := uint16(0xE000) |
			((uint16(k) & 0xF0) << 4) | // górna część K w bitach 11..8
			(rdBits << 4) |             // bity rejestru dddd
			(uint16(k) & 0x0F)          // dolne 4 bity K

		out := uint16(0xB800) |
			(uint16(reg) << 4) | // Rr
			(0x0C & 0x0F)        // port 0x0C

		program = append(program, ldi, out)
	}

	return program
}

func DisasmProgram(prog []uint16) {
	for _, instr := range prog {
		switch {
		case instr&0xF000 == 0xE000: // LDI
			rd := 16 + int((instr>>4)&0xF)
			k := ((instr >> 4) & 0xF0) | (instr & 0xF)
			fmt.Printf("0x%04X  ; LDI R%d, 0x%02X ('%c')\n", instr, rd, k, byte(k))

		case instr&0xF800 == 0xB800: // OUT
			port := instr & 0x3F
			rr := (instr >> 4) & 0x1F
			fmt.Printf("0x%04X  ; OUT 0x%02X, R%d\n", instr, port, rr)

		case instr == 0x0000:
			fmt.Printf("0x0000  ; NOP\n")

		default:
			fmt.Printf("0x%04X  ; ???\n", instr)
		}
	}
}
