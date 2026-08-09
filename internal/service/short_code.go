package service

import (
	"crypto/md5"
	"encoding/binary"

	"github.com/google/uuid"
)

const (
	shortCodeLength   = 7
	shortCodeAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	base62            = uint64(len(shortCodeAlphabet))
)

func generateShortCode(id uuid.UUID, attempt uint64) string {
	return encodeBase62(hashUUID(id, attempt))
}

func hashUUID(id uuid.UUID, attempt uint64) uint64 {
	// MD5 is used as a fast standard-library distribution function, not as a
	// security primitive. UUID randomness provides the code input entropy.
	var input [24]byte
	copy(input[:16], id[:])
	binary.BigEndian.PutUint64(input[16:], attempt)

	sum := md5.Sum(input[:])
	return binary.BigEndian.Uint64(sum[:8])
}

func encodeBase62(value uint64) string {
	code := make([]byte, shortCodeLength)
	for i := shortCodeLength - 1; i >= 0; i-- {
		code[i] = shortCodeAlphabet[value%base62]
		value /= base62
	}
	return string(code)
}
