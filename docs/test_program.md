# program:

0xE010 = 1110 0000 0001 0000

→ KKKK = 0000 (bity 7-4)
→ dddd = 0001 (czyli R17)
→ KKKK (dolne) = 0000

//
LDI R20, 0x48 ('H') → dddd = 0100
→ KKKK KKKK = 0100 1000 = 0x48
→ Instrukcja: 1110 0100 0100 1000 = 0xE448