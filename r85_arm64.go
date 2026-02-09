//go:build arm64 && !purego

package r85

const haveSIMD = true

//go:noescape
func encodeBlocksNEON(dst *byte, src *byte)

//go:noescape
func decodeBlocksNEON(dst *byte, src *byte) uint64

func encodeBlocksSIMD(dst *byte, src *byte)       { encodeBlocksNEON(dst, src) }
func decodeBlocksSIMD(dst *byte, src *byte) uint64 { return decodeBlocksNEON(dst, src) }
