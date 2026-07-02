package internal

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"review-info/internal/config"
	"review-info/internal/domain"
	"review-info/internal/pkg/gitlab"
	"review-info/internal/pkg/jira"
	"review-info/internal/pkg/shower"
	"review-info/internal/preferences"
	"review-info/internal/service/manager"
)

// TestEndToEndCLI tests the complete CLI flow with real API endpoints.
// Skipped unless REVIEW_INFO_INTEGRATION=1 is set.
func TestEndToEndCLI(t *testing.T) {
	if os.Getenv("REVIEW_INFO_INTEGRATION") != "1" {
		t.Skip("Skipping: set REVIEW_INFO_INTEGRATION=1 to run integration tests")
	}

	cnf, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() failed: %v", err)
	}

	if cnf.Gitlab.Token == "" || cnf.Jira.Pass == "" {
		t.Fatal("GITLAB_TOKEN and JIRA_PASSWORD must be set in .env")
	}

	client := &http.Client{Timeout: 30 * time.Second}

	svc := manager.New(
		cnf.Message,
		shower.New(
			gitlab.New(client, cnf.Gitlab),
			jira.New(client, cnf.Jira),
			cnf.Gitlab.BaseURL,
			cnf.Jira.ProjectPrefix,
		),
	)

	mrURL := "https://git.vseinstrumenti.net/fd/account-balance/-/merge_requests/591"
	message, err := svc.Execute(context.Background(), domain.ActionReview, mrURL, domain.ActionOptions{})
	if err != nil {
		t.Fatalf("Execute review failed: %v", err)
	}

	if message == "" {
		t.Error("Execute returned empty message")
	}

	t.Logf("Result: %s", message)
}

// TestPreferencesPersistence tests saving and loading preferences to a file.
func TestPreferencesPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	prefsFile := filepath.Join(tmpDir, "preferences.json")

	// Save preferences
	origPrefs := &preferences.Preferences{
		Action:      "deploy",
		Timezone:    "America/New_York",
		Team:        "@test-team",
		LastUpdated: time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC),
	}

	if err := preferences.SaveToFile(origPrefs, prefsFile); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Load them back
	loadedPrefs, err := preferences.LoadFromFile(prefsFile)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if loadedPrefs.Action != origPrefs.Action {
		t.Errorf("Action = %q, want %q", loadedPrefs.Action, origPrefs.Action)
	}
	if loadedPrefs.Timezone != origPrefs.Timezone {
		t.Errorf("Timezone = %q, want %q", loadedPrefs.Timezone, origPrefs.Timezone)
	}
	if loadedPrefs.Team != origPrefs.Team {
		t.Errorf("Team = %q, want %q", loadedPrefs.Team, origPrefs.Team)
	}

	// LastUpdated should be set to current time on load, not restored
	if loadedPrefs.LastUpdated.IsZero() {
		t.Error("LastUpdated should be set on load")
	}
}

// TestEndToEndCLI_SkipWithoutConfig verifies that config.Load skips cleanly.
func TestEndToEndCLI_SkipWithoutConfig(t *testing.T) {
	if _, err := config.Load(); err != nil {
		t.Logf(".env not found or incomplete: %v", err)
	}

	// Verify domain action constants are valid
	if domain.ActionReview != "review" {
		t.Errorf("ActionReview = %q, want %q", domain.ActionReview, "review")
	}
	if domain.ActionDeploy != "deploy" {
		t.Errorf("ActionDeploy = %q, want %q", domain.ActionDeploy, "deploy")
	}
}
