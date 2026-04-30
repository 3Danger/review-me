package preferences

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Preferences holds user preferences for the GUI application
type Preferences struct {
	Action      string    `json:"action"`
	Timezone    string    `json:"timezone"`
	Team        string    `json:"team"`
	LastUpdated time.Time `json:"last_updated"`
}

// Load reads preferences from the OS-specific config directory
// If the file doesn't exist, it returns default preferences
func Load() (*Preferences, error) {
	filePath, err := GetFilePath()
	if err != nil {
		return nil, fmt.Errorf("getting preferences file path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Return default preferences if file doesn't exist
		return getDefaults(), nil
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading preferences file: %w", err)
	}

	// Parse JSON
	var prefs Preferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return nil, fmt.Errorf("parsing preferences file: %w", err)
	}

	return &prefs, nil
}

// Save writes preferences to the OS-specific config directory
func (p *Preferences) Save() error {
	filePath, err := GetFilePath()
	if err != nil {
		return fmt.Errorf("getting preferences file path: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating preferences directory: %w", err)
	}

	// Update last updated timestamp
	p.LastUpdated = time.Now()

	// Marshal to JSON
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling preferences: %w", err)
	}

	// Write to file with restricted permissions
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("writing preferences file: %w", err)
	}

	return nil
}

// GetFilePath returns the OS-specific path for the preferences file
func GetFilePath() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\review-info\preferences.json
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		configDir = filepath.Join(appData, "review-info")

	case "darwin":
		// macOS: ~/Library/Application Support/review-info/preferences.json
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting user home directory: %w", err)
		}
		configDir = filepath.Join(home, "Library", "Application Support", "review-info")

	case "linux":
		// Linux: ~/.config/review-info/preferences.json
		configDir = os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("getting user home directory: %w", err)
			}
			configDir = filepath.Join(home, ".config")
		}
		configDir = filepath.Join(configDir, "review-info")

	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return filepath.Join(configDir, "preferences.json"), nil
}

// getDefaults returns default preferences
func getDefaults() *Preferences {
	return &Preferences{
		Action:      "review",
		Timezone:    "Europe/Moscow",
		Team:        "", // Will be populated from .env when GUI initializes
		LastUpdated: time.Now(),
	}
}

// LoadWithConfigTeam loads preferences and sets the team from config if preferences team is empty
func LoadWithConfigTeam(configTeam string) (*Preferences, error) {
	prefs, err := Load()
	if err != nil {
		return nil, err
	}

	// If team is empty in preferences, use the value from config
	if prefs.Team == "" {
		prefs.Team = configTeam
	}

	return prefs, nil
}
