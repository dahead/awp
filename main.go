package main

import (
	"database/sql"
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
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Database      string `mapstructure:"database"`
	CentralHotkey string `mapstructure:"central_hotkey"`
}

// TodoItem represents a single todo task
type TodoItem struct {
	ID           int       `db:"id"`
	Status       bool      `db:"status"`
	Created      time.Time `db:"created"`
	LastModified time.Time `db:"lastmodified"`
	DueDate      time.Time `db:"duedate"`
	Title        string    `db:"title"`
	Description  string    `db:"description"`
	Tags         []string  `db:"tags"`
}

// InputMode represents the current input mode
type InputMode int

const (
	NormalMode InputMode = iota
	AddMode
	EditMode
	DeleteConfirmMode
)

// Model represents the application state
type Model struct {
	table         table.Model
	items         []TodoItem
	db            *sql.DB
	showCommands  bool
	width, height int
	err           error

	// Form state
	mode         InputMode
	titleInput   textinput.Model
	descInput    textinput.Model
	tagsInput    textinput.Model
	dueDateInput textinput.Model
	activeInput  int

	// Edit/delete state
	editingItem *TodoItem
}

var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			PaddingLeft(1)

	statusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Background(lipgloss.Color("237")).
			Padding(0, 1)
)

// Define keymaps
type keyMap struct {
	ToggleShowCommands key.Binding
	Quit               key.Binding
	ToggleStatus       key.Binding
	AddTask            key.Binding
	EditTask           key.Binding
	DeleteTask         key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		ToggleShowCommands: key.NewBinding(
			key.WithKeys("ctrl+b"),
			key.WithHelp("ctrl+b", "show/hide commands"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		ToggleStatus: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle status"),
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
	}
}

var keys = defaultKeyMap()

func main() {
	// Parse command line flags
	dbPath := flag.String("database", "", "Path to database file")
	flag.Parse()

	// Load configuration
	config, err := loadConfig(*dbPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

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

func loadConfig(dbPath string) (Config, error) {
	// Get user's home directory for storing the database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	// Default SQLite database in user's home directory
	defaultDbPath := filepath.Join(homeDir, ".config", "awp", "todo.db")

	config := Config{
		Database:      defaultDbPath,
		CentralHotkey: "ctrl+b",
	}

	// Setup viper
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "awp"))

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, err
		}
		// Config file not found, create default config
		if err := os.MkdirAll(filepath.Join(os.Getenv("HOME"), ".config", "awp"), 0755); err != nil {
			return config, err
		}
		viper.Set("database", config.Database)
		viper.Set("central_hotkey", config.CentralHotkey)
		if err := viper.WriteConfigAs(filepath.Join(os.Getenv("HOME"), ".config", "awp", "config.json")); err != nil {
			return config, err
		}
	}

	// Override with command-line flag if provided
	if dbPath != "" {
		config.Database = dbPath
	} else {
		if viper.IsSet("database") {
			config.Database = viper.GetString("database")
		}
	}

	if viper.IsSet("central_hotkey") {
		config.CentralHotkey = viper.GetString("central_hotkey")
	}

	return config, nil
}

