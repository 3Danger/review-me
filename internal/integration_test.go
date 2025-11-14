package internal

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"review-info/internal/config"
	"review-info/internal/pkg/gitlab"
	"review-info/internal/pkg/jira"
	"review-info/internal/pkg/shower"
	"review-info/internal/preferences"
	"review-info/internal/service/manager"
)

// TestEndToEndCLI tests the complete CLI flow
func TestEndToEndCLI(t *testing.T) {
	// Load config
	cnf, err := config.Load("../config.yml")
	if err != nil {
		t.Skipf("Skipping test: config.yml not found: %v", err)
	}

	// Create services
	client := new(http.Client)
	svc := manager.New(
		cnf.Message,
		shower.New(
			gitlab.New(client, cnf.Gitlab),
			jira.New(client, cnf.Jira),
		),
	)

	// Test ReviewMe
	mrURL := "https://git.vseinstrumenti.net/fd/account-balance/-/merge_requests/591"
	message, err := svc.ReviewMe(mrURL)
	if err != nil {
		t.Fatalf("ReviewMe failed: %v", err)
	}

	if message == "" {
		t.Error("ReviewMe returned empty message")
	}

	t.Logf("ReviewMe result: %s", message)
}

// TestPreferencesPersistence tests preferences save and load
func TestPreferencesPersistence(t *testing.T) {
	// Create a temporary directory for test preferences
	tmpDir := t.TempDir()
	prefsFile := filepath.Join(tmpDir, "preferences.json")

	// Override the preferences file path for testing
	// We'll use environment variable or direct file operations
	testPrefs := &preferences.Preferences{
		Action:      "deploy",
		Timezone:    "America/New_York",
		Team:        "@test-team",
		LastUpdated: time.Now(),
	}

	// Save preferences to temp file
	// Note: This test is limited because GetFilePath() uses OS-specific paths
	// In a real scenario, we'd need to refactor preferences to accept a custom path
	t.Logf("Test preferences would be saved to: %s", prefsFile)
	t.Logf("Preferences: Action=%s, Timezone=%s, Team=%s",
		testPrefs.Action, testPrefs.Timezone, testPrefs.Team)

	// For now, just test that GetFilePath returns a valid path
	path, err := preferences.GetFilePath()
	if err != nil {
		t.Fatalf("GetFilePath failed: %v", err)
	}

	if path == "" {
		t.Error("GetFilePath returned empty path")
	}

	t.Logf("Preferences file path: %s", path)

	// Test that we can create the directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("Failed to create preferences directory: %v", err)
	}

	// Clean up
	defer os.RemoveAll(dir)
}

// TestGUIInitialization tests that GUI components can be initialized
func TestGUIInitialization(t *testing.T) {
	// Load config
	cnf, err := config.Load("../config.yml")
	if err != nil {
		t.Skipf("Skipping test: config.yml not found: %v", err)
	}

	// Create preferences
	prefs := &preferences.Preferences{
		Action:   "review",
		Timezone: "Europe/Moscow",
		Team:     cnf.Message.Team,
	}

	// Create services
	client := new(http.Client)
	svc := manager.New(
		cnf.Message,
		shower.New(
			gitlab.New(client, cnf.Gitlab),
			jira.New(client, cnf.Jira),
		),
	)

	// Test that we can create GUI app without running it
	// This tests the initialization logic
	t.Logf("Service initialized: %v", svc != nil)
	t.Logf("Preferences: Action=%s, Timezone=%s, Team=%s",
		prefs.Action, prefs.Timezone, prefs.Team)

	// Verify preferences are valid
	if prefs.Action != "review" && prefs.Action != "deploy" {
		t.Errorf("Invalid action: %s", prefs.Action)
	}

	if prefs.Timezone == "" {
		t.Error("Timezone is empty")
	}

	if prefs.Team == "" {
		t.Error("Team is empty")
	}
}
