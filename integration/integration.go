//go:build integration
// +build integration

package integration

import (
	"math/rand/v2"
)

func generateRandomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyz"

	//nolint:varnamelen
	b := make([]byte, length)

	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}

	return string(b)
}

func generateRandomNamespace(prefix string) string {
	// Generate a random 12-character string
	return prefix + "-" + generateRandomString(12)
}
