package service

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestEncodeBase62UsesFixedSevenCharacterWidth(t *testing.T) {
	tests := []struct {
		value    uint64
		expected string
	}{
		{value: 0, expected: "0000000"},
		{value: 1, expected: "0000001"},
		{value: base62 - 1, expected: "000000z"},
	}

	for _, test := range tests {
		if actual := encodeBase62(test.value); actual != test.expected {
			t.Errorf("encodeBase62(%d) = %q, want %q", test.value, actual, test.expected)
		}
	}
}

func TestGenerateShortCodeIsBase62AndFixedLength(t *testing.T) {
	code := generateShortCode(uuid.MustParse("8a72efb5-10d4-4d28-9223-e8951be3f226"), 0)

	if len(code) != shortCodeLength {
		t.Fatalf("code length = %d, want %d", len(code), shortCodeLength)
	}
	for _, character := range code {
		if !strings.ContainsRune(shortCodeAlphabet, character) {
			t.Fatalf("code %q contains non-Base62 character %q", code, character)
		}
	}
}

func TestGenerateShortCodeChangesForCollisionRetry(t *testing.T) {
	id := uuid.MustParse("8a72efb5-10d4-4d28-9223-e8951be3f226")
	firstCode := generateShortCode(id, 0)
	retryCode := generateShortCode(id, 1)

	if firstCode == retryCode {
		t.Fatalf("collision retry generated the same short code %q", firstCode)
	}
}
