package r85

import (
	"bytes"
	"io"
	"testing"
)

// benchSizes covers a range from small to large payloads.
var benchSizes = []struct {
	name string
	n    int
}{
	{"16B", 16},
	{"256B", 256},
	{"1KB", 1024},
	{"16KB", 16384},
	{"256KB", 256 * 1024},
	{"1MB", 1024 * 1024},
}

func makeSrc(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i*37 + 13)
	}
	return b
}

// BenchmarkEncode benchmarks the direct Encode function.
func BenchmarkEncode(b *testing.B) {
	for _, sz := range benchSizes {
		src := makeSrc(sz.n)
		dst := make([]byte, MaxEncodedLen(sz.n))
		b.Run(sz.name, func(b *testing.B) {
			b.SetBytes(int64(sz.n))
			for b.Loop() {
				Encode(dst, src)
			}
		})
	}
}

// BenchmarkDecode benchmarks the direct Decode function.
func BenchmarkDecode(b *testing.B) {
	for _, sz := range benchSizes {
		src := makeSrc(sz.n)
		enc := make([]byte, MaxEncodedLen(sz.n))
		Encode(enc, src)
		dst := make([]byte, sz.n)
		b.Run(sz.name, func(b *testing.B) {
			b.SetBytes(int64(sz.n))
			for b.Loop() {
				Decode(dst, enc)
			}
		})
	}
}

// BenchmarkEncoder benchmarks the streaming NewEncoder writing all data
// in a single Write call.
func BenchmarkEncoder(b *testing.B) {
	for _, sz := range benchSizes {
		src := makeSrc(sz.n)
		b.Run(sz.name, func(b *testing.B) {
			b.SetBytes(int64(sz.n))
			for b.Loop() {
				var buf bytes.Buffer
				buf.Grow(MaxEncodedLen(sz.n))
				enc := NewEncoder(&buf)
				enc.Write(src)
				enc.(io.Closer).Close()
			}
		})
	}
}

// BenchmarkDecoder benchmarks the streaming NewDecoder reading all data.
func BenchmarkDecoder(b *testing.B) {
	for _, sz := range benchSizes {
		src := makeSrc(sz.n)
		enc := make([]byte, MaxEncodedLen(sz.n))
		Encode(enc, src)
		b.Run(sz.name, func(b *testing.B) {
			b.SetBytes(int64(sz.n))
			dst := make([]byte, sz.n)
			for b.Loop() {
				dec := NewDecoder(bytes.NewReader(enc))
				io.ReadFull(dec, dst)
			}
		})
	}
}

// BenchmarkEncoderChunked writes data in fixed-size chunks through the
// streaming encoder to measure the overhead of the 4-byte internal
// buffer and 5-byte writes to the underlying writer.
func BenchmarkEncoderChunked(b *testing.B) {
	chunkSizes := []struct {
		name string
		n    int
	}{
		{"Chunk1", 1},
		{"Chunk3", 3},
		{"Chunk4", 4},
		{"Chunk5", 5},
		{"Chunk16", 16},
		{"Chunk256", 256},
		{"Chunk4096", 4096},
	}
	const dataSize = 16384
	src := makeSrc(dataSize)

	for _, cs := range chunkSizes {
		b.Run(cs.name, func(b *testing.B) {
			b.SetBytes(int64(dataSize))
			for b.Loop() {
				var buf bytes.Buffer
				buf.Grow(MaxEncodedLen(dataSize))
				enc := NewEncoder(&buf)
				for i := 0; i < len(src); i += cs.n {
					end := min(i+cs.n, len(src))
					enc.Write(src[i:end])
				}
				enc.(io.Closer).Close()
			}
		})
	}
}

// countingWriter counts Write calls and total bytes written,
// without buffering. This isolates the encoder's write pattern.
type countingWriter struct {
	calls int
	bytes int
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.calls++
	w.bytes += len(p)
	return len(p), nil
}

// BenchmarkEncoderWriteCalls measures the overhead of frequent small
// writes from the encoder to the underlying writer. By comparing
// against a countingWriter (which does almost nothing), we can see
// how much time the encoder spends on the Write calls themselves.
func BenchmarkEncoderWriteCalls(b *testing.B) {
	for _, sz := range benchSizes {
		src := makeSrc(sz.n)
		b.Run(sz.name, func(b *testing.B) {
			b.SetBytes(int64(sz.n))
			var w countingWriter
			for b.Loop() {
				w.calls = 0
				w.bytes = 0
				enc := NewEncoder(&w)
				enc.Write(src)
				enc.(io.Closer).Close()
			}
			// Report the write pattern for analysis.
			b.ReportMetric(float64(w.calls), "writes/op")
			b.ReportMetric(float64(w.bytes)/float64(w.calls), "bytes/write")
		})
	}
}

// BenchmarkEncodeVsEncoder directly compares Encode to NewEncoder
// overhead at a fixed size to make the cost ratio easy to see.
func BenchmarkEncodeVsEncoder(b *testing.B) {
	const dataSize = 4096
	src := makeSrc(dataSize)

	b.Run("Encode", func(b *testing.B) {
		dst := make([]byte, MaxEncodedLen(dataSize))
		b.SetBytes(dataSize)
		for b.Loop() {
			Encode(dst, src)
		}
	})

	b.Run("Encoder", func(b *testing.B) {
		b.SetBytes(dataSize)
		for b.Loop() {
			var buf bytes.Buffer
			buf.Grow(MaxEncodedLen(dataSize))
			enc := NewEncoder(&buf)
			enc.Write(src)
			enc.(io.Closer).Close()
		}
	})

	b.Run("Encoder/countingWriter", func(b *testing.B) {
		b.SetBytes(dataSize)
		var w countingWriter
		for b.Loop() {
			enc := NewEncoder(&w)
			enc.Write(src)
			enc.(io.Closer).Close()
		}
	})
}

// BenchmarkDecodeVsDecoder directly compares Decode to NewDecoder
// overhead at a fixed size.
func BenchmarkDecodeVsDecoder(b *testing.B) {
	const dataSize = 4096
	src := makeSrc(dataSize)
	enc := make([]byte, MaxEncodedLen(dataSize))
	Encode(enc, src)

	b.Run("Decode", func(b *testing.B) {
		dst := make([]byte, dataSize)
		b.SetBytes(dataSize)
		for b.Loop() {
			Decode(dst, enc)
		}
	})

	b.Run("Decoder", func(b *testing.B) {
		b.SetBytes(dataSize)
		dst := make([]byte, dataSize)
		for b.Loop() {
			dec := NewDecoder(bytes.NewReader(enc))
			io.ReadFull(dec, dst)
		}
	})
}
