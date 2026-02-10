//go:build arm64 && !purego

#include "textflag.h"

// NEON instructions not supported by Go's ARM64 assembler.
// Encoding: base_opcode | (Rm << 16) | (Rn << 5) | Rd
// Register numbers: V0=0, V1=1, ..., V31=31

// MUL Vd.4S, Vn.4S, Vm.4S — integer vector multiply
#define VMUL_S4(Vd, Vn, Vm) WORD $(0x4EA09C00 | ((Vm)<<16) | ((Vn)<<5) | (Vd))

// MLS Vd.4S, Vn.4S, Vm.4S — Vd = Vd - Vn * Vm
#define VMLS_S4(Vd, Vn, Vm) WORD $(0x6EA09400 | ((Vm)<<16) | ((Vn)<<5) | (Vd))

// UMULL Vd.2D, Vn.2S, Vm.2S — unsigned multiply long (low 2 lanes)
#define VUMULL(Vd, Vn, Vm) WORD $(0x2EA0C000 | ((Vm)<<16) | ((Vn)<<5) | (Vd))

// UMULL2 Vd.2D, Vn.4S, Vm.4S — unsigned multiply long (high 2 lanes)
#define VUMULL2(Vd, Vn, Vm) WORD $(0x6EA0C000 | ((Vm)<<16) | ((Vn)<<5) | (Vd))

// XTN Vd.2S, Vn.2D — narrow 64→32, lower half of Vd
#define VXTN_DS(Vd, Vn) WORD $(0x0EA12800 | ((Vn)<<5) | (Vd))

// XTN2 Vd.4S, Vn.2D — narrow 64→32, upper half of Vd
#define VXTN2_DS(Vd, Vn) WORD $(0x4EA12800 | ((Vn)<<5) | (Vd))

// XTN Vd.4H, Vn.4S — narrow 32→16, lower half of Vd
#define VXTN_SH(Vd, Vn) WORD $(0x0E612800 | ((Vn)<<5) | (Vd))

// XTN2 Vd.8H, Vn.4S — narrow 32→16, upper half of Vd
#define VXTN2_SH(Vd, Vn) WORD $(0x4E612800 | ((Vn)<<5) | (Vd))

// XTN Vd.8B, Vn.8H — narrow 16→8, lower half of Vd
#define VXTN_HB(Vd, Vn) WORD $(0x0E212800 | ((Vn)<<5) | (Vd))

// XTN2 Vd.16B, Vn.8H — narrow 16→8, upper half of Vd
#define VXTN2_HB(Vd, Vn) WORD $(0x4E212800 | ((Vn)<<5) | (Vd))

// Stride-5 interleave table for encode output (little-endian digit order).
// Input byte layout (in two-register VTBL source V11:V12):
//   V11 bytes 0-3: d_msb[0..3] (quotient of last division),
//   bytes 4-7: d3[0..3], bytes 8-11: d2[0..3], bytes 12-15: d1[0..3]
//   V12 bytes 16-19: d_lsb[0..3] (first remainder)
// Output: 20 bytes in stride-5 order, LSB digit first:
//   {d_lsb[0],d1[0],d2[0],d3[0],d_msb[0], d_lsb[1],d1[1],d2[1],d3[1],d_msb[1],
//    d_lsb[2],d1[2],d2[2],d3[2],d_msb[2], d_lsb[3]}
// First 16 output bytes (VTBL indices):
//   10 0C 08 04 00 11 0D 09 | 05 01 12 0E 0A 06 02 13
DATA encInterleaveLo<>+0(SB)/8, $0x090D110004080C10
DATA encInterleaveLo<>+8(SB)/8, $0x1302060A0E120105
GLOBL encInterleaveLo<>(SB), NOPTR|RODATA, $16
// Last 4 output bytes: {d1[3],d2[3],d3[3],d_msb[3]}
//   0F 0B 07 03
DATA encInterleaveHi<>+0(SB)/8, $0x0000000003070B0F
DATA encInterleaveHi<>+8(SB)/8, $0x0000000000000000
GLOBL encInterleaveHi<>(SB), NOPTR|RODATA, $16

