package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// ConnectDB establishes a connection to the SQLite database
func ConnectDB(dbPath string) (*sql.DB, error) {
	// Expand tilde to home directory if present
	if strings.HasPrefix(dbPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dbPath = homeDir + dbPath[1:]
	}

	// Create the directory structure if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if dbDir != "." {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, err
		}
	}

	// Connect to SQLite database
	// SQLite will create the database file if it doesn't exist
	return sql.Open("sqlite3", dbPath)
}

// EnsureSchema creates the database schema if it doesn't exist
func EnsureSchema(db *sql.DB) error {
	// Create todos table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			status BOOLEAN NOT NULL DEFAULT 0,
			created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			lastmodified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			duedate TIMESTAMP,
			title TEXT NOT NULL,
			description TEXT,
			projects TEXT,
			contexts TEXT
		)
	`)
	return err
}
