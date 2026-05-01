package gui

import (
	"strings"

	"gioui.org/app"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/widget"

	"review-info/internal/domain"
	"review-info/internal/preferences"
)

// Controller handles all business logic and mutable state for the GUI.
type Controller struct {
	service domain.ActionRunner
	prefs   *preferences.Preferences
	window  *app.Window

	// UI Components
	mrURLInput         *InputWithPaste
	teamEditor         widget.Editor
	actionEnum         widget.Enum
	timezoneDropdown   *Dropdown
	migrationsCheckbox widget.Bool
	resultOutput       *OutputWithCopy

	// State
	mrURL              string
	team               string
	action             string
	timezone           string
	migrationsApplied  bool
	result             string
	loading            bool
	error              string
	clipboardError     string
	mrURLError         string

	// Tracking for auto-trigger
	lastAction            string
	lastTimezone          string
	lastMigrationsApplied bool
}

// NewController creates a new Controller with initial state from preferences.
func NewController(service domain.ActionRunner, prefs *preferences.Preferences) *Controller {
	timezones := []string{
		"Europe/Moscow",
		"Europe/Berlin",
	}

	timezoneIndex := 0
	for i, tz := range timezones {
		if tz == prefs.Timezone {
			timezoneIndex = i
			break
		}
	}

	teamEditor := widget.Editor{
		SingleLine: true,
		Submit:     false,
	}
	teamEditor.SetText(prefs.Team)

	actionEnum := widget.Enum{Value: prefs.Action}

	return &Controller{
		service: service,
		prefs:   prefs,

		mrURLInput:         NewInputWithPaste(),
		teamEditor:         teamEditor,
		actionEnum:         actionEnum,
		timezoneDropdown:   NewDropdown(timezones, timezoneIndex),
		resultOutput:       NewOutputWithCopy(),
		migrationsCheckbox: widget.Bool{},

		team:     prefs.Team,
		action:   prefs.Action,
		timezone: prefs.Timezone,

		lastAction:            prefs.Action,
		lastTimezone:          prefs.Timezone,
		lastMigrationsApplied: false,
	}
}

// HandleEvents processes UI events and business logic.
func (c *Controller) HandleEvents(gtx layout.Context) {
	triggerGenerate := false

	// Handle clipboard paste events for the window
	for {
		ev, ok := gtx.Event(transfer.TargetFilter{Target: c.window, Type: "application/text"})
		if !ok {
			break
		}
		if de, ok := ev.(transfer.DataEvent); ok {
			clipboardText := ExtractClipboardText(de)
			if clipboardText != "" {
				c.clipboardError = ""
				trimmedText := strings.TrimSpace(clipboardText)
				if trimmedText == "" {
					c.clipboardError = "Ошибка: буфер обмена пуст"
				} else {
					c.mrURLInput.SetText(trimmedText)
					c.mrURL = trimmedText
					c.mrURLError = ""
					triggerGenerate = true
				}
			} else {
				c.clipboardError = "Ошибка: не удалось прочитать буфер обмена"
			}
		}
	}

	// Read current UI state (set during previous frame's layout)
	newAction := c.actionEnum.Value
	newTimezone := c.timezoneDropdown.SelectedText()
	newMigrations := c.migrationsCheckbox.Value

	// Detect changes that should trigger generation
	if newAction != c.lastAction {
		triggerGenerate = true
	}
	if newTimezone != c.lastTimezone {
		triggerGenerate = true
	}
	if newMigrations != c.lastMigrationsApplied {
		triggerGenerate = true
	}

	// Save for next frame comparison
	c.lastAction = newAction
	c.lastTimezone = newTimezone
	c.lastMigrationsApplied = newMigrations

	// Update current state
	c.action = newAction
	c.timezone = newTimezone
	c.migrationsApplied = newMigrations

	oldMRURL := c.mrURL
	c.mrURL = c.mrURLInput.Text()
	c.team = c.teamEditor.Text()

	// Clear errors if user manually edits the MR URL field
	if oldMRURL != c.mrURL {
		c.clipboardError = ""
		c.mrURLError = ""
		if strings.Contains(c.error, "MR URL") || strings.Contains(c.error, "URL") {
			c.error = ""
		}
	}

	// Auto-trigger generation
	if triggerGenerate && !c.loading && strings.TrimSpace(c.mrURL) != "" {
		c.clipboardError = ""
		c.handleGenerate()
	}
}

// handleGenerate processes the generate button click and calls the appropriate service method
func (c *Controller) handleGenerate() {
	c.error = ""
	c.mrURLError = ""

	// Validate MR URL is not empty
	if err := ValidateNonEmpty(c.mrURL, "MR URL"); err != nil {
		c.mrURLError = err.Error()
		return
	}

	// Validate MR URL format
	if err := ValidateMRURL(c.mrURL); err != nil {
		c.mrURLError = err.Error()
		return
	}

	mrURL := strings.TrimSpace(c.mrURL)

	// Set loading state and shrink window to base height
	c.loading = true
	c.result = ""
	c.window.Option(app.Size(windowWidth, windowHeight))

	// Process in a goroutine to avoid blocking the UI
	go func() {
		result, err := c.service.Execute(c.action, mrURL, domain.ActionOptions{
			Timezone:          c.timezone,
			MigrationsApplied: c.migrationsApplied,
		})

		c.loading = false
		if err != nil {
			c.error = FormatErrorMessage(err)
			c.result = ""
			c.window.Option(app.Size(windowWidth, windowHeight))
		} else {
			c.result = result
			c.error = ""
			c.window.Option(app.Size(windowWidth, resultHeight))
			c.savePreferences()
		}

		c.window.Invalidate()
	}()
}

// savePreferences saves the current preferences to disk.
func (c *Controller) savePreferences() {
	c.prefs.Action = c.action
	c.prefs.Timezone = c.timezone
	c.prefs.Team = c.team

	if err := c.prefs.Save(); err != nil {
		_ = err
	}
}
