package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	"awp/pkg/database"
)

// View renders the UI based on the current mode
func (m Model) View() string {
	var sb strings.Builder

	switch m.mode {
	case NormalMode:
		switch m.viewMode {
		case database.CalendarViewMode:
			// Render the calendar
			sb.WriteString(m.renderCalendar())

		default:
			// Existing table view code
			baseStyle := lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(m.styles.BorderColor))

			// Table with tasks
			sb.WriteString(baseStyle.Render(m.table.View()))
			sb.WriteString("\n")

			// Display view mode and date
			viewInfo := ""

			// Build the view mode part
			var viewModePart string
			switch m.viewMode {
			case database.AllViewMode:
				viewModePart = "all tasks"
			case database.TodayViewMode:
				viewModePart = fmt.Sprintf("tasks due on %s", m.viewDate.Format("2006-01-02"))
			}

			// Build the filter part
			var filterPart string
			switch m.taskFilter {
			case database.AllTasksFilter:
				filterPart = " (no filter)"
			case database.DoneTasksFilter:
				filterPart = " (completed only)"
			case database.UndoneTasksFilter:
				filterPart = " (pending only)"
			}

			// show search filter
			if m.searchTerm != "" {
				filterPart = fmt.Sprintf(" (search filter: %s)", m.searchTerm)
			}

			// Combine the parts
			viewInfo = fmt.Sprintf("Showing %s%s", viewModePart, filterPart)
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.NormalTextColor)).Render(viewInfo))
			sb.WriteString("\n")
		}

	case AddMode:
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Add New Task"))
		sb.WriteString("\n\n")
		sb.WriteString(m.renderForm())

	case EditMode:
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Edit Task"))
		sb.WriteString("\n\n")
		sb.WriteString(m.renderForm())

	case DeleteConfirmMode:
		sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.styles.ErrorColor)).Render("Delete Task"))
		sb.WriteString("\n\n")

		if m.editingItem != nil {
			sb.WriteString(fmt.Sprintf("Are you sure you want to delete this task?\n\n"))
			sb.WriteString(fmt.Sprintf("Title: %s\n", m.editingItem.Title))
			sb.WriteString(fmt.Sprintf("Description: %s\n", m.editingItem.Description))
			sb.WriteString("\n")
			sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Press Y to confirm, N to cancel"))
		}

	case SearchMode:
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Search Tasks"))
		sb.WriteString("\n\n")
		sb.WriteString("Enter search term to find tasks:")
		sb.WriteString("\n\n")
		sb.WriteString(m.searchInput.View())

	case HelpViewMode:
		// Fullscreen commands view
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Available Commands"))
		sb.WriteString("\n\n")

		// Define a style for command keys
		keyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.styles.AccentColor)).
			Bold(true)

		// Define a style for command descriptions
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.styles.NormalTextColor))

		// Function to add a command to the view
		addCommand := func(binding key.Binding) {
			// Get the key and help text
			keyStr := binding.Help().Key
			helpStr := binding.Help().Desc

			// Append the formatted command
			sb.WriteString(fmt.Sprintf("%s: %s\n",
				descStyle.Render(helpStr),
				keyStyle.Render(keyStr)))
		}

		// Add all commands line by line
		addCommand(m.keyMap.QuitApp)
		addCommand(m.keyMap.ShowHelp)
		addCommand(m.keyMap.ToggleStatus)
		addCommand(m.keyMap.AddTask)
		addCommand(m.keyMap.EditTask)
		addCommand(m.keyMap.DeleteTask)
		addCommand(m.keyMap.ToggleViewMode)
		addCommand(m.keyMap.ShowDoneTasks)
		addCommand(m.keyMap.ShowUndoneTasks)
		addCommand(m.keyMap.SearchTasks)
		addCommand(m.keyMap.ToggleCalendarView)

		// Navigation commands
		sb.WriteString("\n")
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Navigation Commands"))
		sb.WriteString("\n\n")

		addCommand(m.keyMap.PrevDay)
		addCommand(m.keyMap.NextDay)
		addCommand(m.keyMap.PrevDayWithTasks)
		addCommand(m.keyMap.NextDayWithTasks)

		// Calendar commands
		sb.WriteString("\n")
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Calendar Commands"))
		sb.WriteString("\n\n")
		addCommand(m.keyMap.CalendarLeft)
		addCommand(m.keyMap.CalendarRight)
		addCommand(m.keyMap.CalendarUp)
		addCommand(m.keyMap.CalendarDown)
		addCommand(m.keyMap.JumpToToday)

	}

	// Error message if any
	if m.err != nil {
		sb.WriteString(fmt.Sprintf("\n\nError: %v", m.err))
	}

	return sb.String()
}