// Stride-5 deinterleave table for decode input (little-endian digit order).
// Input: 20 bytes in stride-5 order across V0:V1
//   d_lsb at positions 0,5,10,15; d1 at 1,6,11,16;
//   d2 at 2,7,12,17; d3 at 3,8,13,18; d_msb at 4,9,14,19
// Output V4: {d_lsb[0..3], d1[0..3], d2[0..3], d3[0..3]}
//   00 05 0A 0F 01 06 0B 10 | 02 07 0C 11 03 08 0D 12
DATA decDeinterleaveLo<>+0(SB)/8, $0x100B06010F0A0500
DATA decDeinterleaveLo<>+8(SB)/8, $0x120D0803110C0702
GLOBL decDeinterleaveLo<>(SB), NOPTR|RODATA, $16
// Output V5: {d_msb[0..3]}
//   04 09 0E 13
DATA decDeinterleaveHi<>+0(SB)/8, $0x00000000130E0904
DATA decDeinterleaveHi<>+8(SB)/8, $0x0000000000000000
GLOBL decDeinterleaveHi<>(SB), NOPTR|RODATA, $16

// func encodeBlocksNEON(dst *byte, src *byte)
// Encodes 64 binary input bytes into 80 r85-encoded output bytes.
TEXT ·encodeBlocksNEON(SB), NOSPLIT|NOFRAME, $0-16
	MOVD	dst+0(FP), R0
	MOVD	src+8(FP), R1

	// Set up constants.
	// V26 = magic multiplier 0xC0C0C0C2 for div-by-85
	MOVD	$0xC0C0C0C1, R2
	VDUP	R2, V26.S4
	// V27 = 85 as .S4
	MOVD	$85, R2
	VDUP	R2, V27.S4
	// Byte constants for digit-to-char conversion
	VMOVI	$40, V20.B16   // base offset (digit + 40 = char)
	VMOVI	$20, V21.B16   // detect digit 20
	VMOVI	$56, V22.B16   // detect digit 56
	VMOVI	$65, V23.B16   // fixup: add 65 for digit 20 (60→125)
	VMOVI	$30, V24.B16   // fixup: add 30 for digit 56 (96→126)
	// Load interleave tables
	MOVD	$encInterleaveLo<>(SB), R3
	VLD1	(R3), [V28.B16]
	MOVD	$encInterleaveHi<>(SB), R3
	VLD1	(R3), [V29.B16]

	// Process 4 chunks of 16 bytes each (64 bytes total → 80 bytes).
	MOVD	$4, R5

