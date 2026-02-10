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

// encTable maps an r85 digit value (0–84) to its encoded byte.
// The base alphabet starts at '(' (40).  Two characters are replaced:
// '<' (40+20=60) -> '}' (125), and '`' (40+56=96) -> '~' (126).
var encTable = [85]byte{
	'(', ')', '*', '+', ',', '-', '.', '/', '0', '1', // 0–9
	'2', '3', '4', '5', '6', '7', '8', '9', ':', ';', // 10–19
	'}', '=', '>', '?', '@', 'A', 'B', 'C', 'D', 'E', // 20–29
	'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', // 30–39
	'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', // 40–49
	'Z', '[', '\\', ']', '^', '_', '~', 'a', 'b', 'c', // 50–59
	'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', // 60–69
	'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', // 70–79
	'x', 'y', 'z', '{', '|', // 80–84
}

// decTable maps an encoded byte to its r85 digit value (0–84),
// or 0xFF if the byte is not in the r85 alphabet.
// Both the canonical encoded forms ('}' for 20, '~' for 56) and their
// unescaped equivalents ('<' for 20, '`' for 56) are accepted.
var decTable = [256]byte{
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // 0–15
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // 16–31
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, // 32–47: '(' is 40=0x00
	0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, // 48–63: '<' (60)=0x14=20
	0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, // 64–79
	0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E, 0x2F, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, // 80–95
	0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D, 0x3E, 0x3F, 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, // 96–111: '`' (96)=0x38=56
	0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F, 0x50, 0x51, 0x52, 0x53, 0x54, 0x14, 0x38, 0xFF, // 112–127: '}' (125)=0x14=20, '~' (126)=0x38=56
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // 128–143
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // 144–159
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // 160–175
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // 176–191
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // 192–207
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // 208–223
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // 224–239
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // 240–255
}

func encByte(v byte) byte {
	return encTable[v]
}

// decByte maps an encoded byte back to an r85 digit value (0–84).
// Returns the digit value and 1 if valid, or (0, 0) if invalid.
func decByte(b byte) (byte, byte) {
	v := decTable[b]
	if v == 0xFF {
		return 0, 0
	}
	return v, 1
}

