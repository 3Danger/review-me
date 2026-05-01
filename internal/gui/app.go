package gui

import (
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"review-info/internal/domain"
	"review-info/internal/preferences"
)

const (
	windowWidth  = unit.Dp(700)
	windowHeight = unit.Dp(420)
	resultHeight = unit.Dp(600)
)

// App represents the GUI application view.
type App struct {
	window *app.Window
	theme  *material.Theme
	ctrl   *Controller

	scrollList widget.List
}

// New creates a new GUI application.
func New(service domain.ActionRunner, prefs *preferences.Preferences) *App {
	theme := material.NewTheme()

	return &App{
		window: nil, // Will be created in Run()
		theme:  theme,
		ctrl:   NewController(service, prefs),
		scrollList: widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
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
	a.ctrl.window = a.window

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
			a.ctrl.HandleEvents(gtx)

			// Render layout
			a.layout(gtx)

			// Submit frame
			e.Frame(gtx.Ops)
		}
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
					if a.ctrl.result != "" {
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
						return a.ctrl.mrURLInput.Layout(gtx, a.theme, a.window, "https://git.../merge_requests/...")
					})
				}),
				// MR URL validation error (if any)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if a.ctrl.mrURLError != "" {
						return layout.Inset{Bottom: SmallPadding, Left: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return CreateLabel(gtx, a.theme, a.ctrl.mrURLError, SmallSize)
						})
					}
					return layout.Dimensions{}
				}),
				// Clipboard error message (if any)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if a.ctrl.clipboardError != "" {
						return layout.Inset{Bottom: SmallPadding, Left: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return CreateLabel(gtx, a.theme, a.ctrl.clipboardError, SmallSize)
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
						return CreateBorderedInput(gtx, a.theme, &a.ctrl.teamEditor, "@team-name")
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
								radio := material.RadioButton(a.theme, &a.ctrl.actionEnum, "review", "Review")
								return layout.Inset{Right: unit.Dp(24)}.Layout(gtx, radio.Layout)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								radio := material.RadioButton(a.theme, &a.ctrl.actionEnum, "deploy", "Deploy")
								return radio.Layout(gtx)
							}),
						)
					})
				}),
			)
		}),

		// Timezone dropdown (only shown when Deploy is selected)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if a.ctrl.action == "deploy" {
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
								return a.ctrl.timezoneDropdown.Layout(gtx, a.theme)
							})
						}),
					)
				})
			}
			return layout.Dimensions{}
		}),

		// Migrations checkbox (only shown when Deploy is selected)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if a.ctrl.action == "deploy" {
				return layout.Inset{
					Bottom: SmallPadding,
					Left:   StandardPadding,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					checkBox := material.CheckBox(a.theme, &a.ctrl.migrationsCheckbox, "Миграции применены в проде")
					checkBox.TextSize = BodySize
					return checkBox.Layout(gtx)
				})
			}
			return layout.Dimensions{}
		}),

		// Loading indicator
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if a.ctrl.loading {
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
			if a.ctrl.error != "" {
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
								return CreateLabel(gtx, a.theme, "Ошибка: "+a.ctrl.error, BodySize)
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
				a.ctrl.resultOutput.SetText(a.ctrl.result)

				return layout.Inset{
					Left:   StandardPadding,
					Right:  StandardPadding,
					Bottom: StandardPadding,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return a.ctrl.resultOutput.Layout(gtx, a.theme, a.window)
				})
			}),
		)
	})
}