enc_chunk:
	// Load 16 input bytes as little-endian uint32s (native byte order).
	VLD1.P	16(R1), [V0.B16]

	// Division round 1: d0 = acc % 85, q0 = acc / 85 (least significant digit)
	VUMULL(6, 0, 26)               // V6.2D = low pair * magic
	VUMULL2(7, 0, 26)              // V7.2D = high pair * magic
	VUZP2	V7.S4, V6.S4, V1.S4   // V1 = high 32 bits of each product
	VUSHR	$6, V1.S4, V1.S4      // V1 = quotient (acc / 85)
	VMOV	V0.B16, V5.B16         // V5 = acc (for remainder)
	VMLS_S4(5, 1, 27)             // V5 = acc - quot*85 = d0

	// Division round 2: d1 = q0 % 85, q1 = q0 / 85
	VUMULL(6, 1, 26)
	VUMULL2(7, 1, 26)
	VUZP2	V7.S4, V6.S4, V2.S4
	VUSHR	$6, V2.S4, V2.S4
	VMOV	V1.B16, V4.B16
	VMLS_S4(4, 2, 27)             // V4 = d1

	// Division round 3: d2 = q1 % 85, q2 = q1 / 85
	VUMULL(6, 2, 26)
	VUMULL2(7, 2, 26)
	VUZP2	V7.S4, V6.S4, V3.S4
	VUSHR	$6, V3.S4, V3.S4
	VMOV	V2.B16, V8.B16
	VMLS_S4(8, 3, 27)             // V8 = d2

	// Division round 4: d3 = q2 % 85, d4 = q2 / 85 (most significant digit)
	VUMULL(6, 3, 26)
	VUMULL2(7, 3, 26)
	VUZP2	V7.S4, V6.S4, V9.S4
	VUSHR	$6, V9.S4, V9.S4      // V9 = d4
	VMOV	V3.B16, V10.B16
	VMLS_S4(10, 9, 27)            // V10 = d3

	// Now: V5=d0, V4=d1, V8=d2, V10=d3, V9=d4 (all .S4, values 0–84)

	// Narrow 32-bit digits to bytes.
	// Pair d4+d3: narrow S4→H4, combine into H8, then H8→B8
	VXTN_SH(9, 9)                 // V9.4H = narrow d4
	VXTN2_SH(9, 10)               // V9.8H = {d4[0-3], d3[0-3]}
	VXTN_HB(9, 9)                 // V9.8B = {d4[0-3], d3[0-3]} as bytes

	// Pair d2+d1:
	VXTN_SH(8, 8)                 // V8.4H = narrow d2
	VXTN2_SH(8, 4)                // V8.8H = {d2[0-3], d1[0-3]}
	VXTN_HB(8, 8)                 // V8.8B = {d2[0-3], d1[0-3]} as bytes

	// d0:
	VXTN_SH(5, 5)                 // V5.4H = narrow d0
	VXTN_HB(5, 5)                 // V5.8B = {d0[0-3], ?, ?, ?, ?}

	// Pack d4-d1 (16 bytes) into V11, d0 (4 bytes) in V12
	VZIP1	V8.D2, V9.D2, V11.D2  // V11 = {d4d3_low64, d2d1_low64}

	// VZIP1 on .D2 interleaves 64-bit halves:
	// V11.D[0] = V9.D[0] (d4d3 bytes), V11.D[1] = V8.D[0] (d2d1 bytes)
	// So V11 = {d4[0],d4[1],d4[2],d4[3], d3[0],d3[1],d3[2],d3[3],
	//           d2[0],d2[1],d2[2],d2[3], d1[0],d1[1],d1[2],d1[3]}

	// Interleave digits into stride-5 output order via VTBL.
	// V11 (16 bytes: d4,d3,d2,d1) + V12 (4 bytes: d0) as two-register
	// table for VTBL2 addressing.
	VMOV	V5.B16, V12.B16

	// First 16 output bytes
	VTBL	V28.B16, [V11.B16, V12.B16], V13.B16
	// Last 4 output bytes
	VTBL	V29.B16, [V11.B16, V12.B16], V14.B16

	// Digit-to-char conversion on first 16 bytes.
	// Must do fixup BEFORE adding 40, since we compare digit values.
	VCMEQ	V13.B16, V21.B16, V15.B16  // mask where digit == 20
	VCMEQ	V13.B16, V22.B16, V16.B16  // mask where digit == 56
	VAND	V15.B16, V23.B16, V15.B16  // 65 where digit==20
	VAND	V16.B16, V24.B16, V16.B16  // 30 where digit==56
	VORR	V15.B16, V16.B16, V15.B16  // combined fixup
	VADD	V13.B16, V20.B16, V13.B16  // char = digit + 40
	VADD	V13.B16, V15.B16, V13.B16  // apply fixup

	// Digit-to-char conversion on last 4 bytes.
	VCMEQ	V14.B16, V21.B16, V15.B16
	VCMEQ	V14.B16, V22.B16, V16.B16
	VAND	V15.B16, V23.B16, V15.B16
	VAND	V16.B16, V24.B16, V16.B16
	VORR	V15.B16, V16.B16, V15.B16
	VADD	V14.B16, V20.B16, V14.B16
	VADD	V14.B16, V15.B16, V14.B16

	// Store 20 output bytes: 16 via VST1 + 4 via scalar.
	VST1.P	[V13.B16], 16(R0)
	VMOV	V14.S[0], R4
	MOVW	R4, (R0)
	ADD	$4, R0

	SUB	$1, R5
	CBNZ	R5, enc_chunk

	RET

// func decodeBlocksNEON(dst *byte, src *byte) uint64
// Decodes 80 valid r85-encoded bytes into 64 binary output bytes.
// Returns 0 on success, nonzero if any group overflows uint32.
TEXT ·decodeBlocksNEON(SB), NOSPLIT|NOFRAME, $0-24
	MOVD	dst+0(FP), R0
	MOVD	src+8(FP), R1

	// Set up constants.
	VMOVI	$40, V20.B16   // subtract from chars
	VMOVI	$85, V21.B16   // detect value 85 (from '}')
	VMOVI	$86, V22.B16   // detect value 86 (from '~')
	VMOVI	$65, V23.B16   // fixup: subtract 65 (85→20)
	VMOVI	$30, V24.B16   // fixup: subtract 30 (86→56)
	MOVD	$85, R2
	VDUP	R2, V25.S4     // constant 85 as .S4 for Horner multiply
	// Load deinterleave tables
	MOVD	$decDeinterleaveLo<>(SB), R3
	VLD1	(R3), [V26.B16]
	MOVD	$decDeinterleaveHi<>(SB), R3
	VLD1	(R3), [V27.B16]
	// Overflow accumulator
	VEOR	V28.B16, V28.B16, V28.B16

	// Process 4 chunks of 20 input bytes each (80 bytes → 64 bytes).
	MOVD	$4, R5

