package commands

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"awp/pkg/database"
)

// HandleAddTask processes the --add command
func HandleAddTask(db *sql.DB, taskText string, dateStr string) {
	// Parse date
	var dueDate time.Time
	var err error

	if dateStr != "" {
		dueDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			fmt.Printf("Error parsing date: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Default to today
		dueDate = time.Now()
	}

	// Extract projects from task text (format: +project)
	projects := extractProjects(taskText)

	// Remove project tags from title for clean display
	title := removeProjectTags(taskText)

	// Create task
	task := database.TodoItem{
		Status:      false,
		Title:       title,
		Description: taskText, // Keep original text in description
		DueDate:     dueDate,
		Projects:    projects,
		Contexts:    []string{}, // Empty contexts for now
	}

	if err := database.AddTask(db, task); err != nil {
		fmt.Printf("Error adding task: %v\n", err)
		os.Exit(1)
	}

	// fmt.Printf("Task added successfully: %s\n", title)
}

// extractProjects finds all +project tags in text
func extractProjects(text string) []string {
	re := regexp.MustCompile(`\+(\w+)`)
	matches := re.FindAllStringSubmatch(text, -1)
	var projects []string
	for _, match := range matches {
		projects = append(projects, match[1])
	}
	return projects
}

// removeProjectTags removes +project tags from text for clean title
func removeProjectTags(text string) string {
	re := regexp.MustCompile(`\s*\+\w+\s*`)
	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}
