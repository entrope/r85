package r85

import "io"

// MaxEncodedLen returns the maximum length of an encoding of n source bytes.
func MaxEncodedLen(n int) int {
	res := 5 * (n / 4)
	if r := n % 4; r != 0 {
		res += r + 1
	}
	return res
}

// MaxDecodedLen returns the maximum length of a decoding of n source bytes.
// The actual decoded length may be smaller if the input contains characters
// outside the r85 alphabet (which are skipped during decoding).
func MaxDecodedLen(n int) int {
	res := 4 * (n / 5)
	if r := n % 5; r >= 2 {
		res += r - 1
	}
	return res
}

// encByte maps an r85 digit value (0–84) to its encoded byte.
// The base alphabet starts at '(' (40).  Two characters are replaced:
// '<' (40+20=60) -> '}' (125), and '`' (40+56=96) -> '~' (126).
//
// This uses arithmetic rather than a branch or lookup table:
//
//	eq20 is 1 when v == 20, else 0
//	eq56 is 1 when v == 56, else 0
//	delta is +65 when v == 20, +30 when v == 56, else 0
func encByte(v byte) byte {
	eq20 := subtle_eq(v, 20)
	eq56 := subtle_eq(v, 56)
	return v + 40 + 65*eq20 + 30*eq56
}

// decByte maps an encoded byte back to an r85 digit value (0–84).
// Returns the digit value and 1 if valid, or (0, 0) if invalid.
func decByte(b byte) (byte, byte) {
	// '}' (125) -> value 20, '~' (126) -> value 56
	eq125 := subtle_eq(b, 125)
	eq126 := subtle_eq(b, 126)
	// Map '}' and '~' into the base range for uniform handling:
	//   125 - 65 = 60 (which is 40+20), 126 - 30 = 96 (which is 40+56)
	adjusted := b - 65*eq125 - 30*eq126
	v := adjusted - 40
	// Valid if b >= 40 and v < 85.
	inRange := subtle_geq(b, 40) & subtle_lt(v, 85)
	// Exclude '<' (60) and '`' (96) as raw input bytes.
	notExcluded := 1 - subtle_eq(b, 60)&(1-eq125) - subtle_eq(b, 96)&(1-eq126)
	ok := inRange & notExcluded
	return v & (0xFF * ok), ok
}

// subtle_eq returns 1 if a == b, else 0, without branching.
func subtle_eq(a, b byte) byte {
	x := uint16(a ^ b)
	return 1 - byte((x|-x)>>8&1)
}

// subtle_lt returns 1 if a < b, else 0, without branching.
func subtle_lt(a, b byte) byte {
	return byte((uint16(a) - uint16(b)) >> 15)
}

// subtle_geq returns 1 if a >= b, else 0, without branching.
func subtle_geq(a, b byte) byte {
	return 1 - subtle_lt(a, b)
}

// Encode encodes binary src into text dst, returning the number of
// bytes written to dst.
// If dst is too short, Encode fills dst and returns len(dst),
// and ignores the remaining input.
func Encode(dst, src []byte) int {
	di := 0
	si := 0

	// Process full 4-byte blocks.
	for si+4 <= len(src) {
		if di+5 > len(dst) {
			return len(dst)
		}
		acc := uint32(src[si])<<24 | uint32(src[si+1])<<16 | uint32(src[si+2])<<8 | uint32(src[si+3])
		dst[di+4] = encByte(byte(acc % 85))
		acc /= 85
		dst[di+3] = encByte(byte(acc % 85))
		acc /= 85
		dst[di+2] = encByte(byte(acc % 85))
		acc /= 85
		dst[di+1] = encByte(byte(acc % 85))
		acc /= 85
		dst[di+0] = encByte(byte(acc))
		di += 5
		si += 4
	}

	// Handle trailing 1–3 bytes.
	switch len(src) - si {
	case 1:
		if di+2 > len(dst) {
			return len(dst)
		}
		acc := uint32(src[si])
		dst[di+1] = encByte(byte(acc % 85))
		dst[di+0] = encByte(byte(acc / 85))
		di += 2
	case 2:
		if di+3 > len(dst) {
			return len(dst)
		}
		acc := uint32(src[si])<<8 | uint32(src[si+1])
		dst[di+2] = encByte(byte(acc % 85))
		acc /= 85
		dst[di+1] = encByte(byte(acc % 85))
		dst[di+0] = encByte(byte(acc / 85))
		di += 3
	case 3:
		if di+4 > len(dst) {
			return len(dst)
		}
		acc := uint32(src[si])<<16 | uint32(src[si+1])<<8 | uint32(src[si+2])
		dst[di+3] = encByte(byte(acc % 85))
		acc /= 85
		dst[di+2] = encByte(byte(acc % 85))
		acc /= 85
		dst[di+1] = encByte(byte(acc % 85))
		dst[di+0] = encByte(byte(acc / 85))
		di += 4
	}

	return di
}

