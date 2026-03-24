//go:build arm64

#include "textflag.h"

#define G0ASMCALL_NORET                                 \
	MOVD	RSP, R4                                    \
	MOVD	0x30(g), R3                               \
	MOVD	0x0(R3), R3                               \
	MOVD	g, R5                                      \
	MOVD	R3, g                                      \
	MOVD	0x38(R3), R3                               \
	AND	$~15, R3                                    \
	MOVD	R3, RSP                                    \
	SUB	$0x20, RSP                                  \
	MOVD	R30, 0x10(RSP)                             \
	STP	(R5, R4), (RSP)                             \
	CALL	R8                                          \
	LDP	(RSP), (g, R3)                              \
	MOVD	0x10(RSP), R30                             \
	MOVD	R3, RSP

#define G0ASMCALL_R1 G0ASMCALL_NORET

#define ASMCALL_NORET                                  \
	CALL	R8                                          \
	RET

#define ASMCALL_R1 CALL	R8

TEXT ·CallFuncG0P0(SB), NOSPLIT|NOPTR|NOFRAME, $0-8
	MOVD	fn+0x0(FP), R8
	G0ASMCALL_NORET
	RET

TEXT ·CallFuncG0P0R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-16
	MOVD	fn+0x0(FP), R8
	G0ASMCALL_R1
	MOVD	R0, ret+8(FP)
	RET

TEXT ·CallFuncG0P1(SB), NOSPLIT|NOPTR|NOFRAME, $0-16
	MOVD	fn+0x0(FP), R8
	MOVD	arg0+0x8(FP), R0
	G0ASMCALL_NORET
	RET

TEXT ·CallFuncG0P1R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-24
	MOVD	fn+0x0(FP), R8
	MOVD	arg0+0x8(FP), R0
	G0ASMCALL_R1
	MOVD	R0, ret+16(FP)
	RET

TEXT ·CallFuncG0P2(SB), NOSPLIT|NOPTR|NOFRAME, $0-24
	MOVD	fn+0x0(FP), R8
	MOVD	arg0+0x8(FP), R0
	MOVD	arg1+0x10(FP), R1
	G0ASMCALL_NORET
	RET

TEXT ·CallFuncG0P2StoreR1(SB), NOSPLIT|NOPTR|NOFRAME, $0-32
	MOVD	fn+0x0(FP), R8
	MOVD	arg0+0x8(FP), R0
	MOVD	arg1+0x10(FP), R1
	G0ASMCALL_R1
	MOVD	out+0x18(FP), R9
	CBZ	R9, done_g0_p2_store_r1
	MOVD	R0, (R9)
done_g0_p2_store_r1:
	RET

TEXT ·CallFuncG0P2R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-32
	MOVD	fn+0x0(FP), R8
	MOVD	arg0+0x8(FP), R0
	MOVD	arg1+0x10(FP), R1
	G0ASMCALL_R1
	MOVD	R0, ret+24(FP)
	RET

TEXT ·CallFuncG0P3(SB), NOSPLIT|NOPTR|NOFRAME, $0-32
	MOVD	fn+0x0(FP), R8
	MOVD	arg0+0x8(FP), R0
	MOVD	arg1+0x10(FP), R1
	MOVD	arg2+0x18(FP), R2
	G0ASMCALL_NORET
	RET

TEXT ·CallFuncG0P3R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-40
	MOVD	fn+0x0(FP), R8
	MOVD	arg0+0x8(FP), R0
	MOVD	arg1+0x10(FP), R1
	MOVD	arg2+0x18(FP), R2
	G0ASMCALL_R1
	MOVD	R0, ret+32(FP)
	RET

TEXT ·CallFuncP0(SB), NOSPLIT|NOPTR|NOFRAME, $0-8
	MOVD	fn+0x0(FP), R8
	ASMCALL_NORET

TEXT ·CallFuncP0R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-16
	MOVD	fn+0x0(FP), R8
	ASMCALL_R1
	MOVD	R0, ret+8(FP)
	RET

TEXT ·CallFuncP1(SB), NOSPLIT|NOPTR|NOFRAME, $0-16
	MOVD	fn+0x0(FP), R8
	MOVD	arg0+0x8(FP), R0
	ASMCALL_NORET

TEXT ·CallFuncP1R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-24
	MOVD	fn+0x0(FP), R8
	MOVD	arg0+0x8(FP), R0
	ASMCALL_R1
	MOVD	R0, ret+16(FP)
	RET

TEXT ·CallFuncP2(SB), NOSPLIT|NOPTR|NOFRAME, $0-24
	MOVD	fn+0x0(FP), R8
	MOVD	arg0+0x8(FP), R0
	MOVD	arg1+0x10(FP), R1
	ASMCALL_NORET

TEXT ·CallFuncP2R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-32
	MOVD	fn+0x0(FP), R8
	MOVD	arg0+0x8(FP), R0
	MOVD	arg1+0x10(FP), R1
	ASMCALL_R1
	MOVD	R0, ret+24(FP)
	RET

TEXT ·CallFuncP3(SB), NOSPLIT|NOPTR|NOFRAME, $0-32
	MOVD	fn+0x0(FP), R8
	MOVD	arg0+0x8(FP), R0
	MOVD	arg1+0x10(FP), R1
	MOVD	arg2+0x18(FP), R2
	ASMCALL_NORET

TEXT ·CallFuncP3R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-40
	MOVD	fn+0x0(FP), R8
	MOVD	arg0+0x8(FP), R0
	MOVD	arg1+0x10(FP), R1
	MOVD	arg2+0x18(FP), R2
	ASMCALL_R1
	MOVD	R0, ret+32(FP)
	RET
