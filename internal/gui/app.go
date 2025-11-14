package gui

import (
	"io"
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

// App represents the GUI application
type App struct {
	window  *app.Window
	theme   *material.Theme
	prefs   *preferences.Preferences
	service *manager.Service

	// UI Components
	mrURLInput       *InputWithPaste
	teamEditor       widget.Editor
	actionEnum       widget.Enum
	timezoneDropdown *Dropdown
	generateBtn      widget.Clickable
	resultOutput     *OutputWithCopy
	scrollList       widget.List

	// State
	mrURL          string
	team           string
	action         string
	timezone       string
	result         string
	loading        bool
	error          string
	clipboardError string
	mrURLError     string
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
		generateBtn:      widget.Clickable{},
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
		app.Size(unit.Dp(700), unit.Dp(500)),
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
	// Handle clipboard paste events for the window
	for {
		ev, ok := gtx.Event(transfer.TargetFilter{Target: a.window, Type: "application/text"})
		if !ok {
			break
		}
		if de, ok := ev.(transfer.DataEvent); ok {
			clipboardText := extractClipboardText(de)
			if clipboardText != "" {
				// Clear any previous clipboard error
				a.clipboardError = ""
				// Validate clipboard content is not empty (after trimming whitespace)
				trimmedText := strings.TrimSpace(clipboardText)
				if trimmedText == "" {
					a.clipboardError = "Ошибка: буфер обмена пуст"
				} else {
					// Set MR URL field value to clipboard content
					a.mrURLInput.SetText(trimmedText)
					a.mrURL = trimmedText
					// Clear MR URL error when pasting
					a.mrURLError = ""
				}
			} else {
				// Show error notification if clipboard read fails
				a.clipboardError = "Ошибка: не удалось прочитать буфер обмена"
			}
		}
	}

	// Handle generate button click
	if a.generateBtn.Clicked(gtx) && !a.loading {
		// Clear clipboard error when generating
		a.clipboardError = ""
		a.handleGenerate()
	}

	// Update state from UI components
	oldMRURL := a.mrURL
	a.mrURL = a.mrURLInput.Text()
	a.team = a.teamEditor.Text()
	a.action = a.actionEnum.Value
	a.timezone = a.timezoneDropdown.SelectedText()

	// Clear errors if user manually edits the MR URL field
	if oldMRURL != a.mrURL {
		a.clipboardError = ""
		a.mrURLError = ""
		// Also clear general error if it was related to MR URL
		if strings.Contains(a.error, "MR URL") || strings.Contains(a.error, "URL") {
			a.error = ""
		}
	}
}

// handleGenerate processes the generate button click and calls the appropriate service method
func (a *App) handleGenerate() {
	// Clear previous errors
	a.error = ""
	a.mrURLError = ""

	// Validate MR URL is not empty
	mrURL := strings.TrimSpace(a.mrURL)
	if mrURL == "" {
		a.mrURLError = "Ошибка: MR URL не может быть пустым"
		return
	}

	// Validate MR URL format
	if !a.isValidMRURL(mrURL) {
		a.mrURLError = "Ошибка: неверный формат MR URL. Ожидается URL вида: https://gitlab.../merge_requests/..."
		return
	}

	// Set loading state
	a.loading = true
	a.result = ""

	// Process in a goroutine to avoid blocking the UI
	go func() {
		var result string
		var err error

		// Call appropriate service method based on action
		switch a.action {
		case "review":
			result, err = a.service.ReviewMe(mrURL)
		case "deploy":
			// Use 0 duration for immediate deployment
			result, err = a.service.DeployPlaningWithTimezone(mrURL, 0, a.timezone)
		default:
			// Default to review if action is not recognized
			result, err = a.service.ReviewMe(mrURL)
		}

		// Update UI state (this will be picked up in the next frame)
		a.loading = false
		if err != nil {
			// Format error message based on error type
			a.error = a.formatErrorMessage(err)
			a.result = ""
		} else {
			a.result = result
			a.error = ""

			// Save preferences after successful message generation
			a.savePreferences()
		}

		// Request a redraw
		a.window.Invalidate()
	}()
}

// isValidMRURL validates that the URL looks like a GitLab merge request URL
func (a *App) isValidMRURL(url string) bool {
	// Basic validation: check if it contains merge_requests
	// and looks like a URL (starts with http:// or https://)
	url = strings.TrimSpace(url)
	if url == "" {
		return false
	}

	// Check if it starts with http:// or https://
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}

	// Check if it contains "merge_requests" or "merge_request"
	if !strings.Contains(url, "merge_requests") && !strings.Contains(url, "merge_request") {
		return false
	}

	return true
}

