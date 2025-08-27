package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/mattn/go-sqlite3"
)

// Color constants for consistent styling across the application
const (
	// UI element colors
	ColorBorder = "240" // Border color for UI elements
	ColorAccent = "205" // Pink accent color used in multiple places

	// Text colors
	ColorNormalText   = "86"  // Cyan text for help and view info
	ColorSelectedText = "229" // Yellow text for selected items
	ColorSelectedBg   = "57"  // Blue background for selected items
	ColorError        = "9"   // Red text for error/delete messages

	// Project and context colors
	ColorProject = "2" // Green for project tags
	ColorContext = "4" // Blue for context tags
)

// Config holds the application configuration
type Config struct {
	Database string            `json:"database"`
	KeyMap   map[string]string `json:"keymap"`
}

// TodoItem represents a single todo task
type TodoItem struct {
	ID           int       `db:"id"`
	Status       bool      `db:"status"`
	Title        string    `db:"title"`
	Description  string    `db:"description"`
	Created      time.Time `db:"created"`
	LastModified time.Time `db:"lastmodified"`
	DueDate      time.Time `db:"duedate"`
	Projects     []string  `db:"projects"`
	Contexts     []string  `db:"contexts"`
}

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

// ViewMode represents the current view mode for tasks
type ViewMode int

const (
	TodayViewMode ViewMode = iota // Default - show tasks for today
	AllViewMode                   // Show all tasks (no date filter)
)

// TaskFilter represents the current task filter mode
type TaskFilter int

const (
	AllTasksFilter    TaskFilter = iota // Show all tasks regardless of status
	DoneTasksFilter                     // Show only completed tasks
	UndoneTasksFilter                   // Show only uncompleted tasks
)

// Model represents the application state
type Model struct {
	table         table.Model
	items         []TodoItem
	db            *sql.DB
	showCommands  bool
	width, height int
	err           error

	// View state
	viewMode   ViewMode
	taskFilter TaskFilter
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
	editingItem *TodoItem
}

// Define keymaps
type keyMap struct {
	ShowHelp         key.Binding
	QuitApp          key.Binding
	ToggleStatus     key.Binding
	AddTask          key.Binding
	EditTask         key.Binding
	DeleteTask       key.Binding
	ToggleViewMode   key.Binding
	ShowDoneTasks    key.Binding
	ShowUndoneTasks  key.Binding
	SearchTasks      key.Binding
	PrevDay          key.Binding
	NextDay          key.Binding
	PrevDayWithTasks key.Binding
	NextDayWithTasks key.Binding
	JumpToToday      key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		ShowHelp: key.NewBinding(
			key.WithKeys("ctrl+b"),
			key.WithHelp("ctrl+b", "Show help"),
		),
		QuitApp: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		ToggleStatus: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "toggle status"),
		),
		AddTask: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add task"),
		),
		EditTask: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit task"),
		),
		DeleteTask: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete task"),
		),
		ToggleViewMode: key.NewBinding(
			key.WithKeys("ctrl+v"),
			key.WithHelp("ctrl+v", "toggle between today's tasks and all tasks"),
		),
		ShowDoneTasks: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "show only done tasks"),
		),
		ShowUndoneTasks: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "show only undone tasks"),
		),
		SearchTasks: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("ctrl+f", "search tasks"),
		),
		PrevDay: key.NewBinding(
			key.WithKeys("ctrl+left"),
			key.WithHelp("ctrl+←", "previous day"),
		),
		NextDay: key.NewBinding(
			key.WithKeys("ctrl+right"),
			key.WithHelp("ctrl+→", "next day"),
		),
		PrevDayWithTasks: key.NewBinding(
			key.WithKeys("ctrl+shift+left"),
			key.WithHelp("ctrl+shift+←", "previous day with tasks"),
		),
		NextDayWithTasks: key.NewBinding(
			key.WithKeys("ctrl+shift+right"),
			key.WithHelp("ctrl+shift+→", "next day with tasks"),
		),
		JumpToToday: key.NewBinding( // New key binding
			key.WithKeys("h"),
			key.WithHelp("h", "jump to today"),
		),
	}
}

