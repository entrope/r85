package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/entrope/r85"
)

func TestGenerate96(t *testing.T) {
	var id [12]byte
	generate96(id[:])
	if id == [12]byte{} {
		t.Error("generate96 produced all-zero ID")
	}
}

func TestGenerate96Length(t *testing.T) {
	var id [12]byte
	generate96(id[:])
	s := r85.EncodeToString(id[:])
	if len(s) != 15 {
		t.Errorf("encoded 96-bit ID length = %d, want 15", len(s))
	}
}

func TestGenerateV4Bits(t *testing.T) {
	for i := range 100 {
		var id [16]byte
		generateV4(id[:])
		if id[6]>>4 != 4 {
			t.Errorf("iteration %d: version nibble = %x, want 4", i, id[6]>>4)
		}
		if id[8]>>6 != 2 {
			t.Errorf("iteration %d: variant bits = %d, want 2", i, id[8]>>6)
		}
	}
}

func TestGenerateV4Length(t *testing.T) {
	var id [16]byte
	generateV4(id[:])
	s := r85.EncodeToString(id[:])
	if len(s) != 20 {
		t.Errorf("encoded UUID v4 length = %d, want 20", len(s))
	}
}

func TestGenerateV7Bits(t *testing.T) {
	before := time.Now().UnixMilli()
	var id [16]byte
	msec := time.Now().UnixMilli()
	generateV7(id[:], msec)
	after := time.Now().UnixMilli()

	if id[6]>>4 != 7 {
		t.Errorf("version nibble = %x, want 7", id[6]>>4)
	}
	if id[8]>>6 != 2 {
		t.Errorf("variant bits = %d, want 2", id[8]>>6)
	}

	ts := int64(id[0])<<40 | int64(id[1])<<32 | int64(id[2])<<24 |
		int64(id[3])<<16 | int64(id[4])<<8 | int64(id[5])
	if ts < before || ts > after {
		t.Errorf("timestamp = %d, want in [%d, %d]", ts, before, after)
	}
}

func TestGenerateV7Length(t *testing.T) {
	var id [16]byte
	generateV7(id[:], time.Now().UnixMilli())
	s := r85.EncodeToString(id[:])
	if len(s) != 20 {
		t.Errorf("encoded UUID v7 length = %d, want 20", len(s))
	}
}

func TestIncrement(t *testing.T) {
	tests := []struct {
		in   []byte
		want []byte
		ok   bool
	}{
		{[]byte{0, 0, 0}, []byte{0, 0, 1}, true},
		{[]byte{0, 0, 0xFF}, []byte{0, 1, 0}, true},
		{[]byte{0, 0xFF, 0xFF}, []byte{1, 0, 0}, true},
		{[]byte{0xFF, 0xFF, 0xFF}, []byte{0, 0, 0}, false},
	}
	for _, tt := range tests {
		b := make([]byte, len(tt.in))
		copy(b, tt.in)
		ok := increment(b)
		if ok != tt.ok {
			t.Errorf("increment(%x): ok = %v, want %v", tt.in, ok, tt.ok)
		}
		if !bytes.Equal(b, tt.want) {
			t.Errorf("increment(%x) = %x, want %x", tt.in, b, tt.want)
		}
	}
}

func TestIncrementUUIDv4PreservesBits(t *testing.T) {
	var id [16]byte
	generateV4(id[:])
	for i := range 1000 {
		ok := incrementUUIDv4(id[:])
		if !ok {
			t.Fatalf("overflow at iteration %d", i)
		}
		if id[6]>>4 != 4 {
			t.Fatalf("after %d increments: version = %x, want 4", i+1, id[6]>>4)
		}
		if id[8]>>6 != 2 {
			t.Fatalf("after %d increments: variant = %d, want 2", i+1, id[8]>>6)
		}
	}
}