// formatErrorMessage formats error messages in Russian with appropriate context
func (a *App) formatErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()

	// Check for common error patterns and provide user-friendly messages
	switch {
	case strings.Contains(errMsg, "connection refused"):
		return "Ошибка сети: не удалось подключиться к серверу. Проверьте подключение к интернету."
	case strings.Contains(errMsg, "timeout"):
		return "Ошибка сети: превышено время ожидания ответа от сервера. Попробуйте еще раз."
	case strings.Contains(errMsg, "no such host"):
		return "Ошибка сети: не удалось найти сервер. Проверьте URL."
	case strings.Contains(errMsg, "401") || strings.Contains(errMsg, "unauthorized"):
		return "Ошибка авторизации: проверьте токены доступа в config.yml."
	case strings.Contains(errMsg, "403") || strings.Contains(errMsg, "forbidden"):
		return "Ошибка доступа: недостаточно прав для выполнения операции."
	case strings.Contains(errMsg, "404") || strings.Contains(errMsg, "not found"):
		return "Ошибка: ресурс не найден. Проверьте правильность MR URL."
	case strings.Contains(errMsg, "500") || strings.Contains(errMsg, "internal server error"):
		return "Ошибка сервера: внутренняя ошибка сервера. Попробуйте позже."
	case strings.Contains(errMsg, "bad request"):
		return "Ошибка: неверный запрос. Проверьте правильность введенных данных."
	default:
		// Return the original error message with "Ошибка:" prefix
		return "Ошибка: " + errMsg
	}
}

// savePreferences saves the current preferences to disk
// Errors are logged but don't block the application
func (a *App) savePreferences() {
	// Update preferences with current values
	a.prefs.Action = a.action
	a.prefs.Timezone = a.timezone
	a.prefs.Team = a.team

	// Save to disk
	if err := a.prefs.Save(); err != nil {
		// Log error but don't block the application
		// In a production app, you might want to use a proper logger
		// For now, we'll just silently fail as per requirements
		_ = err
	}
}

// extractClipboardText extracts text from a transfer.DataEvent
func extractClipboardText(de transfer.DataEvent) string {
	// Check if it's text data
	if de.Type == "text/plain" || de.Type == "text/plain;charset=utf-8" || de.Type == "application/text" {
		reader := de.Open()
		if reader == nil {
			return ""
		}
		defer reader.Close()

		// Read the clipboard text
		data, err := io.ReadAll(reader)
		if err != nil {
			return ""
		}

		return string(data)
	}
	return ""
}

