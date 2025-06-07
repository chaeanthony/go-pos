package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	SEED_DIR = "./seed"
)

func seed(db *sql.DB, filename string) error {
	sqlBytes, err := os.ReadFile(filepath.Join(SEED_DIR, filename))
	if err != nil {
		return err
	}
	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return err
	}

	return nil
}

func TestItems(t *testing.T) {
	c, err := CreateTestClient(t)
	require.NoError(t, err, "Failed to create test database")
	defer c.db.Close()

	err = seed(c.db, "items/items.sql")
	require.NoError(t, err, "Failed to seed items table")
	t.Run("Seed items table", func(t *testing.T) {
		query := `SELECT COUNT(*) FROM items`
		var count int
		err = c.db.QueryRow(query).Scan(&count)
		require.NoError(t, err, "Failed to query items")
		assert.Greater(t, count, 0, "Expected items table to have rows")
	})
}
