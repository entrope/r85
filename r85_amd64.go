//go:build amd64 && !purego

package r85

import "golang.org/x/sys/cpu"

const haveSIMD = true

var haveAVX512 = cpu.X86.HasAVX512F && cpu.X86.HasAVX512BW && cpu.X86.HasAVX512VL

//go:noescape
func encodeBlocksAVX2(dst *byte, src *byte)

//go:noescape
func decodeBlocksAVX2(dst *byte, src *byte) uint64

//go:noescape
func encodeBlocksAVX512(dst *byte, src *byte)

//go:noescape
func decodeBlocksAVX512(dst *byte, src *byte) uint64

func encodeBlocksSIMD(dst *byte, src *byte) {
	if haveAVX512 {
		encodeBlocksAVX512(dst, src)
	} else {
		encodeBlocksAVX2(dst, src)
	}
}

func decodeBlocksSIMD(dst *byte, src *byte) uint64 {
	if haveAVX512 {
		return decodeBlocksAVX512(dst, src)
	}
	return decodeBlocksAVX2(dst, src)
}
