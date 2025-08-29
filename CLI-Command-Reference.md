### CLI Command Reference

AWP provides a comprehensive command-line interface for task management. Below are all available CLI options and their usage:

### Configuration Options

#### `--config <path>`
Specify a custom path to the configuration file.
```bash
awp --config /path/to/custom/config.yaml
```

#### `--verbose`
Enable verbose logging for debugging and detailed output.
```bash
awp --verbose
```

### Task Management

#### `--add <task_description>`
Add a new task to the database. Supports project and context tags.
```bash
awp --add "Complete project documentation +work @urgent"
```

**Tag Support:**
- Projects: Use `+projectname` (e.g., `+work`, `+personal`)
- Contexts: Use `@contextname` (e.g., `@urgent`, `@home`, `@calls`)

#### `--date <YYYY-MM-DD>`
Specify a due date for the task when using `--add`. If not provided, defaults to today's date.
```bash
awp --add "Review code" --date 2024-01-15
```

### Database Operations

#### `--database purge`
Delete tasks from the database. Can be combined with filter flags for selective deletion.
```bash
awp --database purge
awp --database purge --project work --yes
```

**Database Filter Flags:**

#### `--project <project_name>`
Filter operations by project name.
```bash
awp --database purge --project work
```

#### `--done`
Filter to include only completed tasks.
```bash
awp --database purge --done
```

#### `--undone`
Filter to include only incomplete tasks.
```bash
awp --database purge --undone
```

#### `--yes`
Skip confirmation prompts (useful for automation).
```bash
awp --database purge --yes
```

### Import/Export Operations

#### `--import <filename>`
Import tasks from a file. Supports date parsing in DD.MM.YYYY format.
```bash
awp --import tasks.txt
```

**Import Format:**
```
01.01.2024:
- Task one +project @context
- Task two +work

02.01.2024:
- Another task +personal
```

#### `--export <filename>`
Export all tasks to a file. Use `--type` to specify the output format.
```bash
awp --export backup.json
awp --export tasks.txt --type txt
```

#### `--type <format>`
Specify export file format. Available options:
- `json` (default): JSON format with full task details
- `txt`: Plain text format with status and dates

```bash
awp --export tasks.json --type json
awp --export tasks.txt --type txt
```

### Usage Examples

**Basic task addition:**
```bash
awp --add "Buy groceries @shopping"
```

**Task with project and date:**
```bash
awp --add "Finish quarterly report +work @urgent" --date 2024-01-31
```

**Export tasks as JSON:**
```bash
awp --export backup/tasks.json --type json
```

**Import tasks from file:**
```bash
awp --import imported_tasks.txt
```

**Purge completed tasks from a specific project:**
```bash
awp --database purge --project work --done --yes
```

**Purge all tasks with confirmation:**
```bash
awp --database purge
```

### Interactive Mode

If no CLI commands are provided, AWP launches in interactive TUI (Text User Interface) mode for visual task management.

```bash
awp  # Launches interactive mode
```