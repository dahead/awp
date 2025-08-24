# AWP - Todo Application

A simple CLI todo application written in Go that uses sqlite3 for task management.

![Screenshot](/media/screenshot.png)

## Features

- Simple, clean user interface using bubbletea and lipgloss
- Shows today's tasks by default, with option to view all tasks
- Date navigation to view tasks due on specific days
- Filtering capabilities to show only done or undone tasks
- Search functionality to find specific tasks
- Stores data in a SQLite database
- Central hotkey (CTRL+B) to show additional commands
- Full CRUD operations for managing tasks

## Todo Item Properties

- Status (boolean): Whether the task is completed or not
- Created/LastModified (datetime): When the task was created or last updated
- Title/Description (string): Task title and details
- Tags (string[]): List of tags associated with the task
- Due (datetime): When the task is due to finish

## Installation

1. Make sure you have Go installed on your system
2. Clone this repository
3. Run `go mod download` to fetch dependencies
4. Build the application with `go build`

## Usage

```bash
# Run with default database (todo.db)
./awp

# Specify a different database
./awp --database my_todos.db
```

## Configuration

The application can be configured in two ways:

1. Command-line flags:
   - `--database`: Path to the PostgreSQL database

2. Configuration file:
```
   - A `config.json` file located at `/home/users/.config/awp/config`
   - Example configuration:
     ```json
    {
    "database": "/tmp/test_todo.db",
    "central_hotkey": "ctrl+b",
        "keymap": {
        "ToggleShowCommands": "ctrl+b",
        "Quit": "[\"q\", \"ctrl+c\"]",
        "ToggleStatus": "t",
        "AddTask": "[\"a\", \"insert\"]",
        "EditTask": "[\"e\", \"enter\"]",
        "DeleteTask": "[\"d\", \"delete\"]",
        "ToggleViewMode": "ctrl+v",
        "ShowDoneTasks": "ctrl+d",
        "ShowUndoneTasks": "ctrl+u",
        "SearchTasks": "ctrl+f",
        "PrevDay": "[\"ctrl+left\", \"left\", \"j\"]",
        "NextDay": "[\"ctrl+right\", \"right\", \"l\"]"
        }
    }
 ```

## Database

The application uses SQLite to store task data. The default database name is `todo.db`. 

The schema includes a `todos` table with the following columns:
- `id`: Serial primary key
- `status`: Boolean indicating completion status
- `created`: Timestamp of creation
- `lastmodified`: Timestamp of last update
- `title`: Text field for task title
- `description`: Text field for task details
- `tags`: Text array for task tags
- `due`: Due date

## User Interface

The application has a simple, clean interface:

- Tasks are displayed in a table showing status and description
- Additional commands are hidden by default
- Press CTRL+B to toggle the visibility of command help
- Commands:
  - q: Quit the application
  - t: Toggle task status
  - a: Add a new task
  - e: Edit the selected task
  - d: Delete the selected task
  - ctrl+v: Toggle between today's tasks and all tasks
  - ctrl+d: Show only done tasks
  - ctrl+u: Show only undone tasks
  - ctrl+f: Search tasks
  - ctrl+←: View tasks due on the previous day
  - ctrl+→: View tasks due on the next day

## Development

This project uses:
- [bubbletea](https://github.com/charmbracelet/bubbletea) for terminal UI
- [lipgloss](https://github.com/charmbracelet/lipgloss) for styling
- [go-sqlite3](github.com/mattn/go-sqlite3) for sqlite3 connectivity
- [viper](https://github.com/spf13/viper) for configuration management