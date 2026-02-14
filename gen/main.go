package main

import (
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/entrope/r85"
)

func main() {
	n := flag.Int("n", 1, "number of IDs to generate")
	uuidVer := flag.String("uuid", "", "UUID version: v4 or v7 (omit for 96-bit random)")
	seq := flag.Bool("seq", false, "sequential IDs from a random base")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n", flag.Args())
		os.Exit(2)
	}
	if *n < 1 {
		fmt.Fprintf(os.Stderr, "-n must be at least 1\n")
		os.Exit(2)
	}

	var (
		size int
		gen  func([]byte)
		inc  func([]byte) bool
	)

	switch *uuidVer {
	case "":
		size = 12
		gen = generate96
		inc = func(b []byte) bool { return increment(b) }
	case "v4":
		size = 16
		gen = generateV4
		inc = func(b []byte) bool { return incrementUUIDv4(b) }
	case "v7":
		size = 16
		msec := time.Now().UnixMilli()
		gen = func(b []byte) { generateV7(b, msec) }
		inc = func(b []byte) bool { return incrementUUIDv7(b) }
	default:
		fmt.Fprintf(os.Stderr, "unknown -uuid value: %q (use v4 or v7)\n", *uuidVer)
		os.Exit(2)
	}

	id := make([]byte, size)
	gen(id)
	fmt.Println(r85.EncodeToString(id))

	for i := 1; i < *n; i++ {
		if *seq {
			if !inc(id) {
				fmt.Fprintf(os.Stderr, "ID overflow at index %d\n", i)
				os.Exit(1)
			}
		} else {
			gen(id)
		}
		fmt.Println(r85.EncodeToString(id))
	}
}

// generate96 fills b (12 bytes) with cryptographic random bytes.
func generate96(b []byte) {
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand: " + err.Error())
	}
}

// generateV4 fills b (16 bytes) with a UUID v4: 122 random bits,
// version=0100, variant=10.
func generateV4(b []byte) {
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	b[6] = (b[6] & 0x0F) | 0x40 // version 4
	b[8] = (b[8] & 0x3F) | 0x80 // variant 10
}

// generateV7 fills b (16 bytes) with a UUID v7: 48-bit millisecond
// timestamp, version=0111, variant=10, 74 random bits.
func generateV7(b []byte, msec int64) {
	binary.BigEndian.PutUint32(b[0:4], uint32(msec>>16))
	binary.BigEndian.PutUint16(b[4:6], uint16(msec))
	if _, err := rand.Read(b[6:16]); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	b[6] = (b[6] & 0x0F) | 0x70 // version 7
	b[8] = (b[8] & 0x3F) | 0x80 // variant 10
}

// increment adds 1 to a big-endian byte slice.
// Returns true on success, false if the value wrapped to zero.
func increment(b []byte) bool {
	for i := len(b) - 1; i >= 0; i-- {
		b[i]++
		if b[i] != 0 {
			return true
		}
	}
	return false
}

// incrementUUID increments the random bits of a UUID, carrying across the
// version and variant fields. The random bits form three segments:
//
//	bytes 0-5:              rand_hi (48 bits) — absent in v7 (timestamp)
//	byte 6 low 4 + byte 7: rand_a  (12 bits)
//	byte 8 low 6 + bytes 9-15: rand_b (62 bits)
//
// Increment rand_b first; on overflow carry into rand_a; on overflow carry
// into rand_hi (for v4) or report overflow (for v7).
func incrementUUID(b []byte, version int) bool {
	// Increment rand_b: 62 bits in byte 8 (low 6) + bytes 9-15.
	if increment(b[9:16]) {
		return true
	}
	// bytes 9-15 wrapped; carry into byte 8 low 6 bits.
	lo6 := b[8] & 0x3F
	lo6++
	if lo6 <= 0x3F {
		b[8] = (b[8] & 0xC0) | lo6
		return true
	}
	// rand_b overflowed; zero its high bits and carry into rand_a.
	b[8] = b[8] & 0xC0

	// Increment rand_a: 12 bits in byte 6 (low 4) + byte 7.
	b[7]++
	if b[7] != 0 {
		return true
	}
	lo4 := b[6] & 0x0F
	lo4++
	if lo4 <= 0x0F {
		b[6] = (b[6] & 0xF0) | lo4
		return true
	}
	// rand_a overflowed; zero its high bits.
	b[6] = b[6] & 0xF0

	// v7: timestamp is not part of the random space.
	if version == 7 {
		return false
	}
	// v4: carry into rand_hi (bytes 0-5).
	return increment(b[0:6])
}

// incrementUUIDv4 increments the 122 random bits of a UUID v4.
func incrementUUIDv4(b []byte) bool {
	return incrementUUID(b, 4)
}

// incrementUUIDv7 increments the 74 random bits of a UUID v7.
// The 48-bit timestamp in bytes 0-5 is preserved.
func incrementUUIDv7(b []byte) bool {
	return incrementUUID(b, 7)
}