// layout renders the application UI with scrolling support
func (a *App) layout(gtx layout.Context) layout.Dimensions {
	return layout.Inset{
		Top:    unit.Dp(12),
		Bottom: unit.Dp(12),
		Left:   unit.Dp(16),
		Right:  unit.Dp(16),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Use a scrollable list to handle overflow
		return material.List(a.theme, &a.scrollList).Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				// Input section
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutInputSection(gtx)
				}),
				// Result section (only shown when there's a result)
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
		// Section title with background
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(4),
				Bottom: unit.Dp(12),
				Left:   unit.Dp(8),
				Right:  unit.Dp(8),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.H6(a.theme, "Параметры")
				return label.Layout(gtx)
			})
		}),

		// MR URL input with paste button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body2(a.theme, "MR URL:")
					return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, label.Layout)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return a.mrURLInput.Layout(gtx, a.theme, a.window, "https://git.../merge_requests/...")
					})
				}),
				// MR URL validation error (if any)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if a.mrURLError != "" {
						return layout.Inset{Bottom: unit.Dp(8), Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							label := material.Body2(a.theme, a.mrURLError)
							label.TextSize = unit.Sp(12)
							return label.Layout(gtx)
						})
					}
					return layout.Dimensions{}
				}),
				// Clipboard error message (if any)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if a.clipboardError != "" {
						return layout.Inset{Bottom: unit.Dp(16), Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							label := material.Body2(a.theme, a.clipboardError)
							label.TextSize = unit.Sp(12)
							return label.Layout(gtx)
						})
					}
					return layout.Inset{Bottom: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{}
					})
				}),
			)
		}),

		// Team input
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Body2(a.theme, "Команда:")
					return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, label.Layout)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						// Add border around the team input field
						border := widget.Border{
							Color:        a.theme.Palette.Fg,
							Width:        unit.Dp(1),
							CornerRadius: unit.Dp(4),
						}
						return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top:    unit.Dp(6),
								Bottom: unit.Dp(6),
								Left:   unit.Dp(8),
								Right:  unit.Dp(8),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								editor := material.Editor(a.theme, &a.teamEditor, "@team-name")
								editor.TextSize = unit.Sp(14)
								return editor.Layout(gtx)
							})
						})
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
					label := material.Body2(a.theme, "Действие:")
					return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, label.Layout)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top:    unit.Dp(4),
						Bottom: unit.Dp(4),
						Left:   unit.Dp(8),
						Right:  unit.Dp(8),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
				return layout.Inset{Top: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							label := material.Body2(a.theme, "Часовой пояс:")
							return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, label.Layout)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Left:   unit.Dp(8),
								Right:  unit.Dp(8),
								Bottom: unit.Dp(4),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								// Constrain the dropdown width
								gtx.Constraints.Max.X = gtx.Dp(unit.Dp(250))
								return a.timezoneDropdown.Layout(gtx, a.theme)
							})
						}),
					)
				})
			}
			return layout.Dimensions{}
		}),

		// Generate button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(16), Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					btnText := "✨ Сгенерировать"
					if a.loading {
						btnText = "⏳ Обработка..."
					}

					// Disable button during loading
					if a.loading {
						gtx = gtx.Disabled()
					}

					btn := material.Button(a.theme, &a.generateBtn, btnText)
					btn.TextSize = unit.Sp(16)
					btn.Inset = layout.Inset{
						Top:    unit.Dp(10),
						Bottom: unit.Dp(10),
						Left:   unit.Dp(24),
						Right:  unit.Dp(24),
					}
					btn.CornerRadius = unit.Dp(6)
					return btn.Layout(gtx)
				})
			})
		}),

		// Loading indicator
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if a.loading {
				return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(a.theme, "Загрузка данных, пожалуйста подождите...")
						label.TextSize = unit.Sp(13)
						return label.Layout(gtx)
					})
				})
			}
			return layout.Dimensions{}
		}),

		// General error message (network/API errors)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if a.error != "" {
				return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							// Error box with padding and red background
							return layout.Inset{
								Top:    unit.Dp(12),
								Bottom: unit.Dp(12),
								Left:   unit.Dp(16),
								Right:  unit.Dp(16),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								label := material.Body2(a.theme, "Ошибка: "+a.error)
								label.TextSize = unit.Sp(14)
								return label.Layout(gtx)
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
	return layout.Inset{Top: unit.Dp(16), Bottom: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// Section title with background
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    unit.Dp(4),
					Bottom: unit.Dp(8),
					Left:   unit.Dp(8),
					Right:  unit.Dp(8),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.H6(a.theme, "Результат")
					return label.Layout(gtx)
				})
			}),
			// Result text area
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Update the result output component with current result
				a.resultOutput.SetText(a.result)

				return layout.Inset{
					Left:   unit.Dp(8),
					Right:  unit.Dp(8),
					Bottom: unit.Dp(8),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return a.resultOutput.Layout(gtx, a.theme, a.window)
				})
			}),
		)
	})
}
