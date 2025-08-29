package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"awp/pkg/keymaps"
)

// Config holds the application configuration
type Config struct {
	Database   string            `json:"database"`
	KeyMap     map[string]string `json:"keymap"`
	StylesFile string            `json:"styles_file"`
}

// Styles holds the application colors and styling information
type Styles struct {
	// UI element colors
	BorderColor string `json:"border_color"`
	AccentColor string `json:"accent_color"`

	// Text colors
	NormalTextColor   string `json:"normal_text_color"`
	SelectedTextColor string `json:"selected_text_color"`
	SelectedBgColor   string `json:"selected_bg_color"`
	ErrorColor        string `json:"error_color"`

	// Project and context colors
	ProjectColor string `json:"project_color"`
	ContextColor string `json:"context_color"`
}

// Load loads the application configuration from the specified path
func Load(configPath string) (Config, Styles, error) {
	// Get user's home directory for storing the database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, Styles{}, err
	}

	// Default SQLite database in user's home directory
	defaultDbPath := filepath.Join(homeDir, ".config", "awp", "todo.db")
	configDir := filepath.Join(homeDir, ".config", "awp")
	defaultConfigPath := filepath.Join(configDir, "config.json")

	// Default configuration using keymaps package
	config := Config{
		Database:   defaultDbPath,
		KeyMap:     keymaps.GetDefaultKeyMappings(),
		StylesFile: filepath.Join(configDir, "styles.json"),
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
				return config, Styles{}, err
			}

			// Marshal the default config to JSON
			configData, err = json.MarshalIndent(config, "", "  ")
			if err != nil {
				return config, Styles{}, err
			}

			// Write the default config file
			if err := os.WriteFile(configPath, configData, 0644); err != nil {
				return config, Styles{}, err
			}
		} else {
			// Some other error occurred while reading the file
			return config, Styles{}, err
		}
	} else {
		// File exists, parse it
		if err := json.Unmarshal(configData, &config); err != nil {
			return config, Styles{}, err
		}
	}

	// Now load the styles file
	styles, err := loadStyles(config.StylesFile)
	if err != nil {
		return config, styles, fmt.Errorf("error loading styles: %w", err)
	}

	return config, styles, nil
}

// loadStyles loads the application styles from the specified path
func loadStyles(stylesPath string) (Styles, error) {
	// Default styles that match the current constants
	defaultStyles := Styles{
		BorderColor:       "240",
		AccentColor:       "205",
		NormalTextColor:   "86",
		SelectedTextColor: "229",
		SelectedBgColor:   "57",
		ErrorColor:        "9",
		ProjectColor:      "2",
		ContextColor:      "4",
	}

	// Try to read the styles file
	stylesData, err := os.ReadFile(stylesPath)
	if err != nil {
		// If the file doesn't exist, create it with default values
		if os.IsNotExist(err) {
			// Create the directory if it doesn't exist
			stylesDir := filepath.Dir(stylesPath)
			if err := os.MkdirAll(stylesDir, 0755); err != nil {
				return defaultStyles, err
			}

			// Marshal the default styles to JSON
			stylesData, err = json.MarshalIndent(defaultStyles, "", "  ")
			if err != nil {
				return defaultStyles, err
			}

			// Write the default styles file
			if err := os.WriteFile(stylesPath, stylesData, 0644); err != nil {
				return defaultStyles, err
			}

			return defaultStyles, nil
		} else {
			// Some other error occurred
			return defaultStyles, err
		}
	}

	// File exists, parse it
	var loadedStyles Styles
	if err := json.Unmarshal(stylesData, &loadedStyles); err != nil {
		return defaultStyles, err
	}

	return loadedStyles, nil
}
