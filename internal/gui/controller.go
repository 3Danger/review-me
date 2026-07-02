package gui

import (
	"context"
	"log/slog"
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
	// Dependencies
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
	form        FormState
	execution   ExecutionState
	autoTrigger AutoTrigger

	// Error fields (UI display only)
	clipboardError string
	mrURLError     string
}

// TimezoneOptions is the list of timezones shown in the GUI dropdown.
// Can be overridden at package level or via config.
var TimezoneOptions = []string{"Europe/Moscow", "Europe/Berlin"}

// NewController creates a new Controller with initial state from preferences.
func NewController(service domain.ActionRunner, prefs *preferences.Preferences) *Controller {
	timezones := TimezoneOptions

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

	ctrl := &Controller{
		service: service,
		prefs:   prefs,

		mrURLInput:         NewInputWithPaste(),
		teamEditor:         teamEditor,
		actionEnum:         actionEnum,
		timezoneDropdown:   NewDropdown(timezones, timezoneIndex),
		resultOutput:       NewOutputWithCopy(),
		migrationsCheckbox: widget.Bool{},
	}

	ctrl.form.Action = prefs.Action
	ctrl.form.Timezone = prefs.Timezone
	ctrl.form.Team = prefs.Team

	ctrl.autoTrigger.Record(prefs.Action, prefs.Timezone, false)

	return ctrl
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
					c.form.MRURL = trimmedText
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
	if c.autoTrigger.HasChanges(newAction, newTimezone, newMigrations) {
		triggerGenerate = true
	}
	c.autoTrigger.Record(newAction, newTimezone, newMigrations)

	// Update current state
	c.form.Action = newAction
	c.form.Timezone = newTimezone
	c.form.MigrationsApplied = newMigrations

	oldMRURL := c.form.MRURL
	c.form.MRURL = c.mrURLInput.Text()
	c.form.Team = c.teamEditor.Text()

	// Clear errors if user manually edits the MR URL field
	if oldMRURL != c.form.MRURL {
		c.clipboardError = ""
		c.mrURLError = ""
		if strings.Contains(c.execution.Error(), "MR URL") || strings.Contains(c.execution.Error(), "URL") {
			c.execution.SetError("")
		}
	}

	// Auto-trigger generation
	if triggerGenerate && !c.execution.IsLoading() && strings.TrimSpace(c.form.MRURL) != "" {
		c.clipboardError = ""
		c.handleGenerate()
	}
}

// handleGenerate processes the generate button click and calls the appropriate service method
func (c *Controller) handleGenerate() {
	if !c.execution.TryAcquire() {
		return
	}
	c.execution.SetError("")

	c.mrURLError = ""

	// Validate MR URL is not empty
	if err := ValidateNonEmpty(c.form.MRURL, "MR URL"); err != nil {
		c.mrURLError = err.Error()
		c.execution.Release()
		return
	}

	// Validate MR URL format
	if err := ValidateMRURL(c.form.MRURL); err != nil {
		c.mrURLError = err.Error()
		c.execution.Release()
		return
	}

	mrURL := strings.TrimSpace(c.form.MRURL)

	// Set loading state and shrink window to base height
	c.execution.SetLoading(true)
	c.execution.SetResult("")
	c.window.Option(app.Size(windowWidth, windowHeight))

	// Process in a goroutine to avoid blocking the UI
	go func() {
		result, err := c.service.Execute(context.Background(), domain.ActionType(c.form.Action), mrURL, domain.ActionOptions{
			Timezone:          c.form.Timezone,
			MigrationsApplied: c.form.MigrationsApplied,
		})

		c.execution.Release()
		c.execution.SetLoading(false)
		if err != nil {
			c.execution.SetError(FormatErrorMessage(err))
			c.execution.SetResult("")
			c.window.Option(app.Size(windowWidth, windowHeight))
		} else {
			c.execution.SetResult(result)
			c.execution.SetError("")
			c.window.Option(app.Size(windowWidth, resultHeight))
			c.savePreferences()
		}

		c.window.Invalidate()
	}()
}

// savePreferences saves the current preferences to disk.
func (c *Controller) savePreferences() {
	c.prefs.Action = c.form.Action
	c.prefs.Timezone = c.form.Timezone
	c.prefs.Team = c.form.Team

	if err := c.prefs.Save(); err != nil {
		slog.Error("saving preferences", "component", "gui", "operation", "save_preferences", "error", err)
	}
}
