# CLAUDE.md

## Project Overview

r85 is a Go package implementing a base-85 binary-to-text encoding scheme, similar to Ascii85 and ZeroMQ's Z85. It uses a custom alphabet starting at ASCII `(` (decimal 40) with substitutions (`<` -> `}`, `` ` `` -> `~`) to minimize escaping in most languages and HTML.

Module path: `github.com/entrope/r85`

## Build & Test Commands

- Run all tests: `go test ./...`
- Run tests with verbose output: `go test -v ./...`
- Run benchmarks: `go test -bench=. -benchmem ./...`
- Run a specific benchmark: `go test -bench=BenchmarkEncode -benchmem ./...`
- Verify generic fallback: `go test -tags purego ./...`
- Vet for another arch: `GOARCH=amd64 go vet ./...`
- Build gen tool: `go build ./gen`
- Run gen tests: `go test ./gen`
- Run gen smoke tests: `bash gen/gen_test.sh`
- Generate a random ID: `go run ./gen`
- Generate 10 sequential UUID v7s: `go run ./gen -uuid v7 -n 10 -seq`

## Code Structure

Single-package library:

### Core files
- `r85.go` — Encoding/decoding logic: `Encode`, `Decode`, `EncodeToString`, `DecodeString`, streaming `NewEncoder`/`NewDecoder`, lookup tables, `CorruptInputError`, `allValidR85` helper, SIMD dispatch entry points

### SIMD dispatch (two-tier, following crypto/sha256 pattern)
- `r85_arm64.go` — `const haveSIMD = true`; NEON asm stubs; `encodeBlocksSIMD`/`decodeBlocksSIMD` wrappers
- `r85_arm64.s` — NEON assembly: `encodeBlocksNEON`, `decodeBlocksNEON` (implemented)
- `r85_amd64.go` — `const haveSIMD = true`; `var haveAVX512` (runtime via `golang.org/x/sys/cpu`); AVX2/AVX-512 asm stubs; `encodeBlocksSIMD`/`decodeBlocksSIMD` dispatch
- `r85_amd64.s` — AVX2 assembly: `encodeBlocksAVX2` (YMM), `decodeBlocksAVX2` (XMM); AVX-512 stubs (TODO)
- `r85_noasm.go` — `const haveSIMD = false`; no-op stubs for other architectures or `purego` builds

### Tests
- `r85_test.go` — Unit tests (21 tests)
- `r85_bench_test.go` — Benchmarks across various payload sizes and chunking strategies

### CLI utility
- `gen/main.go` — Generates random IDs in r85 encoding; supports 96-bit (default), UUID v4, UUID v7. Flags: `-n` (count), `-uuid` (v4/v7), `-seq` (sequential from random base)
- `gen/main_test.go` — Unit tests for ID generation functions
- `gen/gen_test.sh` — Shell-based smoke tests

## Key Design Details

