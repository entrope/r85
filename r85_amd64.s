//go:build amd64 && !purego

#include "textflag.h"

// ===== RODATA constants (32 bytes each for YMM compatibility) =====

// Byte-swap: little-endian <-> big-endian uint32
DATA bswap32<>+0(SB)/8, $0x0405060700010203
DATA bswap32<>+8(SB)/8, $0x0C0D0E0F08090A0B
DATA bswap32<>+16(SB)/8, $0x0405060700010203
DATA bswap32<>+24(SB)/8, $0x0C0D0E0F08090A0B
GLOBL bswap32<>(SB), NOPTR|RODATA, $32

// Magic multiplier M = floor(2^38/85) = 0xC0C0C0C1
DATA magic85<>+0(SB)/4, $0xC0C0C0C1
DATA magic85<>+4(SB)/4, $0xC0C0C0C1
DATA magic85<>+8(SB)/4, $0xC0C0C0C1
DATA magic85<>+12(SB)/4, $0xC0C0C0C1
DATA magic85<>+16(SB)/4, $0xC0C0C0C1
DATA magic85<>+20(SB)/4, $0xC0C0C0C1
DATA magic85<>+24(SB)/4, $0xC0C0C0C1
DATA magic85<>+28(SB)/4, $0xC0C0C0C1
GLOBL magic85<>(SB), NOPTR|RODATA, $32

// Broadcast 85 as uint32
DATA const85d<>+0(SB)/4, $85
DATA const85d<>+4(SB)/4, $85
DATA const85d<>+8(SB)/4, $85
DATA const85d<>+12(SB)/4, $85
DATA const85d<>+16(SB)/4, $85
DATA const85d<>+20(SB)/4, $85
DATA const85d<>+24(SB)/4, $85
DATA const85d<>+28(SB)/4, $85
GLOBL const85d<>(SB), NOPTR|RODATA, $32

// Extract low byte of each uint32 lane
DATA packLow<>+0(SB)/8, $0x808080800C080400
DATA packLow<>+8(SB)/8, $0x8080808080808080
DATA packLow<>+16(SB)/8, $0x808080800C080400
DATA packLow<>+24(SB)/8, $0x8080808080808080
GLOBL packLow<>(SB), NOPTR|RODATA, $32

// Encode: stride-5 interleave d0-d3, first 16 output bytes per lane
DATA encShufMain<>+0(SB)/8, $0x090501800C080400
DATA encShufMain<>+8(SB)/8, $0x03800E0A0602800D
DATA encShufMain<>+16(SB)/8, $0x090501800C080400
DATA encShufMain<>+24(SB)/8, $0x03800E0A0602800D
GLOBL encShufMain<>(SB), NOPTR|RODATA, $32

// Encode: stride-5 d4 positions in first 16 output bytes per lane
DATA encShufD4<>+0(SB)/8, $0x8080800080808080
DATA encShufD4<>+8(SB)/8, $0x8002808080800180
DATA encShufD4<>+16(SB)/8, $0x8080800080808080
DATA encShufD4<>+24(SB)/8, $0x8002808080800180
GLOBL encShufD4<>(SB), NOPTR|RODATA, $32

// Encode: stride-5 d0-d3, last 4 output bytes per lane
DATA encShufHi<>+0(SB)/8, $0x80808080800F0B07
DATA encShufHi<>+8(SB)/8, $0x8080808080808080
DATA encShufHi<>+16(SB)/8, $0x80808080800F0B07
DATA encShufHi<>+24(SB)/8, $0x8080808080808080
GLOBL encShufHi<>(SB), NOPTR|RODATA, $32

// Encode: stride-5 d4, last 4 output bytes per lane
DATA encShufD4Hi<>+0(SB)/8, $0x8080808003808080
DATA encShufD4Hi<>+8(SB)/8, $0x8080808080808080
DATA encShufD4Hi<>+16(SB)/8, $0x8080808003808080
DATA encShufD4Hi<>+24(SB)/8, $0x8080808080808080
GLOBL encShufD4Hi<>(SB), NOPTR|RODATA, $32

// Byte broadcasts (32 bytes each)
DATA const40b<>+0(SB)/8, $0x2828282828282828
DATA const40b<>+8(SB)/8, $0x2828282828282828
DATA const40b<>+16(SB)/8, $0x2828282828282828
DATA const40b<>+24(SB)/8, $0x2828282828282828
GLOBL const40b<>(SB), NOPTR|RODATA, $32

