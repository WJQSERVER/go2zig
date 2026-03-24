//go:build linux && amd64 && !cgo

#include "textflag.h"
#include "go_asm.h"

#define STACK_SIZE 80
#define PTR_ADDRESS (STACK_SIZE - 8)

GLOBL ·syscall15XABI0(SB), NOPTR|RODATA, $8
DATA ·syscall15XABI0(SB)/8, $syscall15X(SB)

TEXT syscall15X(SB), NOSPLIT, $STACK_SIZE
	MOVQ DI, PTR_ADDRESS(SP)
	MOVQ DI, R11

	MOVQ syscall15Args_f1(R11), X0
	MOVQ syscall15Args_f2(R11), X1
	MOVQ syscall15Args_f3(R11), X2
	MOVQ syscall15Args_f4(R11), X3
	MOVQ syscall15Args_f5(R11), X4
	MOVQ syscall15Args_f6(R11), X5
	MOVQ syscall15Args_f7(R11), X6
	MOVQ syscall15Args_f8(R11), X7

	MOVQ syscall15Args_a1(R11), DI
	MOVQ syscall15Args_a2(R11), SI
	MOVQ syscall15Args_a3(R11), DX
	MOVQ syscall15Args_a4(R11), CX
	MOVQ syscall15Args_a5(R11), R8
	MOVQ syscall15Args_a6(R11), R9

	MOVQ syscall15Args_a7(R11), R12
	MOVQ R12, 0(SP)
	MOVQ syscall15Args_a8(R11), R12
	MOVQ R12, 8(SP)
	MOVQ syscall15Args_a9(R11), R12
	MOVQ R12, 16(SP)
	MOVQ syscall15Args_a10(R11), R12
	MOVQ R12, 24(SP)
	MOVQ syscall15Args_a11(R11), R12
	MOVQ R12, 32(SP)
	MOVQ syscall15Args_a12(R11), R12
	MOVQ R12, 40(SP)
	MOVQ syscall15Args_a13(R11), R12
	MOVQ R12, 48(SP)
	MOVQ syscall15Args_a14(R11), R12
	MOVQ R12, 56(SP)
	MOVQ syscall15Args_a15(R11), R12
	MOVQ R12, 64(SP)
	XORL AX, AX

	MOVQ syscall15Args_fn(R11), R10
	CALL R10

	MOVQ PTR_ADDRESS(SP), DI
	MOVQ AX, syscall15Args_a1(DI)
	MOVQ DX, syscall15Args_a2(DI)
	MOVQ X0, syscall15Args_f1(DI)
	MOVQ X1, syscall15Args_f2(DI)
	XORL AX, AX
	RET
