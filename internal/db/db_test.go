package db

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestOpenCreatesDatabaseAndSchema(t *testing.T) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "exif.db")

	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := database.Close(); closeErr != nil {
			t.Fatalf("Close returned error: %v", closeErr)
		}
	})

	assertTableExists(t, database, "images")
	assertPragmaValue(t, database, "journal_mode", "wal")
	assertPragmaValue(t, database, "foreign_keys", int64(1))
	assertPragmaValue(t, database, "synchronous", int64(1))
	assertPragmaValue(t, database, "busy_timeout", int64(5000))
	assertPragmaValue(t, database, "cache_size", int64(-64000))
}

func assertTableExists(t *testing.T, database *sql.DB, tableName string) {
	t.Helper()

	var count int
	err := database.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, tableName).Scan(&count)
	if err != nil {
		t.Fatalf("QueryRow table exists: %v", err)
	}
	if count != 1 {
		t.Fatalf("table %q does not exist", tableName)
	}
}

func assertPragmaValue(t *testing.T, database *sql.DB, pragma string, want any) {
	t.Helper()

	var got any
	err := database.QueryRow("PRAGMA " + pragma).Scan(&got)
	if err != nil {
		t.Fatalf("QueryRow pragma %s: %v", pragma, err)
	}
	if got != want {
		t.Fatalf("PRAGMA %s: got %v, want %v", pragma, got, want)
	}
}
