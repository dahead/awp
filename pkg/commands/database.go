package commands

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

// HandleDatabaseCommand processes --database commands
func HandleDatabaseCommand(db *sql.DB, cmd, dateStr, projectStr string, skipConfirm, doneOnly, undoneOnly bool) {
	if cmd != "purge" {
		fmt.Printf("Unknown database command: %s\n", cmd)
		os.Exit(1)
	}

	// Build where clause for deletion
	whereClause := buildPurgeWhereClause(dateStr, projectStr, doneOnly, undoneOnly)

	// Show confirmation unless --yes flag is used
	if !skipConfirm {
		fmt.Print("Are you sure you want to delete these tasks? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Operation cancelled.")
			return
		}
	}

	// Execute deletion
	query := "DELETE FROM todos"
	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	result, err := db.Exec(query)
	if err != nil {
		fmt.Printf("Error purging tasks: %v\n", err)
		os.Exit(1)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Successfully deleted %d task(s)\n", rowsAffected)
}

// buildPurgeWhereClause builds WHERE clause for purge operations
func buildPurgeWhereClause(dateStr, projectStr string, doneOnly, undoneOnly bool) string {
	var conditions []string

	if dateStr != "" {
		conditions = append(conditions, fmt.Sprintf("date(duedate) = date('%s')", dateStr))
	}

	if projectStr != "" {
		conditions = append(conditions, fmt.Sprintf("projects LIKE '%%%s%%'", projectStr))
	}

	if doneOnly {
		conditions = append(conditions, "status = 1")
	} else if undoneOnly {
		conditions = append(conditions, "status = 0")
	}

	return strings.Join(conditions, " AND ")
}
