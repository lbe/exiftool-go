// Package hasher provides file hashing utilities for computing and storing file checksums.
// It uses SHA-256 for cryptographic hashing of file contents.

package hasher

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// HashFile computes the SHA-256 hash of a file and returns the hash as a lowercase hex string.
// If the file cannot be read, an error is returned.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
