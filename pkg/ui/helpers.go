package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"awp/pkg/config"
	"awp/pkg/database"
)

// loadTasks retrieves and displays tasks based on current filters
func (m *Model) loadTasks() {
	var items []database.TodoItem
	var err error

	// Build where clause using the database package function
	dateStr := m.viewDate.Format("2006-01-02")
	whereClause := database.BuildWhereClause(m.viewMode, m.taskFilter, dateStr, m.searchTerm)

	// Load the tasks with the combined where clause
	items, err = database.LoadTasks(m.db, whereClause)

	if err != nil {
		m.err = err
		return
	}

	m.items = items

	// Apply grouping and sorting
	groupedTasks := m.GroupTasks(items)
	tableRows := []table.Row{}

	for _, group := range groupedTasks {
		// Add group header if grouping is enabled
		if m.groupBy != database.GroupByNone {
			groupHeader := fmt.Sprintf("== %s ==", group.GroupName)
			tableRows = append(tableRows, table.Row{
				lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color(m.styles.AccentColor)).
					Render(groupHeader),
			})
		}

		// Add tasks in the group
		for _, item := range group.Tasks {
			status := "[ ]"
			if item.Status {
				status = "[x]"
			}

			displayText := item.Description
			if item.Title != "" {
				displayText = item.Title
			}

			highlightedText := highlightProjectsAndContexts(displayText, m.styles)
			combinedText := fmt.Sprintf("%s %s", status, highlightedText)
			tableRows = append(tableRows, table.Row{combinedText})
		}

		// Add empty line between groups
		if m.groupBy != database.GroupByNone && len(groupedTasks) > 1 {
			tableRows = append(tableRows, table.Row{""})
		}
	}

	m.table.SetRows(tableRows)
}

// For backward compatibility
func (m *Model) loadTodaysTasks() {
	m.viewDate = time.Now()
	m.viewMode = database.TodayViewMode
	m.loadTasks()
}

// findPrevDayWithTasks finds the previous day that has tasks and updates viewDate
func (m *Model) findPrevDayWithTasks() {
	// Start from the day before current viewDate
	startDate := m.viewDate.AddDate(0, 0, -1)

	// Store original filter to restore it later
	originalFilter := m.taskFilter

	// Set filter to show all tasks to make sure we find any task
	m.taskFilter = database.AllTasksFilter

	// Keep looking back one day at a time until we find a day with tasks
	// We'll limit the search to a year back to avoid infinite loops
	for i := 0; i < 365; i++ {
		testDate := startDate.AddDate(0, 0, -i)
		dateStr := testDate.Format("2006-01-02")

		// Query the database directly to check if there are tasks for this date
		query := fmt.Sprintf("SELECT COUNT(*) FROM todos WHERE date(duedate) = date('%s')", dateStr)
		row := m.db.QueryRow(query)

		var count int
		if err := row.Scan(&count); err != nil {
			m.err = err
			break
		}

		// If we found tasks for this date, update viewDate and load the tasks
		if count > 0 {
			m.viewDate = testDate
			m.loadTasks()

			// Restore original filter
			m.taskFilter = originalFilter
			m.loadTasks()
			return
		}
	}

	// If no day with tasks was found, just restore the filter
	m.taskFilter = originalFilter
	m.loadTasks()
}

// findNextDayWithTasks finds the next day that has tasks and updates viewDate
func (m *Model) findNextDayWithTasks() {
	// Start from the day after current viewDate
	startDate := m.viewDate.AddDate(0, 0, 1)

	// Store original filter to restore it later
	originalFilter := m.taskFilter

	// Set filter to show all tasks to make sure we find any task
	m.taskFilter = database.AllTasksFilter

	// Keep looking forward one day at a time until we find a day with tasks
	// We'll limit the search to a year ahead to avoid infinite loops
	for i := 0; i < 365; i++ {
		testDate := startDate.AddDate(0, 0, i)
		dateStr := testDate.Format("2006-01-02")

		// Query the database directly to check if there are tasks for this date
		query := fmt.Sprintf("SELECT COUNT(*) FROM todos WHERE date(duedate) = date('%s')", dateStr)
		row := m.db.QueryRow(query)

		var count int
		if err := row.Scan(&count); err != nil {
			m.err = err
			break
		}

		// If we found tasks for this date, update viewDate and load the tasks
		if count > 0 {
			m.viewDate = testDate
			m.loadTasks()

			// Restore original filter
			m.taskFilter = originalFilter
			m.loadTasks()
			return
		}
	}

	// If no day with tasks was found, just restore the filter
	m.taskFilter = originalFilter
	m.loadTasks()
}