// renderForm renders the input form for adding/editing tasks
func (m Model) renderForm() string {
	var sb strings.Builder

	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.styles.BorderColor)).
		Padding(1, 2).
		Width(m.width - 4)

	// Title input
	sb.WriteString("Title:\n")
	sb.WriteString(m.titleInput.View())
	sb.WriteString("\n\n")

	// Description input
	sb.WriteString("Description:\n")
	sb.WriteString(m.descInput.View())
	sb.WriteString("\n\n")

	// Due date input
	sb.WriteString("Due Date (YYYY-MM-DD):\n")
	sb.WriteString(m.dueDateInput.View())

	return formStyle.Render(sb.String())
}

// renderCalendar renders the calendar view
func (m Model) renderCalendar() string {
	var sb strings.Builder

	// Get the first day of the month
	firstDay := m.calendarMonth

	// Get the last day of the month
	lastDay := firstDay.AddDate(0, 1, 0).AddDate(0, 0, -1)

	// Get the weekday of the first day
	firstWeekday := int(firstDay.Weekday())

	// Calculate how many days are in the month
	daysInMonth := lastDay.Day()

	// Display the month and year as a header
	monthYearHeader := firstDay.Format("January 2006")
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.styles.AccentColor)).Render(monthYearHeader))
	sb.WriteString("\n\n")

	// Display the weekday headers
	weekdays := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	weekdayRow := ""
	for _, day := range weekdays {
		weekdayRow += fmt.Sprintf("%-4s", day)
	}
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(weekdayRow))
	sb.WriteString("\n")

	// Build a map of days that have tasks
	daysWithTasks := make(map[int]bool)

	// Query the database for days in this month that have tasks
	startDateStr := firstDay.Format("2006-01-02")
	endDateStr := lastDay.Format("2006-01-02")

	query := fmt.Sprintf("SELECT DISTINCT strftime('%%d', duedate) FROM todos WHERE date(duedate) BETWEEN date('%s') AND date('%s')", startDateStr, endDateStr)
	rows, err := m.db.Query(query)
	if err != nil {
		sb.WriteString(fmt.Sprintf("Error querying calendar data: %v", err))
		return sb.String()
	}
	defer rows.Close()

	for rows.Next() {
		var dayStr string
		if err := rows.Scan(&dayStr); err != nil {
			continue
		}

		day, err := strconv.Atoi(dayStr)
		if err != nil {
			continue
		}

		daysWithTasks[day] = true
	}

	// Now render the calendar grid
	currentDay := 1

	// Create each row of the calendar
	for week := 0; week < 6; week++ {
		if currentDay > daysInMonth {
			break // We've displayed all days of the month
		}

		// Start a new row
		row := ""

		for weekday := 0; weekday < 7; weekday++ {
			if week == 0 && weekday < firstWeekday {
				// Empty cell before the first day of the month
				row += "    "
			} else if currentDay <= daysInMonth {
				// Determine the style for this day
				dayStyle := lipgloss.NewStyle()

				// Check if this is the selected day (highest priority)
				isSelected := currentDay == m.calendarSelectedDay

				// Highlight the current day
				today := m.viewDate
				isToday := today.Year() == firstDay.Year() &&
					today.Month() == firstDay.Month() &&
					today.Day() == currentDay

				// Highlight days with tasks
				hasTask := daysWithTasks[currentDay]

				if isSelected {
					// Selected day gets highest priority - use background color instead of border
					dayStyle = dayStyle.Background(lipgloss.Color(m.styles.AccentColor)).
						Foreground(lipgloss.Color(m.styles.SelectedTextColor)).Bold(true)
				} else if isToday {
					dayStyle = dayStyle.Background(lipgloss.Color(m.styles.SelectedBgColor)).
						Foreground(lipgloss.Color(m.styles.SelectedTextColor))
				} else if hasTask {
					dayStyle = dayStyle.Foreground(lipgloss.Color(m.styles.AccentColor)).Bold(true)
				}

				// Render the day with appropriate styling
				row += dayStyle.Render(fmt.Sprintf("%-4d", currentDay))

				currentDay++
			} else {
				// Empty cell after the last day of the month
				row += "    "
			}
		}

		sb.WriteString(row)
		sb.WriteString("\n")
	}

	// Add navigation instructions
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.NormalTextColor)).Render(
		"Navigate: ←→↑↓  |  Select day: enter  |  Return to today: esc  |  Exit: ctrl+c"))

	return sb.String()
}