- Encoding: 4 bytes -> 5 chars (big-endian); partial trailing blocks (1-3 bytes) produce 2-4 chars
- Decoding: skips characters outside the r85 alphabet; accepts both canonical (`}`, `~`) and unescaped (`<`, `` ` ``) forms
- Streaming encoder buffers internally (4096-byte output buffer) and only writes partial blocks on `Close()`
- Streaming decoder carries incomplete blocks across `Read` calls
- A single trailing r85 digit is a `CorruptInputError`; value overflows are also reported as `CorruptInputError`

## SIMD Vectorization

### Dispatch Architecture

`r85.go` uses architecture-neutral names (`haveSIMD`, `encodeBlocksSIMD`, `decodeBlocksSIMD`). Per-architecture files provide the implementations:

- **arm64**: `encodeBlocksSIMD` → `encodeBlocksNEON` (trivial inlined wrapper)
- **amd64**: `encodeBlocksSIMD` → `encodeBlocksAVX512` if `haveAVX512`, else `encodeBlocksAVX2`
- **other/purego**: `haveSIMD = false`, fast path skipped entirely

All ISAs use the same block size: **64 binary bytes ↔ 80 text bytes** per call. The dispatch loop in `r85.go` is identical regardless of architecture.

Adding a new architecture requires only a `r85_<arch>.go` + `.s` pair — `r85.go` never changes.

### Division by 85 — Magic Number

NEON (and x86) have no integer vector divide. Use multiply-high:

```
floor(n / 85) = mulhi(0xC0C0C0C1, n) >> 6
```

- M = `0xC0C0C0C1` = floor(2^38 / 85) = 3,233,857,729
- **Do NOT use** `0xC0C0C0C2` (ceil) — it fails for n=0xFFFFFFFE
- Remainder: `n - floor(n/85) * 85`
- Source: Hacker's Delight Ch.10; Granlund & Montgomery 1994

### Digit ↔ Character Conversion (no lookup table)

Encode: `char = digit + 40`, then fixup digit 20 (add 65 → `}`) and digit 56 (add 30 → `~`) via vector compare + masked add.

Decode: `digit = char - 40`, then fixup value 85 (subtract 65 → 20) and value 86 (subtract 30 → 56) via vector compare + masked subtract.

### ARM64 NEON Implementation (complete)

Each call processes 4 chunks of 16/20 bytes (4 uint32 lanes per 128-bit register):
- **Encode**: VREV32 byte-swap → 4 rounds of UMULL/VUZP2/VUSHR/MLS for div-85 → XTN narrowing chain → VTBL stride-5 interleave → VCMEQ+VADD char fixup
- **Decode**: VSUB+VCMEQ char-to-digit → VTBL deinterleave → VUXTL widening → 3× VMUL+VADD Horner steps (32-bit) → UMULL final step (64-bit, overflow check) → XTN+VREV32 byte-swap

### x86-64 AVX2 Implementation (complete, tested on Zen 2)

**Encode** uses YMM (256-bit): 2 iterations × 8 uint32 lanes.
**Decode** uses XMM (128-bit): 4 iterations × 4 uint32 lanes.

YMM is optimal for encode because the 32-byte input is a power-of-2 aligned load and all division/packing/interleave operations benefit from wider registers. YMM is counterproductive for decode because the 20-byte input chunks require per-chunk XMM processing anyway; combining via VINSERTI128 adds synchronization overhead without enabling better pipelining.

Key x86 adaptations from the NEON algorithm:
- **Division by 85**: `VPMULUDQ` (32×32→64 unsigned multiply) only operates on even uint32 lanes. Odd lanes require `VPSHUFD $0xF5` to rotate odd→even, a second `VPMULUDQ`, then `VPSLLQ $32 + VPOR` to merge results back.
- **Digit packing (encode)**: `VPACKUSDW` + `VPACKUSWB` narrow uint32 digits to bytes. These are lane-local on AVX2 (each 128-bit half packs independently), which is fine since the subsequent `VPSHUFB` interleave is also lane-local.
- **Stride-5 interleave/deinterleave**: `VPSHUFB` with RODATA index tables. Same masks work in both YMM halves (lane-local). Encode stores 20 bytes per 128-bit half via `MOVOU` (16 bytes) + `MOVL` (4 bytes) after `VEXTRACTI128`.
- **Char ↔ digit**: Same compare+masked-add/subtract approach as NEON, using `VPCMPEQB` + `VPAND` + `VPADDB`/`VPSUBB`.

Benchmarks on AMD Ryzen Threadripper 3960X (Zen 2, 3.8 GHz):

| Operation | Scalar (purego) | AVX2 SIMD | Speedup |
|-----------|----------------|-----------|---------|
| Encode    | ~900 MB/s      | ~3,130 MB/s | 3.5×  |
| Decode    | ~520 MB/s      | ~1,035 MB/s | 2.0×  |

### x86-64 AVX-512 (stub only — TODO)

Assembly stubs exist but contain only `VZEROUPPER; RET`.
- 1 iteration × 16 lanes (512-bit ZMM). `VPERMB` (requires AVX512-VBMI, Icelake+) enables full cross-register byte shuffle.
- Go's x86 assembler natively supports all AVX-512 mnemonics — no raw BYTE/LONG encoding needed.
- Every function must execute `VZEROUPPER` before returning.

## Lessons Learned (ARM64 Assembly)

### Go Plan 9 ASM Operand Order

Go's ARM64 assembler **reverses Vn/Vm** compared to ARM documentation for many instructions. For example:

```
Go asm:   VUZP2 Va.S4, Vb.S4, Vd.S4
ARM asm:  UZP2  Vd.4S, Vb.4S, Va.4S
```

Getting this wrong causes silent lane swapping — all-zero inputs work fine (lanes identical), but non-uniform data produces wrong results.

### WORD-Encoded Instructions

Go's ARM64 assembler lacks native support for integer vector multiply (`VMUL .S4`), `MLS`, `UMULL`/`UMULL2`, and `XTN`/`XTN2`. These must be emitted as `WORD $0x...` directives. The encoding formula for 3-register NEON ops is: `base_opcode | (Rm << 16) | (Rn << 5) | Rd`.

Verify with `go tool objdump -s <funcname>` after building.

### RODATA Byte Order

`DATA` directives for RODATA tables use little-endian encoding. When specifying 8-byte values for byte-index tables, each byte must be placed at its correct position within the little-endian uint64. Miscalculating this is a common source of silent data corruption.

### Build Tag for Assembly Files

Assembly `.s` files on ARM64/AMD64 need an explicit `//go:build` constraint to support the `purego` tag. Without it, the assembler includes the file unconditionally for the matching GOARCH, causing duplicate symbol errors when `r85_noasm.go` is also compiled.

