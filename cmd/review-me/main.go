package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"review-info/internal/config"
	"review-info/internal/gui"
	"review-info/internal/pkg/gitlab"
	"review-info/internal/pkg/jira"
	"review-info/internal/pkg/shower"
	"review-info/internal/domain"
	"review-info/internal/preferences"
	"review-info/internal/service/manager"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	// Detect mode: GUI or CLI
	if shouldRunGUI() {
		if err := runGUI(); err != nil {
			panic(err)
		}
	} else {
		if err := runCLI(); err != nil {
			panic(err)
		}
	}
}

// shouldRunGUI determines if the application should run in GUI mode
// Returns true if no CLI flags are present or if -gui flag is set
func shouldRunGUI() bool {
	hasCliFlags := false
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-cli":
			return false
		case "-gui":
			return true
		case "-mr", "-team", "-action":
			hasCliFlags = true
		}
	}
	return !hasCliFlags
}

// runGUI starts the application in GUI mode
func runGUI() error {
	slog.Info("Starting GUI mode...")

	// Load config
	slog.Info("Loading config...")
	cnf, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	slog.Info("Config loaded successfully")

	// Load preferences
	slog.Info("Loading preferences...")
	prefs, err := preferences.LoadWithConfigTeam(cnf.Message.Team)
	if err != nil {
		// Log error but continue with defaults
		slog.Warn("failed to load preferences, using defaults", "error", err)
		prefs = &preferences.Preferences{
			Action:   "review",
			Timezone: "Europe/Moscow",
			Team:     cnf.Message.Team,
		}
	}
	slog.Info("Preferences loaded", "team", prefs.Team, "action", prefs.Action)

	// Create HTTP client and services
	slog.Info("Creating services...")
	client := &http.Client{Timeout: 30 * time.Second}
	gitlabHost := extractHost(cnf.Gitlab.BaseURL)
	svc := manager.New(
		cnf.Message,
		shower.New(
			gitlab.New(client, cnf.Gitlab),
			jira.New(client, cnf.Jira),
			gitlabHost,
			cnf.Jira.ProjectPrefix,
		),
	)
	slog.Info("Services created successfully")

	// Create and run GUI application
	slog.Info("Creating GUI application...")
	guiApp := gui.New(svc, prefs)
	slog.Info("GUI application created, starting event loop...")
	fmt.Println("✓ Opening GUI window...")
	fmt.Println("Note: If window doesn't appear, you may need to run from Terminal.app or iTerm2")

	// Run GUI in a goroutine
	go func() {
		if err := guiApp.Run(); err != nil {
			slog.Error("GUI error", "error", err)
		}
	}()

	// app.Main() must be called from the main goroutine on macOS
	guiApp.Main()
	return nil
}

// runCLI starts the application in CLI mode (existing functionality)
func runCLI() error {
	// Backward compatibility: if no arguments provided, use default URL
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "https://git.vseinstrumenti.net/fd/account-balance/-/merge_requests/591")
	}

	cnf, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	var (
		client = &http.Client{Timeout: 30 * time.Second}
		mrURL  = flag.String("mr", "", "URL of the merge request")
		team   = flag.String("team", "", "Team to notify (e.g., @team-name)")
		action = flag.String("action", "review", "what to do (deploy or review)")
		_      = flag.Bool("cli", false, "Force CLI mode (default when using -mr, -team, or -action flags)")
	)

	flag.Parse()

	if *team != "" {
		cnf.Message.Team = *team
	}

	gitlabHost := extractHost(cnf.Gitlab.BaseURL)
	svc := manager.New(
		cnf.Message,
		shower.New(
			gitlab.New(client, cnf.Gitlab),
			jira.New(client, cnf.Jira),
			gitlabHost,
			cnf.Jira.ProjectPrefix,
		),
	)

	message, err := svc.Execute(context.Background(), domain.ActionType(*action), *mrURL, domain.ActionOptions{
		After:    time.Minute * 30,
		Timezone: "Europe/Moscow",
	})

	if err != nil {
		return fmt.Errorf("format: %w", err)
	}

	// Вывод
	fmt.Println(message)

	return nil
}

// extractHost extracts the hostname from a base URL string.
// e.g. "https://git.vseinstrumenti.net" -> "git.vseinstrumenti.net"
func extractHost(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return parsed.Host
}
