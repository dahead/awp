package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"awp/pkg/database"
)

// HandleExportCommand processes --export commands
func HandleExportCommand(db *sql.DB, filename, exportType string) {
	// Load all tasks
	tasks, err := database.LoadTasks(db, "")
	if err != nil {
		fmt.Printf("Error loading tasks: %v\n", err)
		os.Exit(1)
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	var content []byte

	switch exportType {
	case "json":
		content, err = json.MarshalIndent(tasks, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling tasks to JSON: %v\n", err)
			os.Exit(1)
		}
	case "txt":
		var lines []string
		for _, task := range tasks {
			status := " "
			if task.Status {
				status = "x"
			}
			dateStr := task.DueDate.Format("2006-01-02")
			lines = append(lines, fmt.Sprintf("[%s] %s: %s", status, dateStr, task.Title))
		}
		content = []byte(strings.Join(lines, "\n"))
	default:
		fmt.Printf("Unknown export type: %s\n", exportType)
		os.Exit(1)
	}

	if err := os.WriteFile(filename, content, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully exported %d task(s) to %s\n", len(tasks), filename)
}
