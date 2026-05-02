package util

import (
	"github.com/spaolacci/murmur3"
)

// MurmurHash3Guava computes MurmurHash3 32-bit compatible with Guava's
// Hashing.murmur3_32().hashUnencodedChars(param).padToLong().
// Guava hashes UTF-16 code units (2 bytes per char, little-endian), NOT UTF-8 bytes.
func MurmurHash3Guava(s string) uint64 {
	// Convert string to UTF-16 LE byte sequence (BMP chars only for ASCII/Latin)
	utf16le := stringToUTF16LE(s)
	h := murmur3.Sum32(utf16le)
	return uint64(h) // zero-extend to 64-bit (padToLong)
}

// stringToUTF16LE converts a Go string (UTF-8) to UTF-16 Little-Endian bytes.
// Each rune is encoded as 2 bytes (BMP) or 4 bytes (surrogate pair for non-BMP).
func stringToUTF16LE(s string) []byte {
	buf := make([]byte, 0, len(s)*2)
	for _, r := range s {
		if r <= 0xFFFF {
			// BMP character: 2 bytes LE
			buf = append(buf, byte(r), byte(r>>8))
		} else {
			// Non-BMP: surrogate pair
			r -= 0x10000
			high := 0xD800 + (r>>10)&0x3FF
			low := 0xDC00 + r&0x3FF
			buf = append(buf, byte(high), byte(high>>8))
			buf = append(buf, byte(low), byte(low>>8))
		}
	}
	return buf
}

// JavaStringHashCode implements Java's String.hashCode() algorithm.
// Used for sharding routing in ShardingDBConfig and ShardingTableConfig.
// Processes UTF-16 code units (same as Java char).
func JavaStringHashCode(s string) int32 {
	var h int32
	for _, r := range s {
		// Java processes char values (UTF-16 code units)
		h = 31*h + int32(r&0xFFFF)
	}
	return h
}

// AbsInt32Mod returns a non-negative modulo result compatible with
// Math.abs(code.hashCode()) % n in Java. Handles the MIN_VALUE edge case.
func AbsInt32Mod(hash int32, n int) int {
	if hash >= 0 {
		return int(hash) % n
	}
	// For negative values, use unsigned interpretation to get absolute value
	return int(uint32(hash) % uint32(n))
}