DATA const20b<>+0(SB)/8, $0x1414141414141414
DATA const20b<>+8(SB)/8, $0x1414141414141414
DATA const20b<>+16(SB)/8, $0x1414141414141414
DATA const20b<>+24(SB)/8, $0x1414141414141414
GLOBL const20b<>(SB), NOPTR|RODATA, $32

DATA const56b<>+0(SB)/8, $0x3838383838383838
DATA const56b<>+8(SB)/8, $0x3838383838383838
DATA const56b<>+16(SB)/8, $0x3838383838383838
DATA const56b<>+24(SB)/8, $0x3838383838383838
GLOBL const56b<>(SB), NOPTR|RODATA, $32

DATA const65b<>+0(SB)/8, $0x4141414141414141
DATA const65b<>+8(SB)/8, $0x4141414141414141
DATA const65b<>+16(SB)/8, $0x4141414141414141
DATA const65b<>+24(SB)/8, $0x4141414141414141
GLOBL const65b<>(SB), NOPTR|RODATA, $32

DATA const30b<>+0(SB)/8, $0x1E1E1E1E1E1E1E1E
DATA const30b<>+8(SB)/8, $0x1E1E1E1E1E1E1E1E
DATA const30b<>+16(SB)/8, $0x1E1E1E1E1E1E1E1E
DATA const30b<>+24(SB)/8, $0x1E1E1E1E1E1E1E1E
GLOBL const30b<>(SB), NOPTR|RODATA, $32

DATA const85b<>+0(SB)/8, $0x5555555555555555
DATA const85b<>+8(SB)/8, $0x5555555555555555
DATA const85b<>+16(SB)/8, $0x5555555555555555
DATA const85b<>+24(SB)/8, $0x5555555555555555
GLOBL const85b<>(SB), NOPTR|RODATA, $32

DATA const86b<>+0(SB)/8, $0x5656565656565656
DATA const86b<>+8(SB)/8, $0x5656565656565656
DATA const86b<>+16(SB)/8, $0x5656565656565656
DATA const86b<>+24(SB)/8, $0x5656565656565656
GLOBL const86b<>(SB), NOPTR|RODATA, $32

// Decode-only masks (16 bytes, used with XMM only)
DATA decShufMain<>+0(SB)/8, $0x800B06010F0A0500
DATA decShufMain<>+8(SB)/8, $0x800D0803800C0702
GLOBL decShufMain<>(SB), NOPTR|RODATA, $16

DATA decShufFill<>+0(SB)/8, $0x0080808080808080
DATA decShufFill<>+8(SB)/8, $0x0280808001808080
GLOBL decShufFill<>(SB), NOPTR|RODATA, $16

DATA decShufD4<>+0(SB)/8, $0x80808080800E0904
DATA decShufD4<>+8(SB)/8, $0x8080808080808080
GLOBL decShufD4<>(SB), NOPTR|RODATA, $16

DATA decShufD4Fill<>+0(SB)/8, $0x8080808003808080
DATA decShufD4Fill<>+8(SB)/8, $0x8080808080808080
GLOBL decShufD4Fill<>(SB), NOPTR|RODATA, $16

// Widen: zero-extend bytes 0-3 to uint32 lanes (32 bytes for YMM)
DATA widen0<>+0(SB)/8, $0x8080800180808000
DATA widen0<>+8(SB)/8, $0x8080800380808002
DATA widen0<>+16(SB)/8, $0x8080800180808000
DATA widen0<>+24(SB)/8, $0x8080800380808002
GLOBL widen0<>(SB), NOPTR|RODATA, $32

// Mask: even uint32 lanes as uint64 (32 bytes for YMM)
DATA maskEven<>+0(SB)/8, $0x00000000FFFFFFFF
DATA maskEven<>+8(SB)/8, $0x00000000FFFFFFFF
DATA maskEven<>+16(SB)/8, $0x00000000FFFFFFFF
DATA maskEven<>+24(SB)/8, $0x00000000FFFFFFFF
GLOBL maskEven<>(SB), NOPTR|RODATA, $32

// ===== Macros =====

// DIV85_YMM: q = acc / 85, r = acc % 85 (YMM, 8 lanes)
// Clobbers Y1, Y2, Y3.  Uses Y12 (magic85), Y13 (const85d).
#define DIV85_YMM(Y_acc, Y_q, Y_r) \
	VPSHUFD	$0xF5, Y_acc, Y1;       \
	VPMULUDQ	Y12, Y_acc, Y2;     \
	VPMULUDQ	Y12, Y1, Y3;        \
	VPSRLQ	$38, Y2, Y2;            \
	VPSRLQ	$38, Y3, Y3;            \
	VPSLLQ	$32, Y3, Y3;            \
	VPOR	Y2, Y3, Y_q;            \
	VPMULLD	Y13, Y_q, Y1;          \
	VPSUBD	Y1, Y_acc, Y_r

