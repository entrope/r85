package r85

import "io"

func MaxEncodedLen(n int) int {
	res := 5 * (n / 4)
	if r := n % 4; r != 0 {
		res += r + 1
	}
	return res
}

// Encode encodes binary `src` into text `dst`, returning the number of
// bytes written to `dst`.
// If `dst` is too short, Encode fills `dst` and returns `len(dst)`,
// and ignores the remaining input.
func Encode(dst, src []byte) int {}

// Decode decodes text `src` into binary `dst`.
// `ndst` contains the number of bytes written into `dst`.
// `nsrc` contains the number of bytes consumed from `src`, including
// skipped bytes.
// If `dst` is too short, Decode fills `dst` and returns `len(dst)`,
// and ignores the remaining input.
func Decode(dst, src []byte) (ndst, nsrc int, err error) {}

// NewEncoder wraps a buffer and io.WriteCloser interface around `Encode`.
// This will only write a short block (less than 4 bytes of binary input)
// when Close is called.
func NewEncoder(w io.Writer) io.WriteCloser {}

// NewDecoder wraps a buffer and io.Reader interface around `Decode`.
func NewDecoder(r io.Reader) io.Reader {}

type CorruptInputError struct{}

func (e CorruptInputError) Error() string {
	return "invalid input length"
}