// EncodeToString returns the r85 encoding of src as a string.
func EncodeToString(src []byte) string {
	dst := make([]byte, MaxEncodedLen(len(src)))
	n := Encode(dst, src)
	return string(dst[:n])
}

// DecodeString returns the bytes represented by the r85 string s.
func DecodeString(s string) ([]byte, error) {
	src := []byte(s)
	dst := make([]byte, MaxDecodedLen(len(src)))
	ndst, _, err := Decode(dst, src)
	return dst[:ndst], err
}

// Decode decodes text src into binary dst.
// ndst contains the number of bytes written into dst.
// nsrc contains the number of bytes consumed from src, including
// skipped bytes.
// If dst is too short, Decode fills dst and returns len(dst),
// and ignores the remaining input.
func Decode(dst, src []byte) (ndst, nsrc int, err error) {
	di := 0
	si := 0

	// Collect valid characters into a block buffer.
	var block [5]byte
	bi := 0

	for si < len(src) {
		v, ok := decByte(src[si])
		si++
		if ok == 0 {
			continue
		}
		block[bi] = v
		bi++

		if bi < 5 {
			continue
		}

		// Full 5-character block -> 4 bytes.
		if di+4 > len(dst) {
			return len(dst), si, nil
		}
		// Use uint64 because 85^5 = 4,437,053,125 > 2^32.
		acc := uint64(block[0])
		for k := 1; k < 5; k++ {
			acc = acc*85 + uint64(block[k])
		}
		if acc > 0xFFFFFFFF {
			return di, si, CorruptInputError{"value overflow in 5-character block"}
		}
		dst[di+0] = byte(acc >> 24)
		dst[di+1] = byte(acc >> 16)
		dst[di+2] = byte(acc >> 8)
		dst[di+3] = byte(acc)
		di += 4
		bi = 0
	}

	// Handle trailing block.
	if bi == 0 {
		return di, si, nil
	}
	if bi == 1 {
		return di, si, CorruptInputError{"incomplete block: single trailing character"}
	}

	// bi is 2, 3, or 4 -> 1, 2, or 3 output bytes.
	outLen := bi - 1
	if di+outLen > len(dst) {
		return len(dst), si, nil
	}

	var acc uint32
	for k := 0; k < bi; k++ {
		acc = acc*85 + uint32(block[k])
	}

	maxVal := [3]uint32{0xFF, 0xFFFF, 0xFFFFFF}
	if acc > maxVal[outLen-1] {
		return di, si, CorruptInputError{"value overflow in trailing block"}
	}

	for i := outLen - 1; i >= 0; i-- {
		dst[di+i] = byte(acc)
		acc >>= 8
	}
	di += outLen

	return di, si, nil
}

// NewEncoder wraps a buffer and io.WriteCloser interface around Encode.
// This will only write a short block (less than 4 bytes of binary input)
// when Close is called.
func NewEncoder(w io.Writer) io.WriteCloser {
	return &encoder{w: w}
}

type encoder struct {
	w   io.Writer
	buf [4]byte
	n   int
	err error
}