func configuredKeyMap(config Config) keyMap {
	// Parse key bindings from config
	log("Parsing key bindings from configuration")
	km := keyMap{
		ShowHelp:         parseKeyBinding(config.KeyMap["ShowHelp"], "ctrl+b", "show/hide commands"),
		QuitApp:          parseKeyBinding(config.KeyMap["QuitApp"], "q", "quit"),
		ToggleStatus:     parseKeyBinding(config.KeyMap["ToggleStatus"], "t", "toggle status"),
		AddTask:          parseKeyBinding(config.KeyMap["AddTask"], "a", "add task"),
		EditTask:         parseKeyBinding(config.KeyMap["EditTask"], "e", "edit task"),
		DeleteTask:       parseKeyBinding(config.KeyMap["DeleteTask"], "d", "delete task"),
		ToggleViewMode:   parseKeyBinding(config.KeyMap["ToggleViewMode"], "ctrl+v", "toggle between today's tasks and all tasks"),
		ShowDoneTasks:    parseKeyBinding(config.KeyMap["ShowDoneTasks"], "ctrl+d", "show only done tasks"),
		ShowUndoneTasks:  parseKeyBinding(config.KeyMap["ShowUndoneTasks"], "ctrl+u", "show only undone tasks"),
		SearchTasks:      parseKeyBinding(config.KeyMap["SearchTasks"], "ctrl+f", "search tasks"),
		PrevDay:          parseKeyBinding(config.KeyMap["PrevDay"], "ctrl+left", "previous day"),
		NextDay:          parseKeyBinding(config.KeyMap["NextDay"], "ctrl+right", "next day"),
		PrevDayWithTasks: parseKeyBinding(config.KeyMap["PrevDayWithTasks"], "ctrl+shift+left", "previous day with tasks"),
		NextDayWithTasks: parseKeyBinding(config.KeyMap["NextDayWithTasks"], "ctrl+shift+right", "next day with tasks"),
		JumpToToday:      parseKeyBinding(config.KeyMap["JumpToToday"], "ctrl+shift+right", "next day with tasks"),
	}
	log("Finished parsing key bindings")
	return km
}

func parseKeyBinding(configKey, defaultKey, help string) key.Binding {
	log("Parsing key binding for '%s'", help)

	if configKey == "" {
		log("No configured key for '%s', using default: %s", help, defaultKey)
		configKey = defaultKey
	} else {
		log("Using configured key for '%s': %s", help, configKey)
	}

	// Handle JSON array format ["key1", "key2", "key3"] by removing brackets and quotes
	if strings.HasPrefix(configKey, "[") && strings.HasSuffix(configKey, "]") {
		// Remove the brackets
		configKey = strings.TrimPrefix(configKey, "[")
		configKey = strings.TrimSuffix(configKey, "]")

		// Remove quotes and split by comma
		var keys []string
		parts := strings.Split(configKey, ",")
		for _, part := range parts {
			// Trim whitespace and quotes
			k := strings.Trim(part, " \"'")
			if k != "" {
				keys = append(keys, k)
			}
		}

		log("Parsed keys from JSON array for '%s': %v", help, keys)

		binding := key.NewBinding(
			key.WithKeys(keys...),
			key.WithHelp(strings.Join(keys, "/"), help),
		)

		log("Created key binding for '%s'", help)
		return binding
	}

	// Handle space-separated keys (original behavior)
	keys := strings.Fields(configKey)

	// Also handle comma-separated keys without brackets
	if len(keys) == 1 && strings.Contains(configKey, ",") {
		var commaKeys []string
		parts := strings.Split(configKey, ",")
		for _, part := range parts {
			k := strings.TrimSpace(part)
			if k != "" {
				commaKeys = append(commaKeys, k)
			}
		}

		if len(commaKeys) > 0 {
			keys = commaKeys
		}
	}

	log("Parsed keys for '%s': %v", help, keys)

	binding := key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(strings.Join(keys, "/"), help),
	)

	log("Created key binding for '%s'", help)
	return binding
}

// Logger for debug messages
var (
	keys      = defaultKeyMap()
	isVerbose = false
	logFile   *os.File
)

// log prints debug messages to the log file if verbose mode is enabled
func log(text string, args ...interface{}) {
	if isVerbose && logFile != nil {
		fmt.Fprintf(logFile, text+"\n", args...)
	}
}

// initLogger initializes the logging system
func initLogger(verbose bool) {
	isVerbose = verbose

	if verbose {
		// Create log filename with current date
		now := time.Now()
		logFileName := fmt.Sprintf("/tmp/awp_%s.log", now.Format("2006-01-02"))

		var err error
		logFile, err = os.Create(logFileName)
		if err != nil {
			fmt.Printf("Error creating log file: %v\n", err)
			return
		}

		log("Verbose logging enabled")
	}
}

