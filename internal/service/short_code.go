package service

import (
	"github.com/google/uuid"
)

const (
	shortCodeLength   = 7
	shortCodeAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	base62            = uint64(len(shortCodeAlphabet))

	// FNV-1a is a fast, non-cryptographic hash suitable for distributing a
	// randomly generated UUID over the short-code key space.
	fnvOffsetBasis64 = uint64(14695981039346656037)
	fnvPrime64       = uint64(1099511628211)
)

func generateShortCode(id uuid.UUID, attempt uint64) string {
	return encodeBase62(hashUUID(id, attempt))
}

func hashUUID(id uuid.UUID, attempt uint64) uint64 {
	hash := fnvOffsetBasis64
	for _, value := range id {
		hash ^= uint64(value)
		hash *= fnvPrime64
	}

	// The retry attempt gives each collision retry a different deterministic code.
	for range 8 {
		hash ^= attempt & 0xff
		hash *= fnvPrime64
		attempt >>= 8
	}

	return hash
}

func encodeBase62(value uint64) string {
	code := make([]byte, shortCodeLength)
	for i := shortCodeLength - 1; i >= 0; i-- {
		code[i] = shortCodeAlphabet[value%base62]
		value /= base62
	}
	return string(code)
}
