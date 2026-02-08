//go:build arm64 && !purego

package r85

const haveNEON = true

//go:noescape
func encodeBlocksNEON(dst *byte, src *byte)

//go:noescape
func decodeBlocksNEON(dst *byte, src *byte) uint64
