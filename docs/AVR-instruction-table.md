# AVR
> how to implement instruction set from manual table to code

// dodawanie instrukcji:

(plik /instructions/asr.go)

```go
/**
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

    //!!!!! jeżeli we flagach w dokumentacja jest kreska nad wartością to trzeba wykonać:
    // dla R& negowanego: (^R7 & 1)
    // wykonuje to negacje i pobranie 1 bitu z wartości

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

```