package gui

import (
	"strings"

	"gioui.org/app"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"review-info/internal/preferences"
	"review-info/internal/service/manager"
)

const (
	windowWidth  = unit.Dp(700)
	windowHeight = unit.Dp(420)
	resultHeight = unit.Dp(600) // windowHeight + space for result section
)

// App represents the GUI application
type App struct {
	window  *app.Window
	theme   *material.Theme
	prefs   *preferences.Preferences
	service *manager.Service

	// UI Components
	mrURLInput          *InputWithPaste
	teamEditor          widget.Editor
	actionEnum          widget.Enum
	timezoneDropdown    *Dropdown
	migrationsCheckbox  widget.Bool
	resultOutput        *OutputWithCopy
	scrollList          widget.List

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

// New creates a new GUI application
func New(service *manager.Service, prefs *preferences.Preferences) *App {
	// Initialize timezone dropdown with available options
	timezones := []string{
		"Europe/Moscow",
		"Europe/Berlin",
	}

	// Find the index of the preferred timezone
	timezoneIndex := 0
	for i, tz := range timezones {
		if tz == prefs.Timezone {
			timezoneIndex = i
			break
		}
	}

	// Initialize team editor
	teamEditor := widget.Editor{
		SingleLine: true,
		Submit:     false,
	}
	teamEditor.SetText(prefs.Team)

	// Initialize action enum based on preferences
	actionEnum := widget.Enum{Value: prefs.Action}

	// Create custom theme with better colors
	theme := material.NewTheme()

	return &App{
		window:  nil, // Will be created in Run()
		theme:   theme,
		prefs:   prefs,
		service: service,

		// Initialize UI components
		mrURLInput:       NewInputWithPaste(),
		teamEditor:       teamEditor,
		actionEnum:       actionEnum,
		timezoneDropdown: NewDropdown(timezones, timezoneIndex),
		resultOutput:     NewOutputWithCopy(),
		scrollList: widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},

		// Initialize state from preferences
		mrURL:    "",
		team:     prefs.Team,
		action:   prefs.Action,
		timezone: prefs.Timezone,
		result:   "",
		loading:  false,
		error:    "",

		// Initialize tracking fields
		lastAction:            prefs.Action,
		lastTimezone:          prefs.Timezone,
		lastMigrationsApplied: false,
	}
}

// Main starts the main event loop (must be called from main goroutine on macOS)
func (a *App) Main() {
	app.Main()
}

// Run starts the GUI application and event loop
func (a *App) Run() error {
	// Create window
	a.window = new(app.Window)
	a.window.Option(
		app.Title("Review Info - GUI"),
		app.Size(windowWidth, windowHeight),
	)

	// Start event loop
	var ops op.Ops
	for {
		switch e := a.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			// Create graphics context
			gtx := app.NewContext(&ops, e)

			// Handle events
			a.handleEvents(gtx)

			// Render layout
			a.layout(gtx)

			// Submit frame
			e.Frame(gtx.Ops)
		}
	}
}

// handleEvents processes UI events and business logic
func (a *App) handleEvents(gtx layout.Context) {
	triggerGenerate := false

	// Handle clipboard paste events for the window
	for {
		ev, ok := gtx.Event(transfer.TargetFilter{Target: a.window, Type: "application/text"})
		if !ok {
			break
		}
		if de, ok := ev.(transfer.DataEvent); ok {
			clipboardText := ExtractClipboardText(de)
			if clipboardText != "" {
				a.clipboardError = ""
				trimmedText := strings.TrimSpace(clipboardText)
				if trimmedText == "" {
					a.clipboardError = "Ошибка: буфер обмена пуст"
				} else {
					a.mrURLInput.SetText(trimmedText)
					a.mrURL = trimmedText
					a.mrURLError = ""
					triggerGenerate = true
				}
			} else {
				a.clipboardError = "Ошибка: не удалось прочитать буфер обмена"
			}
		}
	}

	// Read current UI state (set during previous frame's layout)
	newAction := a.actionEnum.Value
	newTimezone := a.timezoneDropdown.SelectedText()
	newMigrations := a.migrationsCheckbox.Value

	// Detect changes that should trigger generation
	if newAction != a.lastAction {
		triggerGenerate = true
	}
	if newTimezone != a.lastTimezone {
		triggerGenerate = true
	}
	if newMigrations != a.lastMigrationsApplied {
		triggerGenerate = true
	}

	// Save for next frame comparison
	a.lastAction = newAction
	a.lastTimezone = newTimezone
	a.lastMigrationsApplied = newMigrations

	// Update current state
	a.action = newAction
	a.timezone = newTimezone
	a.migrationsApplied = newMigrations

	oldMRURL := a.mrURL
	a.mrURL = a.mrURLInput.Text()
	a.team = a.teamEditor.Text()

	// Clear errors if user manually edits the MR URL field
	if oldMRURL != a.mrURL {
		a.clipboardError = ""
		a.mrURLError = ""
		if strings.Contains(a.error, "MR URL") || strings.Contains(a.error, "URL") {
			a.error = ""
		}
	}

	// Auto-trigger generation
	if triggerGenerate && !a.loading && strings.TrimSpace(a.mrURL) != "" {
		a.clipboardError = ""
		a.handleGenerate()
	}
}