func main() {
	//
	log("=== Starting AWP ===")

	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Initialize logger
	initLogger(*verbose)

	// Close log file when application exits
	if logFile != nil {
		defer logFile.Close()
	}

	// Load configuration
	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize keys with configured keybindings
	log("Initializing key bindings from configuration")
	keys = configuredKeyMap(config)
	log("Key bindings initialized successfully")

	// Connect to database
	db, err := connectDB(config.Database)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Ensure database schema
	if err := ensureSchema(db); err != nil {
		fmt.Printf("Error creating schema: %v\n", err)
		os.Exit(1)
	}

	// Create and run the Bubble Tea program
	p := tea.NewProgram(initialModel(db), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func loadConfig(configPath string) (Config, error) {
	// Get user's home directory for storing the database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	// Default SQLite database in user's home directory
	defaultDbPath := filepath.Join(homeDir, ".config", "awp", "todo.db")
	configDir := filepath.Join(homeDir, ".config", "awp")
	defaultConfigPath := filepath.Join(configDir, "config.json")

	// Default configuration
	config := Config{
		Database: defaultDbPath,
		KeyMap: map[string]string{
			"ShowHelp":         "ctrl+b",
			"QuitApp":          "q",
			"ToggleStatus":     "space",
			"AddTask":          "a",
			"EditTask":         "e",
			"DeleteTask":       "d",
			"ToggleViewMode":   "ctrl+v",
			"ShowDoneTasks":    "ctrl+d",
			"ShowUndoneTasks":  "ctrl+u",
			"SearchTasks":      "ctrl+f",
			"PrevDay":          "ctrl+left",
			"NextDay":          "ctrl+right",
			"PrevDayWithTasks": "ctrl+shift+left",
			"NextDayWithTasks": "ctrl+shift+right",
			"JumpToToday":      "h",
		},
	}

	// If configPath is empty, use the default path
	if configPath == "" {
		configPath = defaultConfigPath
	}

	// Try to read the config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		// If the file doesn't exist, create it with default values
		if os.IsNotExist(err) {
			// Create the config directory if it doesn't exist
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return config, err
			}

			// Marshal the default config to JSON
			configData, err = json.MarshalIndent(config, "", "  ")
			if err != nil {
				return config, err
			}

			// Write the default config file
			if err := os.WriteFile(configPath, configData, 0644); err != nil {
				return config, err
			}
		} else {
			// Some other error occurred while reading the file
			return config, err
		}
	} else {
		// File exists, parse it
		if err := json.Unmarshal(configData, &config); err != nil {
			return config, err
		}
	}

	return config, nil
}