// Encode encodes binary src into text dst, returning the number of
// bytes written to dst.
// If dst is too short, Encode fills dst and returns len(dst),
// and ignores the remaining input.
func Encode(dst, src []byte) int {
	di := 0
	si := 0

	// SIMD fast path: process 64-byte blocks.
	if haveSIMD {
		for si+64 <= len(src) && di+80 <= len(dst) {
			encodeBlocksSIMD(&dst[di], &src[si])
			di += 80
			si += 64
		}
	}

	// Process full 4-byte blocks.
	for si+4 <= len(src) {
		if di+5 > len(dst) {
			return di
		}
		acc := uint32(src[si]) | uint32(src[si+1])<<8 | uint32(src[si+2])<<16 | uint32(src[si+3])<<24
		dst[di+0] = encByte(byte(acc % 85))
		acc /= 85
		dst[di+1] = encByte(byte(acc % 85))
		acc /= 85
		dst[di+2] = encByte(byte(acc % 85))
		acc /= 85
		dst[di+3] = encByte(byte(acc % 85))
		acc /= 85
		dst[di+4] = encByte(byte(acc))
		di += 5
		si += 4
	}

	// Handle trailing 1–3 bytes.
	res := len(src) - si
	if res != 0 && di+res+1 > len(dst) {
		return di
	}
	switch res {
	case 1:
		acc := uint32(src[si])
		dst[di+0] = encByte(byte(acc % 85))
		dst[di+1] = encByte(byte(acc / 85))
		di += 2
	case 2:
		acc := uint32(src[si]) | uint32(src[si+1])<<8
		dst[di+0] = encByte(byte(acc % 85))
		acc /= 85
		dst[di+1] = encByte(byte(acc % 85))
		dst[di+2] = encByte(byte(acc / 85))
		di += 3
	case 3:
		acc := uint32(src[si]) | uint32(src[si+1])<<8 | uint32(src[si+2])<<16
		dst[di+0] = encByte(byte(acc % 85))
		acc /= 85
		dst[di+1] = encByte(byte(acc % 85))
		acc /= 85
		dst[di+2] = encByte(byte(acc % 85))
		dst[di+3] = encByte(byte(acc / 85))
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
// allValidR85 reports whether all bytes in b are valid r85 characters
// (in the range [40, 126]).
func allValidR85(b []byte) bool {
	for _, c := range b {
		if c < 40 || c > 126 {
			return false
		}
	}
	return true
}

func Decode(dst, src []byte) (ndst, nsrc int, err error) {
	di := 0
	si := 0

	// SIMD fast path: process runs of 80 valid r85 bytes.
	if haveSIMD {
		for di+64 <= len(dst) && si+80 <= len(src) {
			if !allValidR85(src[si : si+80]) {
				break
			}
			ovf := decodeBlocksSIMD(&dst[di], &src[si])
			if ovf != 0 {
				// Overflow detected; fall through to scalar for error reporting.
				break
			}
			di += 64
			si += 80
		}
	}

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
			return di, si, nil
		}
		// Use uint64 because 85^5 = 4,437,053,125 > 2^32.
		acc := uint64(block[4])*85 + uint64(block[3])
		acc = acc*85 + uint64(block[2])
		acc = acc*85 + uint64(block[1])
		acc = acc*85 + uint64(block[0])
		if acc > 0xFFFFFFFF {
			return di, si, CorruptInputError{"value overflow in 5-character block"}
		}
		dst[di+0] = byte(acc)
		dst[di+1] = byte(acc >> 8)
		dst[di+2] = byte(acc >> 16)
		dst[di+3] = byte(acc >> 24)
		di += 4
		bi = 0
	}

	// Handle trailing block.
	switch bi {
	case 0:
		return di, si, nil
	case 1:
		return di, si, CorruptInputError{"incomplete block: single trailing character"}
	case 2: // 2 chars -> 1 byte
		if di+1 > len(dst) {
			return di, si, nil
		}
		acc := uint32(block[1])*85 + uint32(block[0])
		if acc > 0xFF {
			return di, si, CorruptInputError{"value overflow in trailing block"}
		}
		dst[di] = byte(acc)
		di++
	case 3: // 3 chars -> 2 bytes
		if di+2 > len(dst) {
			return di, si, nil
		}
		acc := (uint32(block[2])*85+uint32(block[1]))*85 + uint32(block[0])
		if acc > 0xFFFF {
			return di, si, CorruptInputError{"value overflow in trailing block"}
		}
		dst[di+0] = byte(acc)
		dst[di+1] = byte(acc >> 8)
		di += 2
	case 4: // 4 chars -> 3 bytes
		if di+3 > len(dst) {
			return di, si, nil
		}
		acc := ((uint32(block[3])*85+uint32(block[2]))*85+uint32(block[1]))*85 + uint32(block[0])
		if acc > 0xFFFFFF {
			return di, si, CorruptInputError{"value overflow in trailing block"}
		}
		dst[di+0] = byte(acc)
		dst[di+1] = byte(acc >> 8)
		dst[di+2] = byte(acc >> 16)
		di += 3
	}

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
	out [4096]byte
	on  int
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
		e.on += Encode(e.out[e.on:], e.buf[:])
		e.n = 0
	}

	// Encode directly from p into the output buffer, flushing as needed.
	for len(p) >= 4 {
		if e.on+5 > len(e.out) {
			if e.err = e.flush(); e.err != nil {
				return written, e.err
			}
		}
		// Encode as many full blocks as fit in the remaining output buffer.
		// Each 4 input bytes produce 5 output bytes.
		outAvail := len(e.out) - e.on
		maxIn := (outAvail / 5) * 4
		if maxIn > len(p) {
			// Round down to a 4-byte boundary so we only encode full blocks.
			maxIn = len(p) &^ 3
		}
		e.on += Encode(e.out[e.on:], p[:maxIn])
		p = p[maxIn:]
		written += maxIn
	}

	// Buffer remaining 0–3 bytes.
	for len(p) > 0 {
		e.buf[e.n] = p[0]
		e.n++
		p = p[1:]
		written++
	}

	return written, nil
}

func (e *encoder) flush() error {
	if e.on > 0 {
		_, e.err = e.w.Write(e.out[:e.on])
		e.on = 0
	}
	return e.err
}

func (e *encoder) Close() error {
	if e.err != nil {
		return e.err
	}
	if e.n > 0 {
		e.on += Encode(e.out[e.on:], e.buf[:e.n])
		e.n = 0
	}
	return e.flush()
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
