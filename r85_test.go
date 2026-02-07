package r85

import (
	"bytes"
	"io"
	"testing"
)

// TestEncByteAlphabet verifies the full encoding alphabet.
func TestEncByteAlphabet(t *testing.T) {
	// Build expected alphabet: ASCII 40..124, with 60->'}'.and 96->'~'
	var expected [85]byte
	for i := range 85 {
		b := byte(i + 40)
		if b == '<' {
			b = '}'
		} else if b == '`' {
			b = '~'
		}
		expected[i] = b
	}
	for i := range 85 {
		got := encByte(byte(i))
		if got != expected[i] {
			t.Errorf("encByte(%d) = %d (%c), want %d (%c)", i, got, got, expected[i], expected[i])
		}
	}
}

// TestDecByteRoundtrip verifies decByte inverts encByte for all 85 values.
func TestDecByteRoundtrip(t *testing.T) {
	for i := range 85 {
		enc := encByte(byte(i))
		v, ok := decByte(enc)
		if ok != 1 {
			t.Errorf("decByte(encByte(%d)=%d) returned ok=0", i, enc)
			continue
		}
		if v != byte(i) {
			t.Errorf("decByte(encByte(%d)) = %d, want %d", i, v, i)
		}
	}
}

// TestDecByteInvalid verifies that bytes outside the alphabet are rejected.
func TestDecByteInvalid(t *testing.T) {
	valid := make(map[byte]bool)
	for i := range 85 {
		valid[encByte(byte(i))] = true
	}
	for b := range 256 {
		_, ok := decByte(byte(b))
		if valid[byte(b)] {
			if ok != 1 {
				t.Errorf("decByte(%d=%c) should be valid", b, b)
			}
		} else {
			if ok != 0 {
				t.Errorf("decByte(%d=%c) should be invalid", b, b)
			}
		}
	}
}

// TestEncodeDecodeRoundtrip tests encode/decode for various lengths.
func TestEncodeDecodeRoundtrip(t *testing.T) {
	for n := range 20 {
		src := make([]byte, n)
		for i := range n {
			src[i] = byte(i*37 + 13)
		}
		encLen := MaxEncodedLen(n)
		enc := make([]byte, encLen)
		nw := Encode(enc, src)
		if nw != encLen {
			t.Errorf("Encode(%d bytes): wrote %d, want %d", n, nw, encLen)
		}

		dec := make([]byte, n)
		ndst, nsrc, err := Decode(dec, enc)
		if err != nil {
			t.Errorf("Decode(%d bytes): err = %v", n, err)
		}
		if ndst != n {
			t.Errorf("Decode(%d bytes): ndst = %d", n, ndst)
		}
		if nsrc != encLen {
			t.Errorf("Decode(%d bytes): nsrc = %d, want %d", n, nsrc, encLen)
		}
		if !bytes.Equal(dec, src) {
			t.Errorf("Decode(%d bytes): got %v, want %v", n, dec, src)
		}
	}
}

