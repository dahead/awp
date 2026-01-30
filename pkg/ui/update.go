package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"awp/pkg/database"
	"awp/pkg/utils"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case NormalMode:
			switch {
			case key.Matches(msg, m.keyMap.ShowHelp):
				m.mode = HelpViewMode

			case key.Matches(msg, m.keyMap.QuitApp):
				return m, tea.Quit

			case key.Matches(msg, m.keyMap.JumpToToday):
				m.loadTodaysTasks()

			case key.Matches(msg, m.keyMap.ToggleStatus):
				if len(m.items) > 0 && m.table.Cursor() < len(m.items) {
					// Update in database
					idx := m.table.Cursor()
					if idx < len(m.items) {
						m.items[idx].Status = !m.items[idx].Status
						// Update status using the database function
						err := database.UpdateTaskStatus(m.db, m.items[idx].ID, m.items[idx].Status)
						if err != nil {
							m.err = err
						} else {
							// Update the display
							selectedRow := m.table.SelectedRow()
							statusPrefix := "[ ]"
							if m.items[idx].Status {
								statusPrefix = "[x]"
							}

							// Extract the text part (everything after the status)
							text := m.items[idx].Title

							// Highlight project and context tags in the text
							highlightedText := highlightProjectsAndContexts(text, m.styles)

							// Create the new row with highlighted text
							selectedRow[0] = fmt.Sprintf("%s %s", statusPrefix, highlightedText)
							rows := m.table.Rows()
							rows[m.table.Cursor()] = selectedRow
							m.table.SetRows(rows)
						}
					}
				}

			case key.Matches(msg, m.keyMap.AddTask):
				m.mode = AddMode
				m.resetInputs()

			case key.Matches(msg, m.keyMap.EditTask):
				if len(m.items) > 0 && m.table.Cursor() < len(m.items) {
					m.mode = EditMode
					m.editingItem = &m.items[m.table.Cursor()]
					m.resetInputs()

					// Populate form with existing values
					m.titleInput.SetValue(m.editingItem.Title)
					m.descInput.SetValue(m.editingItem.Description)

					// Format and set due date
					if !m.editingItem.DueDate.IsZero() {
						m.dueDateInput.SetValue(m.editingItem.DueDate.Format("2006-01-02"))
					}
				}

			case key.Matches(msg, m.keyMap.DeleteTask):
				if len(m.items) > 0 && m.table.Cursor() < len(m.items) {
					m.mode = DeleteConfirmMode
					m.editingItem = &m.items[m.table.Cursor()]
				}

			case key.Matches(msg, m.keyMap.ToggleViewMode):
				// Toggle between today's tasks and all tasks
				if m.viewMode == database.TodayViewMode {
					m.viewMode = database.AllViewMode
				} else {
					m.viewMode = database.TodayViewMode
				}
				m.loadTasks()

			case key.Matches(msg, m.keyMap.PrevDay):
				if m.viewMode == database.TodayViewMode {
					m.viewDate = m.viewDate.AddDate(0, 0, -1)
					m.loadTasks()
				}

			case key.Matches(msg, m.keyMap.NextDay):
				if m.viewMode == database.TodayViewMode {
					m.viewDate = m.viewDate.AddDate(0, 0, 1)
					m.loadTasks()
				}

			case key.Matches(msg, m.keyMap.PrevDayWithTasks):
				if m.viewMode == database.TodayViewMode {
					m.findPrevDayWithTasks()
				}

			case key.Matches(msg, m.keyMap.NextDayWithTasks):
				if m.viewMode == database.TodayViewMode {
					m.findNextDayWithTasks()
				}

			case key.Matches(msg, m.keyMap.ShowDoneTasks):
				// Toggle between done tasks and all tasks
				if m.taskFilter == database.DoneTasksFilter {
					m.taskFilter = database.AllTasksFilter
				} else {
					m.taskFilter = database.DoneTasksFilter
				}
				m.loadTasks()

			case key.Matches(msg, m.keyMap.ShowUndoneTasks):
				// Toggle between undone tasks and all tasks
				if m.taskFilter == database.UndoneTasksFilter {
					m.taskFilter = database.AllTasksFilter
				} else {
					m.taskFilter = database.UndoneTasksFilter
				}
				m.loadTasks()

			case key.Matches(msg, m.keyMap.SearchTasks):
				// Enter search mode
				m.mode = SearchMode
				m.searchInput.Focus()
				m.searchInput.SetValue("") // Clear previous search
				return m, nil

			case key.Matches(msg, m.keyMap.ToggleSortBy):
				m.sortBy = (m.sortBy + 1) % 7 // Cycle through all sort options
				m.loadTasks()

			case key.Matches(msg, m.keyMap.ToggleGroupBy):
				m.groupBy = (m.groupBy + 1) % 7 // Cycle through all group options
				m.loadTasks()

			case key.Matches(msg, m.keyMap.ToggleSortOrder):
				if m.sortOrder == database.SortAsc {
					m.sortOrder = database.SortDesc
				} else {
					m.sortOrder = database.SortAsc
				}
				m.loadTasks()

			case key.Matches(msg, m.keyMap.ToggleCalendarView):
				// Toggle calendar view mode
				if m.viewMode == database.CalendarViewMode {
					m.viewMode = database.TodayViewMode
				} else {
					m.viewMode = database.CalendarViewMode
				}
				m.loadTasks()

			// Calendar navigation (only when in calendar view)
			case key.Matches(msg, m.keyMap.CalendarLeft) && m.viewMode == database.CalendarViewMode:
				if m.calendarSelectedDay > 1 {
					m.calendarSelectedDay--
				} else {
					// Move to previous month and set to last day
					m.calendarMonth = m.calendarMonth.AddDate(0, -1, 0)
					lastDay := time.Date(m.calendarMonth.Year(), m.calendarMonth.Month()+1, 0, 0, 0, 0, 0, m.calendarMonth.Location())
					m.calendarSelectedDay = lastDay.Day()
				}

			case key.Matches(msg, m.keyMap.CalendarRight) && m.viewMode == database.CalendarViewMode:
				lastDay := time.Date(m.calendarMonth.Year(), m.calendarMonth.Month()+1, 0, 0, 0, 0, 0, m.calendarMonth.Location())
				if m.calendarSelectedDay < lastDay.Day() {
					m.calendarSelectedDay++
				} else {
					// Move to next month and set to first day
					m.calendarMonth = m.calendarMonth.AddDate(0, 1, 0)
					m.calendarSelectedDay = 1
				}

			case key.Matches(msg, m.keyMap.CalendarUp) && m.viewMode == database.CalendarViewMode:
				newDay := m.calendarSelectedDay - 7
				if newDay < 1 {
					// Move to previous month
					m.calendarMonth = m.calendarMonth.AddDate(0, -1, 0)
					lastDay := time.Date(m.calendarMonth.Year(), m.calendarMonth.Month()+1, 0, 0, 0, 0, 0, m.calendarMonth.Location())
					m.calendarSelectedDay = lastDay.Day() + newDay
					if m.calendarSelectedDay < 1 {
						m.calendarSelectedDay = 1
					}
				} else {
					m.calendarSelectedDay = newDay
				}

			case key.Matches(msg, m.keyMap.CalendarDown) && m.viewMode == database.CalendarViewMode:
				lastDay := time.Date(m.calendarMonth.Year(), m.calendarMonth.Month()+1, 0, 0, 0, 0, 0, m.calendarMonth.Location())
				newDay := m.calendarSelectedDay + 7
				if newDay > lastDay.Day() {
					// Move to next month
					m.calendarMonth = m.calendarMonth.AddDate(0, 1, 0)
					m.calendarSelectedDay = newDay - lastDay.Day()
				} else {
					m.calendarSelectedDay = newDay
				}

			case key.Matches(msg, m.keyMap.CalendarSelect) && m.viewMode == database.CalendarViewMode:
				// Jump to selected day in today view
				selectedDate := time.Date(m.calendarMonth.Year(), m.calendarMonth.Month(), m.calendarSelectedDay, 0, 0, 0, 0, m.calendarMonth.Location())
				m.viewDate = selectedDate
				m.viewMode = database.TodayViewMode
				m.loadTasks()

			case msg.String() == "esc" && m.viewMode == database.CalendarViewMode:
				// Return to today view from calendar
				m.viewDate = time.Now()
				m.viewMode = database.TodayViewMode
				m.loadTasks()

			case msg.String() == "/":
				// Enter search mode when "/" is pressed
				m.mode = SearchMode
				m.searchInput.Focus()
				m.searchInput.SetValue("") // Clear previous search
				return m, nil
			}

		case AddMode, EditMode:
			switch msg.String() {
			case "esc":
				m.mode = NormalMode
				m.resetInputs()
				m.editingItem = nil

			case "tab":
				m.focusNextInput()

			case "shift+tab":
				m.focusPreviousInput()

			case "enter":
				if m.activeInput == 2 { // Submit on enter from the last field (due date)
					m.submitForm()
				} else {
					m.focusNextInput()
				}
			}

			// Handle input updates
			switch m.activeInput {
			case 0:
				m.titleInput, cmd = m.titleInput.Update(msg)
				cmds = append(cmds, cmd)
			case 1:
				m.descInput, cmd = m.descInput.Update(msg)
				cmds = append(cmds, cmd)
			case 2:
				m.dueDateInput, cmd = m.dueDateInput.Update(msg)
				cmds = append(cmds, cmd)
			}

		case SearchMode:
			// Handle search mode key presses
			switch msg.String() {
			case "esc":
				// Exit search mode
				m.mode = NormalMode
				m.searchTerm = ""
				m.loadTasks()

			case "enter":
				// Perform search
				m.searchTerm = m.searchInput.Value()
				utils.Log("Searching for: %s", m.searchTerm)
				m.mode = NormalMode
				m.loadTasks()
			}

			// Update search input
			m.searchInput, cmd = m.searchInput.Update(msg)
			cmds = append(cmds, cmd)

		case DeleteConfirmMode:
			// Handle delete confirmation
			switch msg.String() {
			case "y", "Y":
				if m.editingItem != nil {
					utils.Log("Deleting task ID: %d", m.editingItem.ID)
					// Delete from database using the database function
					err := database.DeleteTask(m.db, m.editingItem.ID)
					if err != nil {
						utils.Log("Error deleting task: %v", err)
						m.err = err
					} else {
						utils.Log("Task deleted successfully")
						m.loadTodaysTasks()
					}
				}
				m.mode = NormalMode
				m.editingItem = nil

			case "n", "N", "esc":
				m.mode = NormalMode
				m.editingItem = nil
			}

		case HelpViewMode:
			switch msg.String() {
			case "esc":
				// Exit commands view mode
				m.mode = NormalMode

			case "ctrl+b":
				// Exit commands view mode
				m.mode = NormalMode
			}
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.table.SetWidth(msg.Width - 4)
		m.table.SetHeight(msg.Height - 4)
	}

	// Only update table in normal mode
	if m.mode == NormalMode {
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	// tea.ClearScreen()

	// Force table to recalculate its display area
	// m.table.SetHeight(m.table.Height())

	return m, tea.Batch(cmds...)
}
