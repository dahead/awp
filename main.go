package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"awp/pkg/cli"
	"awp/pkg/config"
	"awp/pkg/database"
	"awp/pkg/ui"
	"awp/pkg/utils"
)

func main() {
	utils.Log("=== Starting AWP 0.2 ===")

	// Parse command line arguments
	args := cli.ParseArgs()

	// Initialize logger
	utils.InitLogger(args.Verbose)
	defer utils.CloseLogger()

	// Load configuration and styles
	cfg, styles, err := config.Load(args.ConfigPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	db, err := database.ConnectDB(cfg.Database)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Ensure database schema
	if err := database.EnsureSchema(db); err != nil {
		fmt.Printf("Error creating schema: %v\n", err)
		os.Exit(1)
	}

	// Handle CLI commands
	if cli.HandleCommands(db, args) {
		return
	}

	// If no CLI commands, launch the TUI
	utils.Log("Configuration loaded successfully!")
	utils.Log("Database connection established and schema ensured")

	// Create the UI model
	model := ui.NewModel(db, cfg, styles)

	// Create and run the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
