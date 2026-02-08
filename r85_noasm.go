//go:build !arm64 || purego

package r85

const haveNEON = false

func encodeBlocksNEON(dst *byte, src *byte)       {}
func decodeBlocksNEON(dst *byte, src *byte) uint64 { return 1 }