dec_chunk:
	// Load 20 input bytes: 16 via VLD1 + 4 via scalar.
	VLD1	(R1), [V0.B16]
	MOVWU	16(R1), R4
	VEOR	V1.B16, V1.B16, V1.B16  // clear V1
	VMOV	R4, V1.S[0]
	ADD	$20, R1

	// Character-to-digit: digit = char - 40, then fixup 85→20 and 86→56.
	VSUB	V20.B16, V0.B16, V0.B16
	VSUB	V20.B16, V1.B16, V1.B16

	// Fixup for V0: where digit==85, subtract 65 (→20)
	VCMEQ	V0.B16, V21.B16, V2.B16
	VAND	V2.B16, V23.B16, V2.B16
	VSUB	V2.B16, V0.B16, V0.B16
	// where digit==86, subtract 30 (→56)
	VCMEQ	V0.B16, V22.B16, V3.B16
	VAND	V3.B16, V24.B16, V3.B16
	VSUB	V3.B16, V0.B16, V0.B16

	// Fixup for V1 (last 4 bytes)
	VCMEQ	V1.B16, V21.B16, V2.B16
	VAND	V2.B16, V23.B16, V2.B16
	VSUB	V2.B16, V1.B16, V1.B16
	VCMEQ	V1.B16, V22.B16, V3.B16
	VAND	V3.B16, V24.B16, V3.B16
	VSUB	V3.B16, V1.B16, V1.B16

	// Deinterleave stride-5: extract d0[0-3], d1[0-3], d2[0-3], d3[0-3] → V4
	// and d4[0-3] → V5
	VTBL	V26.B16, [V0.B16, V1.B16], V4.B16
	VTBL	V27.B16, [V0.B16, V1.B16], V5.B16

	// Widen digit bytes to uint32 for Horner accumulation.
	// V4 layout: bytes 0-3=d_lsb, 4-7=d1, 8-11=d2, 12-15=d3
	VUXTL	V4.B8, V6.H8          // d_lsb+d1 bytes → halfwords
	VUXTL	V6.H4, V7.S4          // V7 = d_lsb as uint32
	VUXTL2	V6.H8, V8.S4          // V8 = d1 as uint32
	VUXTL2	V4.B16, V6.H8         // d2+d3 bytes → halfwords
	VUXTL	V6.H4, V9.S4          // V9 = d2 as uint32
	VUXTL2	V6.H8, V10.S4         // V10 = d3 as uint32
	VUXTL	V5.B8, V6.H8          // d_msb bytes → halfwords
	VUXTL	V6.H4, V11.S4         // V11 = d_msb as uint32

	// Horner's method: acc = d_msb*85^4 + d3*85^3 + d2*85^2 + d1*85 + d_lsb
	// Start from most significant digit, working down to least significant.
	// First 3 multiply-adds in 32-bit (max 26 bits, safe).
	VMOV	V11.B16, V12.B16       // acc = d_msb
	VMUL_S4(12, 12, 25)           // acc *= 85
	VADD	V12.S4, V10.S4, V12.S4 // acc += d3
	VMUL_S4(12, 12, 25)           // acc *= 85
	VADD	V12.S4, V9.S4, V12.S4 // acc += d2
	VMUL_S4(12, 12, 25)           // acc *= 85
	VADD	V12.S4, V8.S4, V12.S4 // acc += d1
	// Max value here: 52,200,624 (26 bits) — fits in uint32.

	// Final multiply-add in 64-bit (may overflow uint32).
	VUMULL(13, 12, 25)            // V13.2D = low 2 lanes * 85
	VUMULL2(14, 12, 25)           // V14.2D = high 2 lanes * 85
	VUSHLL	$0, V7.S2, V15.D2    // V15 = widen low 2 d_lsb to uint64
	VUSHLL2	$0, V7.S4, V16.D2    // V16 = widen high 2 d_lsb to uint64
	VADD	V13.D2, V15.D2, V13.D2 // add d_lsb (low pair)
	VADD	V14.D2, V16.D2, V14.D2 // add d_lsb (high pair)

	// Overflow detection: check if any high 32 bits are nonzero.
	VUZP2	V13.S4, V14.S4, V15.S4
	VORR	V15.B16, V28.B16, V28.B16  // accumulate overflow

	// Extract 32-bit results (little-endian, native byte order).
	VXTN_DS(13, 13)               // V13.2S = low 32 bits of acc[0..1]
	VXTN2_DS(13, 14)              // V13.4S = {acc0, acc1, acc2, acc3}

	// Store 16 output bytes.
	VST1.P	[V13.B16], 16(R0)

	SUB	$1, R5
	CBNZ	R5, dec_chunk

	// Return overflow status: 0 = success, nonzero = overflow.
	VMOV	V28.D[0], R2
	VMOV	V28.D[1], R3
	ORR	R2, R3, R2
	MOVD	R2, ret+16(FP)
	RET
