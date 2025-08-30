package ui

import (
	"database/sql"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"awp/pkg/config"
	"awp/pkg/database"
	"awp/pkg/keymaps"
)

// InputMode represents the current input mode
type InputMode int

const (
	NormalMode InputMode = iota
	AddMode
	EditMode
	DeleteConfirmMode
	SearchMode   // Mode for searching tasks
	HelpViewMode // Mode for displaying help
)

// Model represents the application state
type Model struct {
	table         table.Model
	items         []database.TodoItem
	db            *sql.DB
	showCommands  bool
	width, height int
	err           error

	// Configuration
	config config.Config
	styles config.Styles
	keyMap keymaps.KeyMap

	// View state
	viewMode   database.ViewMode
	taskFilter database.TaskFilter
	viewDate   time.Time
	searchTerm string

	// Form state
	mode         InputMode
	titleInput   textinput.Model
	descInput    textinput.Model
	dueDateInput textinput.Model
	searchInput  textinput.Model
	activeInput  int

	// Edit/delete state
	editingItem *database.TodoItem

	// Sorting and grouping state
	sortBy    database.SortBy
	groupBy   database.GroupBy
	sortOrder database.SortOrder

	calendarMonth       time.Time
	calendarSelectedDay int // Selected day in calendar view (1-31)
}

// NewModel creates a new UI model with the provided configuration
func NewModel(db *sql.DB, cfg config.Config, styles config.Styles) Model {
	// Create an empty column - the title will be empty to avoid showing a header
	columns := []table.Column{
		{Title: "", Width: 60},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Set table styles using the loaded styles
	s := table.DefaultStyles()
	// Remove the header border and styling to make it invisible
	s.Header = s.Header.
		BorderStyle(lipgloss.HiddenBorder()). // Hidden border
		BorderBottom(false).                  // No border at bottom
		Bold(false).                          // Not bold
		Foreground(lipgloss.NoColor{})        // No color (transparent)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color(styles.SelectedTextColor)).
		Background(lipgloss.Color(styles.SelectedBgColor)).
		Bold(true)
	t.SetStyles(s)

	// Initialize text inputs
	titleInput := textinput.New()
	titleInput.Placeholder = "Title (you can include +project and @context tags)"
	titleInput.Focus()
	titleInput.Width = 40

	descInput := textinput.New()
	descInput.Placeholder = "Description"
	descInput.Width = 40

	// Initialize due date input with today's date as default
	dueDateInput := textinput.New()
	dueDateInput.Placeholder = "Due Date (YYYY-MM-DD, optional)"
	dueDateInput.Width = 40
	dueDateInput.SetValue(time.Now().Format("2006-01-02"))

	// Initialize search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search tasks (you can use +project or @context)"
	searchInput.Focus()
	searchInput.Width = 40

	m := Model{
		table:               t,
		db:                  db,
		config:              cfg,
		styles:              styles,
		keyMap:              keymaps.BuildKeyMap(cfg.KeyMap),
		showCommands:        false,
		mode:                NormalMode,
		titleInput:          titleInput,
		descInput:           descInput,
		dueDateInput:        dueDateInput,
		searchInput:         searchInput,
		activeInput:         0,
		viewMode:            database.TodayViewMode,  // Default view mode shows today's tasks
		taskFilter:          database.AllTasksFilter, // Default to showing all tasks (both done and undone)
		viewDate:            time.Now(),
		searchTerm:          "", // Initialize empty search term
		calendarMonth:       time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location()),
		calendarSelectedDay: time.Now().Day(), // Initialize to today's day
	}

	// Load initial data
	m.loadTodaysTasks()

	return m
}

// Init initializes the model (required by Bubble Tea Model interface)
func (m Model) Init() tea.Cmd {
	return nil
}

// resetInputs clears all form inputs
func (m *Model) resetInputs() {
	m.titleInput.Reset()
	m.descInput.Reset()
	m.dueDateInput.SetValue(m.viewDate.Format("2006-01-02"))

	m.activeInput = 0
	m.titleInput.Focus()
	m.descInput.Blur()
	m.dueDateInput.Blur()
}