func connectDB(dbPath string) (*sql.DB, error) {
	log("Connecting to database at %s", dbPath)

	// Expand tilde to home directory if present
	if strings.HasPrefix(dbPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dbPath = homeDir + dbPath[1:]
		log("Expanded path to %s", dbPath)
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

func ensureSchema(db *sql.DB) error {
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

func initialModel(db *sql.DB) Model {
	// Create an empty column - the title will be empty to avoid showing a header
	columns := []table.Column{
		{Title: "", Width: 60},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	// Remove the header border and styling to make it invisible
	s.Header = s.Header.
		BorderStyle(lipgloss.HiddenBorder()). // Hidden border
		BorderBottom(false).                  // No border at bottom
		Bold(false).                          // Not bold
		Foreground(lipgloss.NoColor{})        // No color (transparent)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color(ColorSelectedText)).
		Background(lipgloss.Color(ColorSelectedBg)).
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
		table:        t,
		db:           db,
		showCommands: false,
		mode:         NormalMode,
		titleInput:   titleInput,
		descInput:    descInput,
		dueDateInput: dueDateInput,
		searchInput:  searchInput,
		activeInput:  0,
		viewMode:     TodayViewMode,  // Default view mode shows today's tasks
		taskFilter:   AllTasksFilter, // Default to showing all tasks (both done and undone)
		viewDate:     time.Now(),
		searchTerm:   "", // Initialize empty search term
	}

	// Load initial data
	m.loadTodaysTasks()

	return m
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
		task := TodoItem{
			Status:      false,
			DueDate:     parsedDueDate,
			Title:       title,
			Description: desc,
			Projects:    projects,
			Contexts:    contexts,
		}

		// Insert new task using the database function
		err := addTask(m.db, task)
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
			err := updateTask(m.db, *m.editingItem)
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

// Database operation functions
func loadTasks(db *sql.DB, whereClause string) ([]TodoItem, error) {
	query := `
		SELECT id, status, title, description, created, lastmodified, duedate, projects, contexts
		FROM todos
	`
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	query += " ORDER BY duedate DESC"

	log("Query: '%s'", query)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TodoItem

	for rows.Next() {
		var item TodoItem
		var dueDate sql.NullTime
		var projectsStr string
		var contextsStr string

		if err := rows.Scan(
			&item.ID,
			&item.Status,
			&item.Title,
			&item.Description,
			&item.Created,
			&item.LastModified,
			&dueDate,
			&projectsStr,
			&contextsStr,
		); err != nil {
			return nil, err
		}

		if dueDate.Valid {
			item.DueDate = dueDate.Time
		}

		// Parse projects from comma-separated string
		if projectsStr != "" {
			item.Projects = strings.Split(projectsStr, ",")
			for i, project := range item.Projects {
				item.Projects[i] = strings.TrimSpace(project)
			}
		} else {
			item.Projects = []string{}
		}

		// Parse contexts from comma-separated string
		if contextsStr != "" {
			item.Contexts = strings.Split(contextsStr, ",")
			for i, context := range item.Contexts {
				item.Contexts[i] = strings.TrimSpace(context)
			}
		} else {
			item.Contexts = []string{}
		}

		items = append(items, item)
	}

	return items, nil
}

func addTask(db *sql.DB, task TodoItem) error {
	_, err := db.Exec(
		`INSERT INTO todos (status, title, description, created, lastmodified, duedate, projects, contexts) 
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)`,
		task.Status,
		task.Title,
		task.Description,
		task.DueDate,
		strings.Join(task.Projects, ","),
		strings.Join(task.Contexts, ","),
	)
	return err
}

func updateTask(db *sql.DB, task TodoItem) error {
	_, err := db.Exec(
		`UPDATE todos SET status = ?, title = ?, description = ?, lastmodified = CURRENT_TIMESTAMP, duedate = ?, projects = ?, contexts = ? 
		 WHERE id = ?`,
		task.Status,
		task.Title,
		task.Description,
		task.DueDate,
		strings.Join(task.Projects, ","),
		strings.Join(task.Contexts, ","),
		task.ID,
	)
	return err
}

func updateTaskStatus(db *sql.DB, id int, status bool) error {
	_, err := db.Exec(
		"UPDATE todos SET status = ?, lastmodified = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	return err
}

func deleteTask(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	return err
}

// Model methods that use the database functions
func (m *Model) loadTasks() {
	var items []TodoItem
	var err error
	var whereClause string

	// First, set up the viewMode and taskFilter parts of the where clause
	switch m.viewMode {
	case AllViewMode:
		// In AllViewMode, initially no date filter
		whereClause = ""

		// Then, handle task filters within AllViewMode
		switch m.taskFilter {
		case AllTasksFilter:
			// No additional filter needed for all tasks
		case DoneTasksFilter:
			whereClause = "status = 1" // SQLite uses 1 for true
		case UndoneTasksFilter:
			whereClause = "status = 0" // SQLite uses 0 for false
		}

	case TodayViewMode:
		// Show tasks for specific date
		dateStr := m.viewDate.Format("2006-01-02")
		whereClause = fmt.Sprintf("date(duedate) = date('%s')", dateStr)

		// Then, handle task filters within TodayViewMode
		switch m.taskFilter {
		case AllTasksFilter:
			// No additional filter needed for all tasks
		case DoneTasksFilter:
			whereClause = whereClause + " AND status = 1"
		case UndoneTasksFilter:
			whereClause = whereClause + " AND status = 0"
		}
	}

	// Finally, add search term filter if one is set
	if m.searchTerm != "" {
		var searchClause string

		// Check if searching for project with +project syntax
		if strings.HasPrefix(m.searchTerm, "+") && len(m.searchTerm) > 1 {
			projectName := m.searchTerm[1:] // Remove the + prefix
			// Search in projects column or in description
			searchClause = fmt.Sprintf("(projects LIKE '%%%s%%' OR description LIKE '%%%s%%')",
				projectName, m.searchTerm)
		} else if strings.HasPrefix(m.searchTerm, "@") && len(m.searchTerm) > 1 {
			// Check if searching for context with @context syntax
			contextName := m.searchTerm[1:] // Remove the @ prefix
			// Search in contexts column or in description
			searchClause = fmt.Sprintf("(contexts LIKE '%%%s%%' OR description LIKE '%%%s%%')",
				contextName, m.searchTerm)
		} else {
			// Regular search in title or description
			searchClause = fmt.Sprintf("(title LIKE '%%%s%%' OR description LIKE '%%%s%%')",
				m.searchTerm, m.searchTerm)
		}

		if whereClause == "" {
			whereClause = searchClause
		} else {
			whereClause = whereClause + " AND " + searchClause
		}
	}

	// Load the tasks with the combined where clause
	items, err = loadTasks(m.db, whereClause)

	if err != nil {
		m.err = err
		return
	}

	m.items = items
	tableRows := []table.Row{}

	for _, item := range m.items {
		// Add to table rows
		status := "[ ]"
		if item.Status {
			status = "[x]"
		}

		// Use title if available, otherwise description
		displayText := item.Description
		if item.Title != "" {
			displayText = item.Title
		}

		// Highlight project and context tags in the display text
		highlightedText := highlightProjectsAndContexts(displayText)

		// Combined display with just status and highlighted text
		combinedText := fmt.Sprintf("%s %s", status, highlightedText)
		tableRows = append(tableRows, table.Row{combinedText})
	}

	// Set table rows and update table
	m.table.SetRows(tableRows)
}

// For backward compatibility
func (m *Model) loadTodaysTasks() {
	m.viewDate = time.Now()
	m.viewMode = TodayViewMode
	m.loadTasks()
}

// findPrevDayWithTasks finds the previous day that has tasks and updates viewDate
func (m *Model) findPrevDayWithTasks() {
	// Start from the day before current viewDate
	startDate := m.viewDate.AddDate(0, 0, -1)

	// Store original filter to restore it later
	originalFilter := m.taskFilter

	// Set filter to show all tasks to make sure we find any task
	m.taskFilter = AllTasksFilter

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
	m.taskFilter = AllTasksFilter

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

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case NormalMode:
			switch {
			case key.Matches(msg, keys.ShowHelp):
				m.mode = HelpViewMode

			case key.Matches(msg, keys.QuitApp):
				return m, tea.Quit

			case key.Matches(msg, keys.JumpToToday):
				m.loadTodaysTasks()

			case key.Matches(msg, keys.ToggleStatus):
				if len(m.items) > 0 && m.table.Cursor() < len(m.items) {
					// Update in database
					idx := m.table.Cursor()
					if idx < len(m.items) {
						m.items[idx].Status = !m.items[idx].Status
						// Update status using the database function
						err := updateTaskStatus(m.db, m.items[idx].ID, m.items[idx].Status)
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
							// text := selectedRow[0][4:] // Skip the status part "[ ] " or "[x] "
							text := m.items[idx].Title

							// Highlight project and context tags in the text
							highlightedText := highlightProjectsAndContexts(text)

							// Create the new row with highlighted text
							selectedRow[0] = fmt.Sprintf("%s %s", statusPrefix, highlightedText)
							rows := m.table.Rows()
							rows[m.table.Cursor()] = selectedRow
							m.table.SetRows(rows)
						}
					}
				}

			case key.Matches(msg, keys.AddTask):
				m.mode = AddMode
				m.resetInputs()

			case key.Matches(msg, keys.EditTask):
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

			case key.Matches(msg, keys.DeleteTask):
				if len(m.items) > 0 && m.table.Cursor() < len(m.items) {
					m.mode = DeleteConfirmMode
					m.editingItem = &m.items[m.table.Cursor()]
				}

			case key.Matches(msg, keys.ToggleViewMode):
				// Toggle between today's tasks and all tasks
				if m.viewMode == TodayViewMode {
					m.viewMode = AllViewMode
				} else {
					m.viewMode = TodayViewMode
				}
				m.loadTasks()

			case key.Matches(msg, keys.PrevDay):
				if m.viewMode == TodayViewMode {
					m.viewDate = m.viewDate.AddDate(0, 0, -1)
					m.loadTasks()
				}

			case key.Matches(msg, keys.NextDay):
				if m.viewMode == TodayViewMode {
					m.viewDate = m.viewDate.AddDate(0, 0, 1)
					m.loadTasks()
				}

			case key.Matches(msg, keys.PrevDayWithTasks):
				if m.viewMode == TodayViewMode {
					m.findPrevDayWithTasks()
				}

			case key.Matches(msg, keys.NextDayWithTasks):
				if m.viewMode == TodayViewMode {
					m.findNextDayWithTasks()
				}

			case key.Matches(msg, keys.ShowDoneTasks):
				// Toggle between done tasks and all tasks
				if m.taskFilter == DoneTasksFilter {
					m.taskFilter = AllTasksFilter
				} else {
					m.taskFilter = DoneTasksFilter
				}
				m.loadTasks()

			case key.Matches(msg, keys.ShowUndoneTasks):
				// Toggle between undone tasks and all tasks
				if m.taskFilter == UndoneTasksFilter {
					m.taskFilter = AllTasksFilter
				} else {
					m.taskFilter = UndoneTasksFilter
				}
				m.loadTasks()

			case key.Matches(msg, keys.SearchTasks):
				// Enter search mode
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
				log("Searching for: %s", m.searchTerm)
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
					log("Deleting task ID: %d", m.editingItem.ID)
					// Delete from database using the database function
					err := deleteTask(m.db, m.editingItem.ID)
					if err != nil {
						log("Error deleting task: %v", err)
						m.err = err
					} else {
						log("Task deleted successfully")
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

		// Adjust table height based on whether commands are shown
		if m.mode == NormalMode && !m.showCommands {
			m.table.SetHeight(msg.Height - 6) // More space for tasks when commands are hidden
		} else {
			m.table.SetHeight(msg.Height - 10) // Less space when showing commands or in other modes
		}
	}

	// Only update table in normal mode
	if m.mode == NormalMode {
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var sb strings.Builder

	switch m.mode {
	case NormalMode:
		baseStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ColorBorder))

		// Table with tasks
		sb.WriteString(baseStyle.Render(m.table.View()))
		sb.WriteString("\n")

		// Display view mode and date
		viewInfo := ""

		// Build the view mode part
		var viewModePart string
		switch m.viewMode {
		case AllViewMode:
			viewModePart = "all tasks"
		case TodayViewMode:
			viewModePart = fmt.Sprintf("tasks due on %s", m.viewDate.Format("2006-01-02"))
		}

		// Build the filter part
		var filterPart string
		switch m.taskFilter {
		case AllTasksFilter:
			filterPart = " (no filter)"
		case DoneTasksFilter:
			filterPart = " (completed only)"
		case UndoneTasksFilter:
			filterPart = " (pending only)"
		}

		// show search filter
		if m.searchTerm != "" {
			filterPart = fmt.Sprintf(" (search filter: %s)", m.searchTerm)
		}

		// Combine the parts
		viewInfo = fmt.Sprintf("Showing %s%s", viewModePart, filterPart)
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(ColorNormalText)).Render(viewInfo))
		sb.WriteString("\n")

	case AddMode:
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Add New Task"))
		sb.WriteString("\n\n")
		sb.WriteString(m.renderForm())

	case EditMode:
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Edit Task"))
		sb.WriteString("\n\n")
		sb.WriteString(m.renderForm())

	case DeleteConfirmMode:
		sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorError)).Render("Delete Task"))
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
			Foreground(lipgloss.Color(ColorAccent)).
			Bold(true)

		// Define a style for command descriptions
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorNormalText))

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
		addCommand(keys.QuitApp)
		addCommand(keys.ShowHelp)
		addCommand(keys.ToggleStatus)
		addCommand(keys.AddTask)
		addCommand(keys.EditTask)
		addCommand(keys.DeleteTask)
		addCommand(keys.ToggleViewMode)
		addCommand(keys.ShowDoneTasks)
		addCommand(keys.ShowUndoneTasks)
		addCommand(keys.SearchTasks)

		// Navigation commands
		sb.WriteString("\n")
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Navigation Commands"))
		sb.WriteString("\n\n")

		addCommand(keys.PrevDay)
		addCommand(keys.NextDay)
		addCommand(keys.PrevDayWithTasks)
		addCommand(keys.NextDayWithTasks)
	}

	// Error message if any
	if m.err != nil {
		sb.WriteString(fmt.Sprintf("\n\nError: %v", m.err))
	}

	return sb.String()
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
func highlightProjectsAndContexts(text string) string {
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
			result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(ColorProject)).Render(word))
		} else if strings.HasPrefix(word, "@") && len(word) > 1 {
			// Highlight context with a different color (blue)
			result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(ColorContext)).Render(word))
		} else {
			// Regular word, no highlighting
			result.WriteString(word)
		}
	}

	return result.String()
}

// renderForm renders the input form for adding/editing tasks
func (m Model) renderForm() string {
	var sb strings.Builder

	//formStyle := lipgloss.NewStyle().
	//	Border(lipgloss.RoundedBorder()).
	//	BorderForeground(lipgloss.Color(ColorBorder)).
	//	Padding(1, 2)

	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
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
