package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"awp/pkg/config"
	"awp/pkg/database"
	"awp/pkg/ui"
	"awp/pkg/utils"
)

func main() {
	utils.Log("=== Starting AWP ===")

	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Initialize logger
	utils.InitLogger(*verbose)

	// Close log file when application exits
	defer utils.CloseLogger()

	// Load configuration and styles
	cfg, styles, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	utils.Log("Configuration loaded successfully!")

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
