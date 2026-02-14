// Command gen generates random IDs and prints them in r85 encoding.
package main

import (
	"bufio"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/entrope/r85"
)

var (
	n   = flag.Int("n", 1, "number of IDs to generate")
	v4  = flag.Bool("v4", false, "generate UUID v4 (RFC 4122)")
	v7  = flag.Bool("v7", false, "generate UUID v7 (RFC 9562)")
	seq = flag.Bool("seq", false, "sequential: increment from a base random ID")
)

// applyV4 sets the version and variant bits for a UUID v4.
func applyV4(id []byte) {
	id[6] = (id[6] & 0x0F) | 0x40 // version = 0100
	id[8] = (id[8] & 0x3F) | 0x80 // variant = 10
}

// applyV7 writes the millisecond timestamp and sets the version and variant
// bits for a UUID v7.
func applyV7(id []byte, ms uint64) {
	id[0] = byte(ms >> 40)
	id[1] = byte(ms >> 32)
	id[2] = byte(ms >> 24)
	id[3] = byte(ms >> 16)
	id[4] = byte(ms >> 8)
	id[5] = byte(ms)
	id[6] = (id[6] & 0x0F) | 0x70 // version = 0111
	id[8] = (id[8] & 0x3F) | 0x80 // variant = 10
}

// incrementLittleEndian increments a byte slice as a little-endian integer.
// It returns false if the value overflows to zero.
func incrementLittleEndian(id []byte) bool {
	for i := range id {
		id[i]++
		if id[i] != 0 {
			return true
		}
	}
	return false
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("gen: ")
	flag.Parse()

	if *v4 && *v7 {
		fmt.Fprintln(os.Stderr, "gen: -v4 and -v7 are mutually exclusive")
		os.Exit(2)
	}
	if *n < 1 {
		fmt.Fprintln(os.Stderr, "gen: -n must be at least 1")
		os.Exit(2)
	}

	idSize := 12 // 96-bit default
	if *v4 || *v7 {
		idSize = 16
	}

	id := make([]byte, idSize)
	if _, err := rand.Read(id); err != nil {
		log.Fatal(err)
	}

	var ms uint64
	if *v4 {
		applyV4(id)
	} else if *v7 {
		ms = uint64(time.Now().UnixMilli())
		applyV7(id, ms)
	}

	w := bufio.NewWriter(os.Stdout)

	for i := range *n {
		if i > 0 {
			if *seq {
				incr := id
				if *v7 {
					incr = id[6:] // only increment the random portion
				}
				if !incrementLittleEndian(incr) {
					log.Fatal("overflow")
				}
				if *v4 {
					applyV4(id)
				} else if *v7 {
					applyV7(id, ms)
				}
			} else {
				if _, err := rand.Read(id); err != nil {
					log.Fatal(err)
				}
				if *v4 {
					applyV4(id)
				} else if *v7 {
					ms = uint64(time.Now().UnixMilli())
					applyV7(id, ms)
				}
			}
		}

		w.WriteString(r85.EncodeToString(id))
		w.WriteByte('\n')
	}

	if err := w.Flush(); err != nil {
		log.Fatal(err)
	}
}
