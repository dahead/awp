package database

import (
	"awp/pkg/utils"
	"database/sql"
	"fmt"
	"strings"
)

// LoadTasks retrieves tasks from the database based on the where clause
func LoadTasks(db *sql.DB, whereClause string) ([]TodoItem, error) {
	query := `
		SELECT id, status, title, description, created, lastmodified, duedate, projects, contexts
		FROM todos
	`
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	query += " ORDER BY duedate DESC"

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

	utils.Log("Loaded %d tasks from database", len(items))

	return items, nil
}

// AddTask inserts a new task into the database
func AddTask(db *sql.DB, task TodoItem) error {
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
	utils.Log("Added task: %s", task.ID)
	return err
}

// UpdateTask updates an existing task in the database
func UpdateTask(db *sql.DB, task TodoItem) error {
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
	utils.Log("Updated task: %s", task.ID)
	return err
}

// UpdateTaskStatus updates only the status of a task
func UpdateTaskStatus(db *sql.DB, id int, status bool) error {
	_, err := db.Exec(
		"UPDATE todos SET status = ?, lastmodified = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	return err
}

// DeleteTask removes a task from the database
func DeleteTask(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	return err
}

// BuildWhereClause builds a SQL where clause based on view mode, task filter, and search term
func BuildWhereClause(viewMode ViewMode, taskFilter TaskFilter, viewDate string, searchTerm string) string {
	var whereClause string

	// First, set up the viewMode and taskFilter parts of the where clause
	switch viewMode {
	case AllViewMode:
		// In AllViewMode, initially no date filter
		whereClause = ""

		// Then, handle task filters within AllViewMode
		switch taskFilter {
		case AllTasksFilter:
			// No additional filter needed for all tasks
		case DoneTasksFilter:
			whereClause = "status = 1" // SQLite uses 1 for true
		case UndoneTasksFilter:
			whereClause = "status = 0" // SQLite uses 0 for false
		}

	case TodayViewMode:
		// Show tasks for specific date
		whereClause = fmt.Sprintf("date(duedate) = date('%s')", viewDate)

		// Then, handle task filters within TodayViewMode
		switch taskFilter {
		case AllTasksFilter:
			// No additional filter needed for all tasks
		case DoneTasksFilter:
			whereClause = whereClause + " AND status = 1"
		case UndoneTasksFilter:
			whereClause = whereClause + " AND status = 0"
		}
	}

	// Finally, add search term filter if one is set
	if searchTerm != "" {
		var searchClause string

		// Check if searching for project with +project syntax
		if strings.HasPrefix(searchTerm, "+") && len(searchTerm) > 1 {
			projectName := searchTerm[1:] // Remove the + prefix
			// Search in projects column or in description
			searchClause = fmt.Sprintf("(projects LIKE '%%%s%%' OR description LIKE '%%%s%%')",
				projectName, searchTerm)
		} else if strings.HasPrefix(searchTerm, "@") && len(searchTerm) > 1 {
			// Check if searching for context with @context syntax
			contextName := searchTerm[1:] // Remove the @ prefix
			// Search in contexts column or in description
			searchClause = fmt.Sprintf("(contexts LIKE '%%%s%%' OR description LIKE '%%%s%%')",
				contextName, searchTerm)
		} else {
			// Regular search in title or description
			searchClause = fmt.Sprintf("(title LIKE '%%%s%%' OR description LIKE '%%%s%%')",
				searchTerm, searchTerm)
		}

		if whereClause == "" {
			whereClause = searchClause
		} else {
			whereClause = whereClause + " AND " + searchClause
		}
	}

	utils.Log("Built where clause: %s", whereClause)

	return whereClause
}
