//go:build ((linux || darwin) && arm64) && !cgo

#include "textflag.h"
#include "go_asm.h"

#define STACK_SIZE 64
#define PTR_ADDRESS (STACK_SIZE - 8)

GLOBL ·syscall15XABI0(SB), NOPTR|RODATA, $8
DATA ·syscall15XABI0(SB)/8, $syscall15X(SB)

TEXT syscall15X(SB), NOSPLIT, $0
	SUB $STACK_SIZE, RSP
	MOVD R0, PTR_ADDRESS(RSP)
	MOVD R0, R9

	FMOVD syscall15Args_f1(R9), F0
	FMOVD syscall15Args_f2(R9), F1
	FMOVD syscall15Args_f3(R9), F2
	FMOVD syscall15Args_f4(R9), F3
	FMOVD syscall15Args_f5(R9), F4
	FMOVD syscall15Args_f6(R9), F5
	FMOVD syscall15Args_f7(R9), F6
	FMOVD syscall15Args_f8(R9), F7

	MOVD syscall15Args_a1(R9), R0
	MOVD syscall15Args_a2(R9), R1
	MOVD syscall15Args_a3(R9), R2
	MOVD syscall15Args_a4(R9), R3
	MOVD syscall15Args_a5(R9), R4
	MOVD syscall15Args_a6(R9), R5
	MOVD syscall15Args_a7(R9), R6
	MOVD syscall15Args_a8(R9), R7
	MOVD syscall15Args_arm64_r8(R9), R8

	MOVD syscall15Args_a9(R9), R10
	MOVD R10, 0(RSP)
	MOVD syscall15Args_a10(R9), R10
	MOVD R10, 8(RSP)
	MOVD syscall15Args_a11(R9), R10
	MOVD R10, 16(RSP)
	MOVD syscall15Args_a12(R9), R10
	MOVD R10, 24(RSP)
	MOVD syscall15Args_a13(R9), R10
	MOVD R10, 32(RSP)
	MOVD syscall15Args_a14(R9), R10
	MOVD R10, 40(RSP)
	MOVD syscall15Args_a15(R9), R10
	MOVD R10, 48(RSP)

	MOVD syscall15Args_fn(R9), R10
	BL (R10)

	MOVD PTR_ADDRESS(RSP), R2
	ADD $STACK_SIZE, RSP

	MOVD R0, syscall15Args_a1(R2)
	MOVD R1, syscall15Args_a2(R2)
	FMOVD F0, syscall15Args_f1(R2)
	FMOVD F1, syscall15Args_f2(R2)
	FMOVD F2, syscall15Args_f3(R2)
	FMOVD F3, syscall15Args_f4(R2)
	RET
