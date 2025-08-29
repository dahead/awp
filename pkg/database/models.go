package database

import (
	"time"
)

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

// ViewMode represents the current view mode for tasks
type ViewMode int

const (
	TodayViewMode ViewMode = iota // Default - show tasks for today
	AllViewMode                   // Show all tasks (no date filter)
	CalendarViewMode
)

// TaskFilter represents the current task filter mode
type TaskFilter int

const (
	AllTasksFilter    TaskFilter = iota // Show all tasks regardless of status
	DoneTasksFilter                     // Show only completed tasks
	UndoneTasksFilter                   // Show only uncompleted tasks
)