// TestEncodeKnownValues tests specific known encodings.
func TestEncodeKnownValues(t *testing.T) {
	tests := []struct {
		input []byte
		want  string
	}{
		{[]byte{0}, "(("}, // 0 -> value 0 -> '(' '('
		{[]byte{0, 0, 0, 0}, "((((("}, // 0x00000000 -> 5 * '('
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF}, "z?^4("}, // 2^32-1 = 4294967295
	}
	for _, tt := range tests {
		enc := make([]byte, MaxEncodedLen(len(tt.input)))
		n := Encode(enc, tt.input)
		got := string(enc[:n])
		if got != tt.want {
			t.Errorf("Encode(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestDecodeSkipsInvalidChars verifies that non-alphabet characters are skipped.
func TestDecodeSkipsInvalidChars(t *testing.T) {
	// Encode [1, 2, 3, 4] then insert spaces/newlines.
	src := []byte{1, 2, 3, 4}
	enc := make([]byte, 5)
	Encode(enc, src)
	// Insert spaces around each character.
	spaced := make([]byte, 0, 15)
	for _, b := range enc {
		spaced = append(spaced, ' ', b, '\n')
	}
	dec := make([]byte, 4)
	ndst, _, err := Decode(dec, spaced)
	if err != nil {
		t.Fatalf("Decode with spaces: err = %v", err)
	}
	if ndst != 4 {
		t.Fatalf("Decode with spaces: ndst = %d", ndst)
	}
	if !bytes.Equal(dec, src) {
		t.Errorf("Decode with spaces: got %v, want %v", dec, src)
	}
}

// TestDecodeSingleCharError verifies that a single trailing character is an error.
func TestDecodeSingleCharError(t *testing.T) {
	_, _, err := Decode(make([]byte, 10), []byte("("))
	if err == nil {
		t.Fatal("Decode single char: expected error, got nil")
	}
	ce, ok := err.(CorruptInputError)
	if !ok {
		t.Fatalf("Decode single char: expected CorruptInputError, got %T", err)
	}
	if ce.Reason != "incomplete block: single trailing character" {
		t.Errorf("Decode single char: Reason = %q", ce.Reason)
	}
}

// TestDecodeOverflow verifies overflow detection.
func TestDecodeOverflow(t *testing.T) {
	// For a 2-char block, max valid is value 255 = 2*85+84+1? No.
	// 255 = 2*85 + 85 = 3*85. Wait: 255 in base 85 is 3*85+0 = (3)(0).
	// encByte(3)='+', encByte(0)='('.  So "+(" should decode to [255].
	// Let's try a value of 256 = 3*85+1 = (3)(1).
	// encByte(3)='+', encByte(1)=')'. So "+)" should be overflow for 2-char.
	dec := make([]byte, 10)
	_, _, err := Decode(dec, []byte("+)"))
	if err == nil {
		t.Fatal("Decode overflow 2-char: expected error, got nil")
	}
	if ce, ok := err.(CorruptInputError); !ok {
		t.Fatalf("Decode overflow 2-char: expected CorruptInputError, got %T", err)
	} else if ce.Reason != "value overflow in trailing block" {
		t.Errorf("Decode overflow 2-char: Reason = %q", ce.Reason)
	}

	// For 5-char block, max uint32 is 4294967295.
	// 4294967296 in base 85 would be overflow.
	// 84*85^4 + 84*85^3 + 84*85^2 + 84*85 + 84 = 4437053124
	// which is > 4294967295, so the all-84s block overflows.
	allMax := make([]byte, 5)
	for i := range 5 {
		allMax[i] = encByte(84)
	}
	_, _, err = Decode(dec, allMax)
	if err == nil {
		t.Fatal("Decode overflow 5-char: expected error, got nil")
	}
	if ce, ok := err.(CorruptInputError); !ok {
		t.Fatalf("Decode overflow 5-char: expected CorruptInputError, got %T", err)
	} else if ce.Reason != "value overflow in 5-character block" {
		t.Errorf("Decode overflow 5-char: Reason = %q", ce.Reason)
	}
}

// TestEncoderWriter tests the streaming encoder.
func TestEncoderWriter(t *testing.T) {
	src := []byte("Hello, World! This is a test of the encoder.")
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// Write in small chunks to test buffering.
	for i := 0; i < len(src); i += 3 {
		end := i + 3
		if end > len(src) {
			end = len(src)
		}
		n, err := enc.Write(src[i:end])
		if err != nil {
			t.Fatalf("encoder.Write: err = %v", err)
		}
		if n != end-i {
			t.Fatalf("encoder.Write: n = %d, want %d", n, end-i)
		}
	}
	if err := enc.(io.Closer).Close(); err != nil {
		t.Fatalf("encoder.Close: err = %v", err)
	}

	// Decode the result and compare.
	encoded := buf.Bytes()
	dec := make([]byte, len(src))
	ndst, _, err := Decode(dec, encoded)
	if err != nil {
		t.Fatalf("Decode after encoder: err = %v", err)
	}
	if ndst != len(src) {
		t.Fatalf("Decode after encoder: ndst = %d, want %d", ndst, len(src))
	}
	if !bytes.Equal(dec[:ndst], src) {
		t.Errorf("Decode after encoder: got %q, want %q", dec[:ndst], src)
	}
}

// TestDecoderReader tests the streaming decoder.
func TestDecoderReader(t *testing.T) {
	src := []byte("Hello, World! This is a test of the decoder.")
	encoded := make([]byte, MaxEncodedLen(len(src)))
	n := Encode(encoded, src)
	encoded = encoded[:n]

	dec := NewDecoder(bytes.NewReader(encoded))
	got, err := io.ReadAll(dec)
	if err != nil {
		t.Fatalf("ReadAll decoder: err = %v", err)
	}
	if !bytes.Equal(got, src) {
		t.Errorf("decoder: got %q, want %q", got, src)
	}
}

// TestMaxEncodedLen verifies the length function.
func TestMaxEncodedLen(t *testing.T) {
	tests := []struct {
		n, want int
	}{
		{0, 0},
		{1, 2},
		{2, 3},
		{3, 4},
		{4, 5},
		{5, 7},
		{8, 10},
		{100, 125},
	}
	for _, tt := range tests {
		got := MaxEncodedLen(tt.n)
		if got != tt.want {
			t.Errorf("MaxEncodedLen(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

// TestEncodeDstTooShort verifies truncation behavior.
func TestEncodeDstTooShort(t *testing.T) {
	src := []byte{1, 2, 3, 4, 5, 6, 7, 8} // needs 10 encoded bytes
	dst := make([]byte, 5)                  // only room for one block
	n := Encode(dst, src)
	if n != 5 {
		t.Errorf("Encode with short dst: n = %d, want 5", n)
	}
}

// TestMaxDecodedLen verifies the inverse length function.
func TestMaxDecodedLen(t *testing.T) {
	tests := []struct {
		n, want int
	}{
		{0, 0},
		{1, 0}, // 1 char is not a valid block, but max possible decoded is 0
		{2, 1},
		{3, 2},
		{4, 3},
		{5, 4},
		{6, 4}, // 5+1: the trailing 1 can't form a block
		{7, 5},
		{10, 8},
		{125, 100},
	}
	for _, tt := range tests {
		got := MaxDecodedLen(tt.n)
		if got != tt.want {
			t.Errorf("MaxDecodedLen(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

// TestMaxDecodedLenConsistency verifies MaxDecodedLen(MaxEncodedLen(n)) >= n.
func TestMaxDecodedLenConsistency(t *testing.T) {
	for n := range 200 {
		enc := MaxEncodedLen(n)
		dec := MaxDecodedLen(enc)
		if dec < n {
			t.Errorf("MaxDecodedLen(MaxEncodedLen(%d)) = %d, want >= %d", n, dec, n)
		}
	}
}

// TestEncodeToString tests the convenience encoder.
func TestEncodeToString(t *testing.T) {
	src := []byte{0, 1, 2, 3}
	got := EncodeToString(src)
	// Verify by decoding back.
	dec := make([]byte, 4)
	ndst, _, err := Decode(dec, []byte(got))
	if err != nil {
		t.Fatalf("EncodeToString roundtrip: err = %v", err)
	}
	if !bytes.Equal(dec[:ndst], src) {
		t.Errorf("EncodeToString roundtrip: got %v, want %v", dec[:ndst], src)
	}
}

// TestDecodeString tests the convenience decoder.
func TestDecodeString(t *testing.T) {
	src := []byte("Hello!")
	encoded := EncodeToString(src)
	got, err := DecodeString(encoded)
	if err != nil {
		t.Fatalf("DecodeString: err = %v", err)
	}
	if !bytes.Equal(got, src) {
		t.Errorf("DecodeString: got %q, want %q", got, src)
	}
}

// TestDecodeStringError tests that DecodeString returns errors.
func TestDecodeStringError(t *testing.T) {
	_, err := DecodeString("(")
	if err == nil {
		t.Error("DecodeString single char: expected error, got nil")
	}
}

// TestCorruptInputErrorMessage tests the error message format.
func TestCorruptInputErrorMessage(t *testing.T) {
	err := CorruptInputError{Reason: "test reason"}
	want := "r85: test reason"
	if err.Error() != want {
		t.Errorf("CorruptInputError.Error() = %q, want %q", err.Error(), want)
	}
}
