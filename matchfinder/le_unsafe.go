//go:build !purego && (amd64 || 386 || arm || arm64 || loong64 || mips64le || mipsle || ppc64le || riscv64 || wasm)

package matchfinder

import "unsafe"

//go:nosplit
func loadU32LE(b []byte, i uint) uint32 {
	return *(*uint32)(unsafe.Add(unsafe.Pointer(unsafe.SliceData(b)), i))
}

//go:nosplit
func loadU64LE(b []byte, i uint) uint64 {
	return *(*uint64)(unsafe.Add(unsafe.Pointer(unsafe.SliceData(b)), i))
}