// DIGIT_TO_CHAR_YMM: convert digit bytes to r85 chars (YMM).
// Modifies Y_data in place.  Clobbers Y1, Y2.
#define DIGIT_TO_CHAR_YMM(Y_data) \
	VMOVDQU	const20b<>(SB), Y1;                  \
	VPCMPEQB	Y1, Y_data, Y1;              \
	VPAND	const65b<>(SB), Y1, Y1;              \
	VMOVDQU	const56b<>(SB), Y2;                  \
	VPCMPEQB	Y2, Y_data, Y2;              \
	VPAND	const30b<>(SB), Y2, Y2;              \
	VPOR	Y1, Y2, Y1;                          \
	VPADDB	const40b<>(SB), Y_data, Y_data;      \
	VPADDB	Y1, Y_data, Y_data

// CHAR_TO_DIGIT: convert r85 chars to digit bytes (XMM).
// Modifies X_data in place.  Clobbers X1, X2.
// Uses VEX-encoded loads (VMOVDQU) to avoid false dependencies
// on upper YMM bits from previous loop iterations.
#define CHAR_TO_DIGIT(X_data) \
	VPSUBB	const40b<>(SB), X_data, X_data;      \
	VMOVDQU	const85b<>(SB), X1;                  \
	VPCMPEQB	X1, X_data, X1;              \
	VPAND	const65b<>(SB), X1, X1;              \
	VPSUBB	X1, X_data, X_data;                  \
	VMOVDQU	const86b<>(SB), X2;                  \
	VPCMPEQB	X2, X_data, X2;              \
	VPAND	const30b<>(SB), X2, X2;              \
	VPSUBB	X2, X_data, X_data

// ===== encodeBlocksAVX2 =====
// func encodeBlocksAVX2(dst *byte, src *byte)
// 2 iterations x 8 uint32 lanes (YMM). 64 bytes in -> 80 bytes out.
TEXT ·encodeBlocksAVX2(SB), NOSPLIT|NOFRAME, $0-16
	MOVQ	dst+0(FP), DI
	MOVQ	src+8(FP), SI

	VMOVDQU	bswap32<>(SB), Y11
	VMOVDQU	magic85<>(SB), Y12
	VMOVDQU	const85d<>(SB), Y13

	MOVQ	$2, CX

enc_loop:
	// Load 32 bytes, byte-swap to big-endian uint32s
	VMOVDQU	(SI), Y0
	VPSHUFB	Y11, Y0, Y0

	// 4 rounds of div-85: value -> (d0, d1, d2, d3, d4)
	DIV85_YMM(Y0, Y4, Y6)      // d4=Y6, q=Y4
	DIV85_YMM(Y4, Y5, Y7)      // d3=Y7, q=Y5
	DIV85_YMM(Y5, Y4, Y8)      // d2=Y8, q=Y4
	DIV85_YMM(Y4, Y9, Y10)     // d1=Y10, d0=Y9

	// Pack uint32 digits to bytes (lane-local packing)
	VPACKUSDW	Y10, Y9, Y0    // [d0,d1] per lane as uint16
	VPACKUSDW	Y7, Y8, Y4     // [d2,d3] per lane as uint16
	VPACKUSWB	Y4, Y0, Y0     // [d0..d3] per lane as bytes

	// Pack d4: extract low byte of each uint32
	VPSHUFB	packLow<>(SB), Y6, Y4

	// Stride-5 interleave (lane-local, same masks both lanes)
	VPSHUFB	encShufMain<>(SB), Y0, Y5
	VPSHUFB	encShufD4<>(SB), Y4, Y14
	VPOR	Y5, Y14, Y5              // first 16 output bytes per lane

	VPSHUFB	encShufHi<>(SB), Y0, Y14
	VPSHUFB	encShufD4Hi<>(SB), Y4, Y6
	VPOR	Y14, Y6, Y14             // last 4 output bytes per lane

	// Digit-to-char conversion (YMM, all 32 bytes at once)
	DIGIT_TO_CHAR_YMM(Y5)
	DIGIT_TO_CHAR_YMM(Y14)

	// Store: low lane = groups 0-3 (20 bytes), high lane = groups 4-7 (20 bytes)
	MOVOU	X5, (DI)
	MOVL	X14, 16(DI)
	VEXTRACTI128	$1, Y5, X5
	VEXTRACTI128	$1, Y14, X14
	MOVOU	X5, 20(DI)
	MOVL	X14, 36(DI)

	ADDQ	$32, SI
	ADDQ	$40, DI
	DECQ	CX
	JNZ	enc_loop

	VZEROUPPER
	RET