// focusNextInput cycles through the form inputs
func (m *Model) focusNextInput() {
	m.activeInput = (m.activeInput + 1) % 3
	switch m.activeInput {
	case 0:
		m.titleInput.Focus()
		m.descInput.Blur()
		m.dueDateInput.Blur()
	case 1:
		m.titleInput.Blur()
		m.descInput.Focus()
		m.dueDateInput.Blur()
	case 2:
		m.titleInput.Blur()
		m.descInput.Blur()
		m.dueDateInput.Focus()
	}
}

// focusPreviousInput cycles through the form inputs
func (m *Model) focusPreviousInput() {
	m.activeInput = (m.activeInput - 1) % 3

	switch m.activeInput {
	case 0:
		m.titleInput.Focus()
		m.descInput.Blur()
		m.dueDateInput.Blur()
	case 1:
		m.titleInput.Blur()
		m.descInput.Focus()
		m.dueDateInput.Blur()
	case 2:
		m.titleInput.Blur()
		m.descInput.Blur()
		m.dueDateInput.Focus()
	}
}

// submitForm processes the form data based on the current mode
func (m *Model) submitForm() {
	title := strings.TrimSpace(m.titleInput.Value())
	desc := strings.TrimSpace(m.descInput.Value())
	dueDate := strings.TrimSpace(m.dueDateInput.Value())

	// Parse projects and contexts from title and description
	projects := parseProjects(title)
	projects = append(projects, parseProjects(desc)...)
	contexts := parseContexts(title)
	contexts = append(contexts, parseContexts(desc)...)

	// Parse due date
	var parsedDueDate time.Time
	var err error
	if dueDate != "" {
		parsedDueDate, err = time.Parse("2006-01-02", dueDate)
		if err != nil {
			m.err = fmt.Errorf("invalid date format: use YYYY-MM-DD")
			return
		}
	} else {
		// Default to current views date if no date provided
		parsedDueDate = m.viewDate
	}

	switch m.mode {
	case AddMode:
		// Create new task with the collected data
		task := database.TodoItem{
			Status:      false,
			DueDate:     parsedDueDate,
			Title:       title,
			Description: desc,
			Projects:    projects,
			Contexts:    contexts,
		}

		// Insert new task using the database function
		err := database.AddTask(m.db, task)
		if err != nil {
			m.err = err
		} else {
			m.loadTodaysTasks()
		}

	case EditMode:
		if m.editingItem != nil {
			// Update task with new values
			m.editingItem.Title = title
			m.editingItem.Description = desc
			m.editingItem.DueDate = parsedDueDate
			m.editingItem.Projects = projects
			m.editingItem.Contexts = contexts

			// Update using the database function
			err := database.UpdateTask(m.db, *m.editingItem)
			if err != nil {
				m.err = err
			} else {
				m.loadTodaysTasks()
			}
		}
	}

	// Reset state
	m.mode = NormalMode
	m.resetInputs()
	m.editingItem = nil
}

// parseProjects extracts all project tags (prefixed with +) from the description
func parseProjects(description string) []string {
	var projects []string
	words := strings.Fields(description)

	for _, word := range words {
		if strings.HasPrefix(word, "+") && len(word) > 1 {
			project := word[1:] // Remove the + prefix
			projects = append(projects, project)
		}
	}

	return projects
}

// parseContexts extracts all context tags (prefixed with @) from the description
func parseContexts(description string) []string {
	var contexts []string
	words := strings.Fields(description)

	for _, word := range words {
		if strings.HasPrefix(word, "@") && len(word) > 1 {
			context := word[1:] // Remove the @ prefix
			contexts = append(contexts, context)
		}
	}

	return contexts
}

// highlightProjectsAndContexts highlights project and context tags in text
func highlightProjectsAndContexts(text string, styles config.Styles) string {
	// Split the text into words
	words := strings.Fields(text)
	var result strings.Builder

	// Process each word
	for i, word := range words {
		if i > 0 {
			result.WriteString(" ") // Add space between words
		}

		// Check if word is a project tag (+project)
		if strings.HasPrefix(word, "+") && len(word) > 1 {
			// Highlight project with a different color (green)
			result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ProjectColor)).Render(word))
		} else if strings.HasPrefix(word, "@") && len(word) > 1 {
			// Highlight context with a different color (blue)
			result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ContextColor)).Render(word))
		} else {
			// Regular word, no highlighting
			result.WriteString(word)
		}
	}

	return result.String()
}
