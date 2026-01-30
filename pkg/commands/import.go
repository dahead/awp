package commands

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"awp/pkg/database"
)

// HandleImportCommand processes --import commands
func HandleImportCommand(db *sql.DB, filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(content), "\n")
	var currentDate time.Time
	var tasksAdded int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line contains a date (DD.MM.YYYY: or YYYY-MM-DD: format)
		dateRegex := regexp.MustCompile(`(?:(\d{2})\.(\d{2})\.(\d{4})|(\d{4})-(\d{2})-(\d{2})):?`)
		if dateMatch := dateRegex.FindStringSubmatch(line); dateMatch != nil {
			var day, month, year int
			if dateMatch[1] != "" {
				day, _ = strconv.Atoi(dateMatch[1])
				month, _ = strconv.Atoi(dateMatch[2])
				year, _ = strconv.Atoi(dateMatch[3])
			} else {
				year, _ = strconv.Atoi(dateMatch[4])
				month, _ = strconv.Atoi(dateMatch[5])
				day, _ = strconv.Atoi(dateMatch[6])
			}
			currentDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			continue
		}

		// Check if line is a task (starts with -)
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, " - ") {
			taskText := strings.TrimPrefix(strings.TrimSpace(line), "- ")
			if taskText == "" {
				continue
			}

			status := false
			if strings.HasPrefix(taskText, "[x]") {
				status = true
				taskText = strings.TrimSpace(strings.TrimPrefix(taskText, "[x]"))
			} else if strings.HasPrefix(taskText, "[ ]") {
				status = false
				taskText = strings.TrimSpace(strings.TrimPrefix(taskText, "[ ]"))
			}

			// Extract projects and contexts
			projects := extractProjects(taskText)
			contexts := extractContexts(taskText)

			// Clean title
			title := removeProjectTags(taskText)
			title = removeContextTags(title)

			task := database.TodoItem{
				Status:      status,
				Title:       title,
				Description: taskText,
				DueDate:     currentDate,
				Projects:    projects,
				Contexts:    contexts,
			}

			if err := database.AddTask(db, task); err != nil {
				fmt.Printf("Error adding task '%s': %v\n", title, err)
				continue
			}
			tasksAdded++
		}
	}

	fmt.Printf("Successfully imported %d task(s) from %s\n", tasksAdded, filename)
}