// ===== decodeBlocksAVX2 =====
// func decodeBlocksAVX2(dst *byte, src *byte) uint64
// 4 iterations x 4 uint32 lanes (XMM). 80 bytes in -> 64 bytes out.
// XMM is optimal for decode: 20-byte input chunks don't benefit from
// YMM widening, and XMM allows better pipelining between iterations.
TEXT ·decodeBlocksAVX2(SB), NOSPLIT|NOFRAME, $0-24
	MOVQ	dst+0(FP), DI
	MOVQ	src+8(FP), SI

	VMOVDQU	bswap32<>(SB), X11
	VMOVDQU	const85d<>(SB), X12
	VMOVDQU	widen0<>(SB), X13
	VMOVDQU	maskEven<>(SB), X14
	VPXOR	X15, X15, X15          // overflow accumulator

	MOVQ	$4, CX

dec_loop:
	// Load 20 bytes: 16 + 4
	VMOVDQU	(SI), X0
	MOVL	16(SI), AX
	VPXOR	X4, X4, X4
	MOVL	AX, X4

	// Char-to-digit conversion
	CHAR_TO_DIGIT(X0)
	CHAR_TO_DIGIT(X4)

	// Deinterleave: extract d0-d3 and d4 from stride-5 layout
	VPSHUFB	decShufMain<>(SB), X0, X5
	VPSHUFB	decShufFill<>(SB), X4, X6
	VPOR	X5, X6, X5              // d0-d3 transposed

	VPSHUFB	decShufD4<>(SB), X0, X6
	VPSHUFB	decShufD4Fill<>(SB), X4, X7
	VPOR	X6, X7, X6              // d4

	// Widen digit bytes to uint32 lanes
	VPSHUFB	X13, X5, X7             // d0 as uint32x4
	VPSHUFD	$0x39, X5, X0
	VPSHUFB	X13, X0, X8             // d1 as uint32x4
	VPSHUFD	$0x4E, X5, X0
	VPSHUFB	X13, X0, X9             // d2 as uint32x4
	VPSHUFD	$0x93, X5, X0
	VPSHUFB	X13, X0, X10            // d3 as uint32x4
	VPSHUFB	X13, X6, X4             // d4 as uint32x4

	// Horner: acc = ((d0*85 + d1)*85 + d2)*85 + d3  (32-bit)
	VPMULLD	X12, X7, X0
	VPADDD	X8, X0, X0
	VPMULLD	X12, X0, X0
	VPADDD	X9, X0, X0
	VPMULLD	X12, X0, X0
	VPADDD	X10, X0, X0

	// Final 64-bit: acc*85 + d4  (detect overflow)
	VPSHUFD	$0xF5, X0, X1           // odd lanes to even
	VPMULUDQ	X12, X0, X2         // even lanes * 85 -> uint64
	VPMULUDQ	X12, X1, X3         // odd lanes * 85 -> uint64

	VPAND	X14, X4, X5              // d4 even lanes
	VPSRLQ	$32, X4, X6             // d4 odd lanes
	VPADDQ	X5, X2, X2
	VPADDQ	X6, X3, X3

	// Overflow check: any high 32 bits nonzero?
	VPSRLQ	$32, X2, X5
	VPSRLQ	$32, X3, X6
	VPOR	X5, X6, X5
	VPOR	X5, X15, X15

	// Merge low 32 bits back to uint32 lanes
	VSHUFPS	$0x88, X3, X2, X0       // [lo0, lo2, lo1, lo3]
	VPSHUFD	$0xD8, X0, X0           // [lo0, lo1, lo2, lo3]

	// Byte-swap to big-endian
	VPSHUFB	X11, X0, X0

	// Store 16 output bytes
	VMOVDQU	X0, (DI)

	ADDQ	$20, SI
	ADDQ	$16, DI
	DECQ	CX
	JNZ	dec_loop

	// Reduce overflow: OR both uint64 lanes into one
	VPSHUFD	$0x4E, X15, X5
	VPOR	X5, X15, X15
	MOVQ	X15, AX
	MOVQ	AX, ret+16(FP)
	VZEROUPPER
	RET

// ===== AVX-512 stubs (not implemented) =====

TEXT ·encodeBlocksAVX512(SB), NOSPLIT|NOFRAME, $0-16
	VZEROUPPER
	RET

TEXT ·decodeBlocksAVX512(SB), NOSPLIT|NOFRAME, $0-24
	MOVQ	$0, ret+16(FP)
	VZEROUPPER
	RET