// handleGenerate processes the generate button click and calls the appropriate service method
func (a *App) handleGenerate() {
	a.error = ""
	a.mrURLError = ""

	// Validate MR URL is not empty
	if err := ValidateNonEmpty(a.mrURL, "MR URL"); err != nil {
		a.mrURLError = err.Error()
		return
	}

	// Validate MR URL format
	if err := ValidateMRURL(a.mrURL); err != nil {
		a.mrURLError = err.Error()
		return
	}

	mrURL := strings.TrimSpace(a.mrURL)

	// Set loading state and shrink window to base height
	a.loading = true
	a.result = ""
	a.window.Option(app.Size(windowWidth, windowHeight))

	// Process in a goroutine to avoid blocking the UI
	go func() {
		var result string
		var err error

		switch a.action {
		case "review":
			result, err = a.service.ReviewMe(mrURL)
		case "deploy":
			result, err = a.service.DeployPlaningWithTimezone(mrURL, 0, a.timezone, a.migrationsApplied)
		default:
			result, err = a.service.ReviewMe(mrURL)
		}

		a.loading = false
		if err != nil {
			a.error = FormatErrorMessage(err)
			a.result = ""
			a.window.Option(app.Size(windowWidth, windowHeight))
		} else {
			a.result = result
			a.error = ""
			a.window.Option(app.Size(windowWidth, resultHeight))
			a.savePreferences()
		}

		a.window.Invalidate()
	}()
}

// savePreferences saves the current preferences to disk
// Errors are logged but don't block the application
func (a *App) savePreferences() {
	a.prefs.Action = a.action
	a.prefs.Timezone = a.timezone
	a.prefs.Team = a.team

	if err := a.prefs.Save(); err != nil {
		_ = err
	}
}

// layout renders the application UI with scrolling support
func (a *App) layout(gtx layout.Context) layout.Dimensions {
	return layout.Inset{
		Top:    StandardPadding,
		Bottom: StandardPadding,
		Left:   StandardPadding,
		Right:  StandardPadding,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return material.List(a.theme, &a.scrollList).Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutInputSection(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if a.result != "" {
						return a.layoutResultSection(gtx)
					}
					return layout.Dimensions{}
				}),
			)
		})
	})
}

// layoutInputSection renders the input parameters section
func (a *App) layoutInputSection(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// MR URL input with paste button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return CreateLabel(gtx, a.theme, "MR URL:", BodySize)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return a.mrURLInput.Layout(gtx, a.theme, a.window, "https://git.../merge_requests/...")
					})
				}),
				// MR URL validation error (if any)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if a.mrURLError != "" {
						return layout.Inset{Bottom: SmallPadding, Left: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return CreateLabel(gtx, a.theme, a.mrURLError, SmallSize)
						})
					}
					return layout.Dimensions{}
				}),
				// Clipboard error message (if any)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if a.clipboardError != "" {
						return layout.Inset{Bottom: SmallPadding, Left: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return CreateLabel(gtx, a.theme, a.clipboardError, SmallSize)
						})
					}
					return layout.Dimensions{}
				}),
			)
		}),

		// Team input
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return CreateLabel(gtx, a.theme, "Команда:", BodySize)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return CreateBorderedInput(gtx, a.theme, &a.teamEditor, "@team-name")
					})
				}),
			)
		}),

		// Action radio buttons
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return CreateLabel(gtx, a.theme, "Действие:", BodySize)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: SmallPadding, Left: StandardPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis:    layout.Horizontal,
							Spacing: layout.SpaceSides,
						}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								radio := material.RadioButton(a.theme, &a.actionEnum, "review", "Review")
								return layout.Inset{Right: unit.Dp(24)}.Layout(gtx, radio.Layout)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								radio := material.RadioButton(a.theme, &a.actionEnum, "deploy", "Deploy")
								return radio.Layout(gtx)
							}),
						)
					})
				}),
			)
		}),

		// Timezone dropdown (only shown when Deploy is selected)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if a.action == "deploy" {
				return layout.Inset{Top: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Bottom: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return CreateLabel(gtx, a.theme, "Часовой пояс:", BodySize)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Left:   StandardPadding,
								Bottom: SmallPadding,
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Max.X = gtx.Dp(unit.Dp(250))
								return a.timezoneDropdown.Layout(gtx, a.theme)
							})
						}),
					)
				})
			}
			return layout.Dimensions{}
		}),

		// Migrations checkbox (only shown when Deploy is selected)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if a.action == "deploy" {
				return layout.Inset{
					Bottom: SmallPadding,
					Left:   StandardPadding,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					checkBox := material.CheckBox(a.theme, &a.migrationsCheckbox, "Миграции применены в проде")
					checkBox.TextSize = BodySize
					return checkBox.Layout(gtx)
				})
			}
			return layout.Dimensions{}
		}),

		// Loading indicator
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if a.loading {
				return layout.Inset{Top: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return CreateLabel(gtx, a.theme, "Загрузка данных, пожалуйста подождите...", unit.Sp(13))
					})
				})
			}
			return layout.Dimensions{}
		}),

		// General error message (network/API errors)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if a.error != "" {
				return layout.Inset{Top: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top:    StandardPadding,
								Bottom: StandardPadding,
								Left:   StandardPadding,
								Right:  StandardPadding,
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return CreateLabel(gtx, a.theme, "Ошибка: "+a.error, BodySize)
							})
						}),
					)
				})
			}
			return layout.Dimensions{}
		}),
	)
}

// layoutResultSection renders the result output section
func (a *App) layoutResultSection(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Top: StandardPadding, Bottom: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Bottom: SmallPadding,
					Left:   StandardPadding,
					Right:  StandardPadding,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return CreateSectionTitle(gtx, a.theme, "Результат")
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				a.resultOutput.SetText(a.result)

				return layout.Inset{
					Left:   StandardPadding,
					Right:  StandardPadding,
					Bottom: StandardPadding,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return a.resultOutput.Layout(gtx, a.theme, a.window)
				})
			}),
		)
	})
}