func TestIncrementUUIDv7CarryRandBToRandA(t *testing.T) {
	// Set up a v7 UUID where rand_b is at max (62 bits all 1).
	// Bytes 9-15 = 0xFF, byte 8 low 6 bits = 0x3F.
	id := [16]byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, // timestamp
		0x70, 0x00, // version=7, rand_a=0x000
		0xBF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // variant=10, rand_b=max
	}
	ok := incrementUUIDv7(id[:])
	if !ok {
		t.Fatal("unexpected overflow")
	}
	// rand_b should wrap to 0, rand_a should become 0x001.
	if id[7] != 0x01 {
		t.Errorf("byte 7 = %02x, want 01 (rand_a incremented)", id[7])
	}
	if id[8]&0x3F != 0x00 {
		t.Errorf("byte 8 low 6 = %02x, want 00 (rand_b wrapped)", id[8]&0x3F)
	}
	if id[8]>>6 != 2 {
		t.Errorf("variant = %d, want 2", id[8]>>6)
	}
	if id[6] != 0x70 {
		t.Errorf("byte 6 = %02x, want 70 (version preserved, rand_a low nibble 0)", id[6])
	}
	// Timestamp unchanged.
	if id[0] != 0x01 || id[5] != 0x06 {
		t.Errorf("timestamp modified")
	}
}

func TestIncrementUUIDv7CarryRandAOverflow(t *testing.T) {
	// Both rand_a (12 bits) and rand_b (62 bits) at max.
	id := [16]byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, // timestamp
		0x7F, 0xFF, // version=7, rand_a=0xFFF
		0xBF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // variant=10, rand_b=max
	}
	ok := incrementUUIDv7(id[:])
	if ok {
		t.Fatal("expected overflow when all 74 random bits are max")
	}
}

func TestIncrementUUIDv4CarryRandBToRandA(t *testing.T) {
	// v4 UUID where rand_b is at max, rand_a=0x000, rand_hi=0.
	id := [16]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // rand_hi=0
		0x40, 0x00, // version=4, rand_a=0x000
		0xBF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // variant=10, rand_b=max
	}
	ok := incrementUUIDv4(id[:])
	if !ok {
		t.Fatal("unexpected overflow")
	}
	if id[7] != 0x01 {
		t.Errorf("byte 7 = %02x, want 01 (rand_a incremented)", id[7])
	}
	if id[8]&0x3F != 0x00 {
		t.Errorf("byte 8 low 6 = %02x, want 00 (rand_b wrapped)", id[8]&0x3F)
	}
}

func TestIncrementUUIDv4CarryToRandHi(t *testing.T) {
	// v4 UUID where rand_a and rand_b are both max, rand_hi=0.
	id := [16]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // rand_hi=0
		0x4F, 0xFF, // version=4, rand_a=0xFFF
		0xBF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // variant=10, rand_b=max
	}
	ok := incrementUUIDv4(id[:])
	if !ok {
		t.Fatal("unexpected overflow")
	}
	// rand_a and rand_b should wrap, rand_hi incremented to 1.
	if id[5] != 0x01 {
		t.Errorf("byte 5 = %02x, want 01 (rand_hi incremented)", id[5])
	}
	if id[6]&0x0F != 0x00 || id[7] != 0x00 {
		t.Errorf("rand_a = %x%02x, want 000", id[6]&0x0F, id[7])
	}
	if id[8]&0x3F != 0x00 {
		t.Errorf("rand_b low bits = %02x, want 00", id[8]&0x3F)
	}
	if id[6]>>4 != 4 {
		t.Errorf("version = %x, want 4", id[6]>>4)
	}
	if id[8]>>6 != 2 {
		t.Errorf("variant = %d, want 2", id[8]>>6)
	}
}

func TestIncrementUUIDv7PreservesBits(t *testing.T) {
	var id [16]byte
	msec := time.Now().UnixMilli()
	generateV7(id[:], msec)
	origTS := make([]byte, 6)
	copy(origTS, id[0:6])

	for i := range 1000 {
		ok := incrementUUIDv7(id[:])
		if !ok {
			t.Fatalf("overflow at iteration %d", i)
		}
		if id[6]>>4 != 7 {
			t.Fatalf("after %d increments: version = %x, want 7", i+1, id[6]>>4)
		}
		if id[8]>>6 != 2 {
			t.Fatalf("after %d increments: variant = %d, want 2", i+1, id[8]>>6)
		}
		if !bytes.Equal(id[0:6], origTS) {
			t.Fatalf("after %d increments: timestamp changed from %x to %x", i+1, origTS, id[0:6])
		}
	}
}
