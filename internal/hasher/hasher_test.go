package hasher

import (
	"path/filepath"
	"testing"
)

func TestHashFile(t *testing.T) {
	t.Helper()

	const want = "945d7dc72fce0e2a9a7b1b02e9434d7142b1e71699f630a0d6f01f7a89f1b410"

	path := filepath.Join("..", "..", "testdata", "sample.jpg")

	got, err := HashFile(path)
	if err != nil {
		t.Fatalf("HashFile returned error: %v", err)
	}
	if got != want {
		t.Fatalf("HashFile: got %q, want %q", got, want)
	}

	again, err := HashFile(path)
	if err != nil {
		t.Fatalf("HashFile second call returned error: %v", err)
	}
	if again != want {
		t.Fatalf("HashFile second call: got %q, want %q", again, want)
	}
}

// TestHashFileDeterminism verifies that the same file always produces the same hash.
func TestHashFileDeterminism(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "sample.jpg")

	hash1, err := HashFile(path)
	if err != nil {
		t.Fatalf("First HashFile call failed: %v", err)
	}

	hash2, err := HashFile(path)
	if err != nil {
		t.Fatalf("Second HashFile call failed: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("HashFile produced different hashes for same file: %q != %q", hash1, hash2)
	}
}

// TestHashFileNotFound tests error handling for non-existent files.
func TestHashFileNotFound(t *testing.T) {
	_, err := HashFile("/nonexistent/file.jpg")
	if err == nil {
		t.Errorf("HashFile should fail on non-existent file, but got nil error")
	}
}

// TestHashFileHexFormat verifies that the returned hash is valid lowercase hex (64 chars for SHA-256).
func TestHashFileHexFormat(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "sample.jpg")

	hash, err := HashFile(path)
	if err != nil {
		t.Fatalf("HashFile failed: %v", err)
	}

	// SHA-256 hex is always 64 characters
	if len(hash) != 64 {
		t.Errorf("SHA-256 hash length: got %d, want 64", len(hash))
	}

	// Verify all characters are valid hex digits
	for i, c := range hash {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("Invalid hex character at position %d: %c", i, c)
		}
	}
}
