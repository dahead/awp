package cli

import (
	"database/sql"
	"flag"

	"awp/pkg/commands"
)

// Args represents parsed command line arguments
type Args struct {
	ConfigPath string
	Verbose    bool

	// Task operations
	AddTask  string
	DateFlag string

	// Database operations
	DatabaseCmd string
	ProjectFlag string
	YesFlag     bool
	DoneFlag    bool
	UndoneFlag  bool

	// Import/Export operations
	ImportFile string
	ExportFile string
	TypeFlag   string
}

// ParseArgs parses command line arguments and returns Args struct
func ParseArgs() *Args {
	args := &Args{}

	// Define command line flags
	flag.StringVar(&args.ConfigPath, "config", "", "Path to configuration file")
	flag.BoolVar(&args.Verbose, "verbose", false, "Enable verbose logging")

	// Task operations
	flag.StringVar(&args.AddTask, "add", "", "Add a new task")
	flag.StringVar(&args.DateFlag, "date", "", "Date for task (YYYY-MM-DD format)")

	// Database operations
	flag.StringVar(&args.DatabaseCmd, "database", "", "Database command (purge)")
	flag.StringVar(&args.ProjectFlag, "project", "", "Filter by project")
	flag.BoolVar(&args.YesFlag, "yes", false, "Skip confirmation")
	flag.BoolVar(&args.DoneFlag, "done", false, "Filter done tasks")
	flag.BoolVar(&args.UndoneFlag, "undone", false, "Filter undone tasks")

	// Import/Export operations
	flag.StringVar(&args.ImportFile, "import", "", "Import tasks from file")
	flag.StringVar(&args.ExportFile, "export", "", "Export tasks to file")
	flag.StringVar(&args.TypeFlag, "type", "json", "Export file type (json, txt)")

	flag.Parse()
	return args
}

// HandleCommands processes CLI commands and returns true if a command was handled
func HandleCommands(db *sql.DB, args *Args) bool {
	// Check for CLI commands
	if args.AddTask != "" {
		commands.HandleAddTask(db, args.AddTask, args.DateFlag)
		return true
	}

	if args.DatabaseCmd != "" {
		commands.HandleDatabaseCommand(db, args.DatabaseCmd, args.DateFlag, args.ProjectFlag, args.YesFlag, args.DoneFlag, args.UndoneFlag)
		return true
	}

	if args.ImportFile != "" {
		commands.HandleImportCommand(db, args.ImportFile)
		return true
	}

	if args.ExportFile != "" {
		commands.HandleExportCommand(db, args.ExportFile, args.TypeFlag)
		return true
	}

	// No CLI command was handled
	return false
}
