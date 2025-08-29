# AWP - Todo Application

A simple CLI todo application written in Go that uses sqlite3 for task management.

Current day view:

![Screenshot](/media/screenshot.png)

All tasks view:

![Screenshot](/media/alltasksview.png)

Calendar view:

![Screenshot](/media/calendar_view.png)

## Features

- Simple, clean user interface using bubbletea and lipgloss
- Shows today's tasks by default, with option to view all tasks
- Date/Month/Calendar navigation to view tasks due on specific days
- Quick navigation with hotkeys (h to jump to today, ctrl+shift+arrow keys to navigate to days with tasks)
- Filtering capabilities to show only done or undone tasks
- Search functionality to find specific tasks
- Stores data in a SQLite database

## Todo Item Properties

- Status (boolean): Whether the task is completed or not
- Created/LastModified (datetime): When the task was created or last updated
- Title/Description (string): Task title and details
- Due (datetime): When the task is due to finish
- Context (string[]): Context for the task
- Project (string[]): Project for the task

## Installation

1. Make sure you have Go installed on your system
2. Clone this repository
3. Run `go mod download` to fetch dependencies
4. Build the application with `go build`

## Usage

```bash
# Run with default database config
./awp

# Specify a different config
./awp --config ~/.config/awp/config.json


# Run with verbose output
./awp --verbose

# Add a task
./awp --add "Buy milk"

# Add a task with due date
./awp --add "Buy milk" --due "2021-01-01"

# Add a task with context
./awp --add "Buy milk" --context "shopping"

# Add a task with project
./awp --add "Buy milk" --project "groceries"

# Add a task with due date and context
./awp --add "Buy milk" --due "2021-01-01" --context "shopping"

# Add a task with context in the description
./awp --add "Buy milk. @shopping"

```

## Configuration

The application can be configured in two ways:

1. Command-line flags:
   - `--verbose`: Debug output to /tmp/awp_%date%.log

2. Configuration file:
```
   - A `config.json` file located at `/home/users/.config/awp/config`
   - Example configuration:
     ```json
    {
    "database": "~/.config/awp/todo.db",
        "keymap": {
            "ShowHelp": "ctrl+b",
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
            "NextDay": "[\"ctrl+right\", \"right\", \"l\"]",
            "PrevDayWithTasks": "ctrl+shift+left",
            "NextDayWithTasks": "ctrl+shift+right",
            "JumpToToday": "h"
        }
    }
 ```

## Database

The application uses SQLite to store task data. The default database name is `todo.db`. 

The schema includes a `todos` table with the following columns:
- `id`: Serial primary key
- `status`: Boolean indicating completion status
- `title`: Text field for task title
- `description`: Text field for task details
- `created`: Timestamp of creation
- `lastmodified`: Timestamp of last update
- `due`: Due date
- `context`: Context tags for the task
- `project`: Project tags for the task

## Development

This project uses:
- [bubbletea](https://github.com/charmbracelet/bubbletea) for terminal UI
- [lipgloss](https://github.com/charmbracelet/lipgloss) for styling
- [go-sqlite3](github.com/mattn/go-sqlite3) for sqlite3 connectivity