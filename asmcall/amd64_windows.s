//go:build windows && amd64

#include "textflag.h"

#define RARG0 CX
#define RARG1 DX
#define RARG2 R8
#define RRET AX
#define RTMP0 R9
#define RTMP1 R10
#define RTMP2 R11
#define SHADOW_SPACE SUBQ $32, SP
#define RESTORE_SHADOW ADDQ $32, SP

#define G0ASMCALL_NORET                                      \
    MOVQ    SP, RTMP0                                        \
    MOVQ    0x30(g), RTMP1                                   \
    MOVQ    0x0(RTMP1), RTMP1                                \
    MOVQ    g, RTMP2                                         \
    MOVQ    RTMP1, g                                         \
    MOVQ    0x38(RTMP1), SP                                  \
    ANDQ    $-16, SP                                         \
    PUSHQ   RTMP0                                            \
    PUSHQ   RTMP2                                            \
    SHADOW_SPACE                                             \
    CALL    AX                                               \
    RESTORE_SHADOW                                           \
    POPQ    g                                                \
    POPQ    SP                                               \
    RET

#define G0ASMCALL_R1                                         \
    MOVQ    SP, RTMP0                                        \
    MOVQ    0x30(g), RTMP1                                   \
    MOVQ    0x0(RTMP1), RTMP1                                \
    MOVQ    g, RTMP2                                         \
    MOVQ    RTMP1, g                                         \
    MOVQ    0x38(RTMP1), SP                                  \
    ANDQ    $-16, SP                                         \
    PUSHQ   RTMP0                                            \
    PUSHQ   RTMP2                                            \
    SHADOW_SPACE                                             \
    CALL    AX                                               \
    RESTORE_SHADOW                                           \
    POPQ    g                                                \
    POPQ    SP

#define ASMCALL_NORET                                        \
    SHADOW_SPACE                                             \
    CALL    AX                                               \
    RESTORE_SHADOW                                           \
    RET

#define ASMCALL_R1                                           \
    SHADOW_SPACE                                             \
    CALL    AX                                               \
    RESTORE_SHADOW

TEXT ·CallFuncG0P0(SB), NOSPLIT|NOPTR|NOFRAME, $0
	MOVQ	fn+0x0(FP), AX
	G0ASMCALL_NORET

TEXT ·CallFuncG0P0R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-16
	MOVQ	fn+0x0(FP), AX
	G0ASMCALL_R1
	MOVQ	RRET, ret+8(FP)
	RET

TEXT ·CallFuncG0P1(SB), NOSPLIT|NOPTR|NOFRAME, $0
	MOVQ	fn+0x0(FP), AX
	MOVQ	arg0+0x8(FP), RARG0
	G0ASMCALL_NORET

TEXT ·CallFuncG0P1R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-24
	MOVQ	fn+0x0(FP), AX
	MOVQ	arg0+0x8(FP), RARG0
	G0ASMCALL_R1
	MOVQ	RRET, ret+16(FP)
	RET

TEXT ·CallFuncG0P2(SB), NOSPLIT|NOPTR|NOFRAME, $0
	MOVQ	fn+0x0(FP), AX
	MOVQ	arg0+0x8(FP), RARG0
	MOVQ	arg1+0x10(FP), RARG1
	G0ASMCALL_NORET

TEXT ·CallFuncG0P2StoreR1(SB), NOSPLIT|NOPTR|NOFRAME, $0
	MOVQ	fn+0x0(FP), AX
	MOVQ	arg0+0x8(FP), RARG0
	MOVQ	arg1+0x10(FP), RARG1
	MOVQ	out+0x18(FP), BX
	MOVQ	SP, RTMP0
	MOVQ	0x30(g), RTMP1
	MOVQ	0x0(RTMP1), RTMP1
	MOVQ	g, RTMP2
	MOVQ	RTMP1, g
	MOVQ	0x38(RTMP1), SP
	ANDQ	$-16, SP
	PUSHQ	RTMP0
	PUSHQ	RTMP2
	PUSHQ	BX
	SUBQ	$8, SP
	SHADOW_SPACE
	CALL	AX
	RESTORE_SHADOW
	ADDQ	$8, SP
	POPQ	BX
	MOVQ	RRET, 0(BX)
	POPQ	g
	POPQ	SP
	RET

TEXT ·CallFuncG0P2R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-32
	MOVQ	fn+0x0(FP), AX
	MOVQ	arg0+0x8(FP), RARG0
	MOVQ	arg1+0x10(FP), RARG1
	G0ASMCALL_R1
	MOVQ	RRET, ret+24(FP)
	RET

TEXT ·CallFuncG0P3(SB), NOSPLIT|NOPTR|NOFRAME, $0
	MOVQ	fn+0x0(FP), AX
	MOVQ	arg0+0x8(FP), RARG0
	MOVQ	arg1+0x10(FP), RARG1
	MOVQ	arg2+0x18(FP), RARG2
	G0ASMCALL_NORET

TEXT ·CallFuncG0P3R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-40
	MOVQ	fn+0x0(FP), AX
	MOVQ	arg0+0x8(FP), RARG0
	MOVQ	arg1+0x10(FP), RARG1
	MOVQ	arg2+0x18(FP), RARG2
	G0ASMCALL_R1
	MOVQ	RRET, ret+32(FP)
	RET

TEXT ·CallFuncP0(SB), NOSPLIT|NOPTR|NOFRAME, $0
	MOVQ	fn+0x0(FP), AX
	ASMCALL_NORET

TEXT ·CallFuncP0R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-16
	MOVQ	fn+0x0(FP), AX
	ASMCALL_R1
	MOVQ	RRET, ret+8(FP)
	RET

TEXT ·CallFuncP1(SB), NOSPLIT|NOPTR|NOFRAME, $0
	MOVQ	fn+0x0(FP), AX
	MOVQ	arg0+0x8(FP), RARG0
	ASMCALL_NORET

TEXT ·CallFuncP1R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-24
	MOVQ	fn+0x0(FP), AX
	MOVQ	arg0+0x8(FP), RARG0
	ASMCALL_R1
	MOVQ	RRET, ret+16(FP)
	RET

TEXT ·CallFuncP2(SB), NOSPLIT|NOPTR|NOFRAME, $0
	MOVQ	fn+0x0(FP), AX
	MOVQ	arg0+0x8(FP), RARG0
	MOVQ	arg1+0x10(FP), RARG1
	ASMCALL_NORET

TEXT ·CallFuncP2R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-32
	MOVQ	fn+0x0(FP), AX
	MOVQ	arg0+0x8(FP), RARG0
	MOVQ	arg1+0x10(FP), RARG1
	ASMCALL_R1
	MOVQ	RRET, ret+24(FP)
	RET

TEXT ·CallFuncP3(SB), NOSPLIT|NOPTR|NOFRAME, $0
	MOVQ	fn+0x0(FP), AX
	MOVQ	arg0+0x8(FP), RARG0
	MOVQ	arg1+0x10(FP), RARG1
	MOVQ	arg2+0x18(FP), RARG2
	ASMCALL_NORET

TEXT ·CallFuncP3R1(SB), NOSPLIT|NOPTR|NOFRAME, $0-40
	MOVQ	fn+0x0(FP), AX
	MOVQ	arg0+0x8(FP), RARG0
	MOVQ	arg1+0x10(FP), RARG1
	MOVQ	arg2+0x18(FP), RARG2
	ASMCALL_R1
	MOVQ	RRET, ret+32(FP)
	RET
