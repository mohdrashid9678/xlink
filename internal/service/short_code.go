package service

import (
	"crypto/rand"
	"math/big"
)

const (
	shortCodeLength   = 8
	shortCodeAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func generateShortCode() (string, error) {
	code := make([]byte, shortCodeLength)
	upperBound := big.NewInt(int64(len(shortCodeAlphabet)))

	for i := range code {
		index, err := rand.Int(rand.Reader, upperBound)
		if err != nil {
			return "", err
		}
		code[i] = shortCodeAlphabet[index.Int64()]
	}

	return string(code), nil
}