func connectDB(dbPath string) (*sql.DB, error) {
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
			tags TEXT
		)
	`)
	return err
}

func initialModel(db *sql.DB) Model {
	columns := []table.Column{
		{Title: "Status", Width: 10},
		{Title: "Description", Width: 50},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s)

	// Initialize text inputs
	titleInput := textinput.New()
	titleInput.Placeholder = "Title"
	titleInput.Focus()
	titleInput.Width = 40

	descInput := textinput.New()
	descInput.Placeholder = "Description"
	descInput.Width = 40

	tagsInput := textinput.New()
	tagsInput.Placeholder = "Tags (comma separated)"
	tagsInput.Width = 40

	// Initialize due date input with today's date as default
	dueDateInput := textinput.New()
	dueDateInput.Placeholder = "Due Date (YYYY-MM-DD, optional)"
	dueDateInput.Width = 40
	dueDateInput.SetValue(time.Now().Format("2006-01-02"))

	m := Model{
		table:        t,
		db:           db,
		showCommands: false,
		mode:         NormalMode,
		titleInput:   titleInput,
		descInput:    descInput,
		tagsInput:    tagsInput,
		dueDateInput: dueDateInput,
		activeInput:  0,
	}

	// Load initial data
	m.loadTodaysTasks()

	return m
}

// resetInputs clears all form inputs
func (m *Model) resetInputs() {
	m.titleInput.Reset()
	m.descInput.Reset()
	m.tagsInput.Reset()
	// Set due date input to today's date
	m.dueDateInput.SetValue(time.Now().Format("2006-01-02"))

	m.activeInput = 0
	m.titleInput.Focus()
	m.descInput.Blur()
	m.tagsInput.Blur()
	m.dueDateInput.Blur()
}

// focusNextInput cycles through the form inputs
func (m *Model) focusNextInput() {
	m.activeInput = (m.activeInput + 1) % 4

	switch m.activeInput {
	case 0:
		m.titleInput.Focus()
		m.descInput.Blur()
		m.tagsInput.Blur()
		m.dueDateInput.Blur()
	case 1:
		m.titleInput.Blur()
		m.descInput.Focus()
		m.tagsInput.Blur()
		m.dueDateInput.Blur()
	case 2:
		m.titleInput.Blur()
		m.descInput.Blur()
		m.tagsInput.Focus()
		m.dueDateInput.Blur()
	case 3:
		m.titleInput.Blur()
		m.descInput.Blur()
		m.tagsInput.Blur()
		m.dueDateInput.Focus()
	}
}

// submitForm processes the form data based on the current mode
func (m *Model) submitForm() {
	title := strings.TrimSpace(m.titleInput.Value())
	desc := strings.TrimSpace(m.descInput.Value())
	tags := strings.Split(strings.TrimSpace(m.tagsInput.Value()), ",")
	dueDate := strings.TrimSpace(m.dueDateInput.Value())

	// Trim spaces from tags
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	// Remove empty tags
	cleanedTags := []string{}
	for _, tag := range tags {
		if tag != "" {
			cleanedTags = append(cleanedTags, tag)
		}
	}

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
		// Default to today if no date provided
		parsedDueDate = time.Now()
	}

	switch m.mode {
	case AddMode:
		// Create new task with the collected data
		task := TodoItem{
			Status:      false,
			DueDate:     parsedDueDate,
			Title:       title,
			Description: desc,
			Tags:        cleanedTags,
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
			m.editingItem.Tags = cleanedTags
			m.editingItem.DueDate = parsedDueDate

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
		SELECT id, status, created, lastmodified, duedate, title, description, tags
		FROM todos
	`
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	query += " ORDER BY created DESC"

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TodoItem

	for rows.Next() {
		var item TodoItem
		var tagsStr string
		var dueDate sql.NullTime

		if err := rows.Scan(
			&item.ID,
			&item.Status,
			&item.Created,
			&item.LastModified,
			&dueDate,
			&item.Title,
			&item.Description,
			&tagsStr,
		); err != nil {
			return nil, err
		}

		if dueDate.Valid {
			item.DueDate = dueDate.Time
		}

		// Parse tags from comma-separated string
		if tagsStr != "" {
			// Split by comma
			item.Tags = strings.Split(tagsStr, ",")

			// Trim any whitespace from tags
			for i, tag := range item.Tags {
				item.Tags[i] = strings.TrimSpace(tag)
			}
		} else {
			item.Tags = []string{}
		}

		items = append(items, item)
	}

	return items, nil
}

func addTask(db *sql.DB, task TodoItem) error {
	_, err := db.Exec(
		`INSERT INTO todos (status, duedate, title, description, tags) 
		 VALUES (?, ?, ?, ?, ?)`,
		task.Status, task.DueDate, task.Title, task.Description, strings.Join(task.Tags, ","),
	)
	return err
}