## Lessons Learned (x86-64 Assembly)

### VEX vs Non-VEX SSE: False Register Dependencies

**Critical**: Never use non-VEX SSE load instructions (`MOVOU`/`MOVDQU`, `MOVL` GP→XMM) in loops that also use YMM registers. Non-VEX SSE instructions preserve the upper 128 bits of YMM registers, creating false dependencies on whatever last wrote those upper bits. If that was a YMM operation from a previous loop iteration (especially a late-stage operation like the Horner chain output), this serializes the entire loop — the next iteration's loads cannot begin until the previous iteration's computation completes.

This caused a **21× slowdown** (1,036 MB/s → 49 MB/s) in the decode path on AMD Zen 2.

**Fix**: Use VEX-encoded loads exclusively:
- `MOVOU mem, X0` → `VMOVDQU mem, X0` (VEX zeros upper YMM bits)
- `MOVL AX, X4` → `VPXOR X4, X4, X4; MOVL AX, X4` (VPXOR breaks the dependency chain; the CPU recognizes the zeroing idiom and resolves it with zero latency)

Non-VEX SSE **stores** (`MOVOU X5, (DI)`, `MOVL X14, 16(DI)`) are safe — they read the XMM register (no merge needed) and write to memory.

### VPMULUDQ Even-Lane Limitation

`VPMULUDQ` multiplies only the even uint32 lanes (lanes 0, 2, 4, 6) producing uint64 results. To process all lanes:
1. `VPMULUDQ M, acc, result_even` — multiply even lanes
2. `VPSHUFD $0xF5, acc, tmp` — rotate odd lanes to even positions
3. `VPMULUDQ M, tmp, result_odd` — multiply (former) odd lanes
4. Merge: `VPSLLQ $32, result_odd, result_odd; VPOR result_even, result_odd, result`

### YMM vs XMM: Choose Width by Data Alignment

YMM (256-bit) is beneficial when the input/output naturally aligns to 32 bytes (encode: 32 bytes in → 40 bytes out per YMM half). YMM is counterproductive when the natural unit is not a power of 2 (decode: 20-byte input chunks), because combining two XMM results via `VINSERTI128` adds a synchronization point that prevents pipelining between chunks.

### Go objdump Limitations

`go tool objdump` does not correctly disassemble many VEX-encoded instructions. It misparses VEX prefixes as legacy instructions (showing `CLC`, `FISTTP`, `HLT`, etc.), causing all subsequent instruction boundaries to cascade into garbage. Use `objdump -d` from binutils for accurate disassembly of AVX2/AVX-512 code. Despite the garbled disassembly, the assembler encodes instructions correctly — verify by checking test results rather than objdump output.