func (e *encoder) Write(p []byte) (int, error) {
	if e.err != nil {
		return 0, e.err
	}
	written := 0

	// If we have pending bytes, fill up to a 4-byte block.
	if e.n > 0 {
		for len(p) > 0 && e.n < 4 {
			e.buf[e.n] = p[0]
			e.n++
			p = p[1:]
			written++
		}
		if e.n < 4 {
			return written, nil
		}
		var out [5]byte
		Encode(out[:], e.buf[:])
		_, e.err = e.w.Write(out[:])
		e.n = 0
		if e.err != nil {
			return written, e.err
		}
	}

	// Encode full 4-byte blocks directly from p.
	for len(p) >= 4 {
		var out [5]byte
		Encode(out[:], p[:4])
		_, e.err = e.w.Write(out[:])
		if e.err != nil {
			return written, e.err
		}
		p = p[4:]
		written += 4
	}

	// Buffer remaining bytes.
	for len(p) > 0 {
		e.buf[e.n] = p[0]
		e.n++
		p = p[1:]
		written++
	}

	return written, nil
}

func (e *encoder) Close() error {
	if e.err != nil {
		return e.err
	}
	if e.n > 0 {
		var out [5]byte
		n := Encode(out[:], e.buf[:e.n])
		_, e.err = e.w.Write(out[:n])
		e.n = 0
	}
	return e.err
}

// NewDecoder wraps a buffer and io.Reader interface around Decode.
//
// The decoder handles the case where the underlying reader delivers data
// at arbitrary byte boundaries that may split an encoded block.  It
// carries incomplete blocks (1–4 valid r85 digits) across Read calls
// and only treats them as a final partial block when the underlying
// reader returns io.EOF (or another error).  A single trailing r85
// digit at true EOF is reported as a CorruptInputError.
func NewDecoder(r io.Reader) io.Reader {
	return &decoder{r: r}
}

type decoder struct {
	r      io.Reader
	carry  [4]byte // up to 4 undecoded r85 digits carried across reads
	cn     int     // number of carried digits
	outbuf [816]byte
	out    []byte
	err    error
}

func (d *decoder) Read(p []byte) (int, error) {
	if len(d.out) > 0 {
		n := copy(p, d.out)
		d.out = d.out[n:]
		return n, nil
	}

	if d.err != nil {
		return 0, d.err
	}

	// Read encoded input into a temporary buffer, prepending any carry.
	var inbuf [1024]byte
	copy(inbuf[:], d.carry[:d.cn])
	nn, readErr := d.r.Read(inbuf[d.cn:])
	total := d.cn + nn
	d.cn = 0

	if total == 0 {
		if readErr != nil {
			d.err = readErr
		}
		return 0, d.err
	}

	if readErr == nil {
		// Not at EOF: keep a partial trailing block for next read.
		// Count valid r85 characters in the tail to find how many
		// to carry over.  We need to carry the last (validCount % 5)
		// valid characters plus any trailing invalid characters after them.
		// Simpler: scan backwards to find up to 4 valid trailing chars
		// that don't form a complete block.

		// Count total valid chars.
		validCount := 0
		for i := 0; i < total; i++ {
			_, ok := decByte(inbuf[i])
			if ok != 0 {
				validCount++
			}
		}
		trailing := validCount % 5
		if trailing > 0 {
			// Walk backwards to find the start of the last `trailing` valid chars.
			found := 0
			carryStart := total
			for i := total - 1; i >= 0 && found < trailing; i-- {
				_, ok := decByte(inbuf[i])
				if ok != 0 {
					found++
					carryStart = i
				}
			}
			d.cn = copy(d.carry[:], inbuf[carryStart:total])
			total = carryStart
		}
	}

	if total > 0 {
		ndst, _, decErr := Decode(d.outbuf[:], inbuf[:total])
		if decErr != nil {
			d.err = decErr
			if ndst == 0 {
				return 0, d.err
			}
		}
		d.out = d.outbuf[:ndst]
		n := copy(p, d.out)
		d.out = d.out[n:]
		if readErr != nil && len(d.out) == 0 && d.err == nil {
			d.err = readErr
		}
		return n, nil
	}

	// total == 0 but readErr == nil shouldn't happen, but handle gracefully.
	if readErr != nil {
		d.err = readErr
	}
	return 0, d.err
}

// CorruptInputError is returned by [Decode] and [DecodeString] when the
// input is not valid r85 text.
type CorruptInputError struct {
	// Reason describes why decoding failed.
	Reason string
}

func (e CorruptInputError) Error() string {
	return "r85: " + e.Reason
}
