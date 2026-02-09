//go:build (!arm64 && !amd64) || purego

package r85

const haveSIMD = false

func encodeBlocksSIMD(dst *byte, src *byte)       {}
func decodeBlocksSIMD(dst *byte, src *byte) uint64 { return 1 }
