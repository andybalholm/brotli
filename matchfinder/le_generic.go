//go:build purego || !(amd64 || 386 || arm64 || loong64 || ppc64le || wasm)

package matchfinder

//go:nosplit
func loadU32LE(b []byte, i uint) uint32 {
	_ = b[i+3]
	return uint32(b[i]) | uint32(b[i+1])<<8 | uint32(b[i+2])<<16 | uint32(b[i+3])<<24
}

//go:nosplit
func loadU64LE(b []byte, i uint) uint64 {
	_ = b[i+7]
	return uint64(b[i]) | uint64(b[i+1])<<8 | uint64(b[i+2])<<16 | uint64(b[i+3])<<24 |
		uint64(b[i+4])<<32 | uint64(b[i+5])<<40 | uint64(b[i+6])<<48 | uint64(b[i+7])<<56
}
