package database

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

func CreateTestClient(t *testing.T) (*Client, error) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	dbURL := "file:" + dbPath

	client, err := NewClient(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create test database client: %v", err)
	}

	cmd := exec.Command("goose", "turso", dbURL, "up", "--dir", "./migrations")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to migrate up test db: %v\nOutput:\n%s", err, string(output))
	}

	return client, nil
}
func TestDBMigrations(t *testing.T) {
	c, err := CreateTestClient(t)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer c.db.Close()
}