func updateTask(db *sql.DB, task TodoItem) error {
	_, err := db.Exec(
		`UPDATE todos SET title = ?, description = ?, tags = ?, duedate = ?, status = ?, lastmodified = CURRENT_TIMESTAMP 
		 WHERE id = ?`,
		task.Title, task.Description, strings.Join(task.Tags, ","), task.DueDate, task.Status, task.ID,
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
func (m *Model) loadTodaysTasks() {
	// Filter tasks where due date is today
	items, err := loadTasks(m.db, "date(duedate) = date('now')")
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

		tableRows = append(tableRows, table.Row{status, displayText})
	}

	m.table.SetRows(tableRows)
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
			// Handle normal mode key presses
			switch {
			case key.Matches(msg, keys.ToggleShowCommands):
				m.showCommands = !m.showCommands

			case key.Matches(msg, keys.Quit):
				return m, tea.Quit

			case key.Matches(msg, keys.ToggleStatus):
				if len(m.items) > 0 && m.table.Cursor() < len(m.items) {
					selectedRow := m.table.SelectedRow()
					if selectedRow[0] == "[ ]" {
						selectedRow[0] = "[x]"
					} else {
						selectedRow[0] = "[ ]"
					}
					rows := m.table.Rows()
					rows[m.table.Cursor()] = selectedRow
					m.table.SetRows(rows)

					// Update in database
					idx := m.table.Cursor()
					if idx < len(m.items) {
						m.items[idx].Status = !m.items[idx].Status
						// Update status using the database function
						err := updateTaskStatus(m.db, m.items[idx].ID, m.items[idx].Status)
						if err != nil {
							m.err = err
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
					m.tagsInput.SetValue(strings.Join(m.editingItem.Tags, ", "))

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
			}

		case AddMode, EditMode:
			// Handle input mode key presses
			switch msg.String() {
			case "esc":
				m.mode = NormalMode
				m.resetInputs()
				m.editingItem = nil

			case "tab", "shift+tab":
				m.focusNextInput()

			case "enter":
				if m.activeInput == 3 { // Submit on enter from the last field (due date)
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
				m.tagsInput, cmd = m.tagsInput.Update(msg)
				cmds = append(cmds, cmd)
			case 3:
				m.dueDateInput, cmd = m.dueDateInput.Update(msg)
				cmds = append(cmds, cmd)
			}

		case DeleteConfirmMode:
			// Handle delete confirmation
			switch msg.String() {
			case "y", "Y":
				if m.editingItem != nil {
					// Delete from database using the database function
					err := deleteTask(m.db, m.editingItem.ID)
					if err != nil {
						m.err = err
					} else {
						m.loadTodaysTasks()
					}
				}
				m.mode = NormalMode
				m.editingItem = nil

			case "n", "N", "esc":
				m.mode = NormalMode
				m.editingItem = nil
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
		// Table with tasks
		sb.WriteString(baseStyle.Render(m.table.View()))
		sb.WriteString("\n\n")

		// Status bar / commands (only shown when showCommands is true)
		if m.showCommands {
			commands := []string{
				"q: quit",
				"t: toggle status",
				"a: add task",
				"e: edit task",
				"d: delete task",
			}
			sb.WriteString(statusBar.Render(strings.Join(commands, " | ")))
		}

	case AddMode:
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Add New Task"))
		sb.WriteString("\n\n")
		sb.WriteString(m.renderForm())
		sb.WriteString("\n\n")
		sb.WriteString(statusBar.Render("Tab: next field • Enter: submit • Esc: cancel"))

	case EditMode:
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Edit Task"))
		sb.WriteString("\n\n")
		sb.WriteString(m.renderForm())
		sb.WriteString("\n\n")
		sb.WriteString(statusBar.Render("Tab: next field • Enter: submit • Esc: cancel"))

	case DeleteConfirmMode:
		sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9")).Render("Delete Task"))
		sb.WriteString("\n\n")

		if m.editingItem != nil {
			sb.WriteString(fmt.Sprintf("Are you sure you want to delete this task?\n\n"))
			sb.WriteString(fmt.Sprintf("Title: %s\n", m.editingItem.Title))
			sb.WriteString(fmt.Sprintf("Description: %s\n", m.editingItem.Description))
			sb.WriteString("\n")
			sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Press Y to confirm, N to cancel"))
		}
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
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2)

	// Title input
	sb.WriteString("Title:\n")
	sb.WriteString(m.titleInput.View())
	sb.WriteString("\n\n")

	// Description input
	sb.WriteString("Description:\n")
	sb.WriteString(m.descInput.View())
	sb.WriteString("\n\n")

	// Tags input
	sb.WriteString("Tags (comma separated):\n")
	sb.WriteString(m.tagsInput.View())
	sb.WriteString("\n\n")

	// Due date input
	sb.WriteString("Due Date (YYYY-MM-DD):\n")
	sb.WriteString(m.dueDateInput.View())

	return formStyle.Render(sb.String())
}
