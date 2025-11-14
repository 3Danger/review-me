package gui

import (
	"io"
	"time"

	"gioui.org/app"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// InputWithPaste is a text editor component with a paste button
type InputWithPaste struct {
	Editor   widget.Editor
	PasteBtn widget.Clickable
}

// NewInputWithPaste creates a new InputWithPaste component
func NewInputWithPaste() *InputWithPaste {
	return &InputWithPaste{
		Editor: widget.Editor{
			SingleLine: true,
			Submit:     false,
		},
	}
}

// Layout renders the input field with paste button
// The parent should handle clipboard events and call SetText when data is received
func (i *InputWithPaste) Layout(gtx layout.Context, th *material.Theme, w *app.Window, hint string) layout.Dimensions {
	// Handle paste button click
	if i.PasteBtn.Clicked(gtx) {
		// Request clipboard read
		readClipboard(w, gtx)
	}

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			// Add border around the input field
			border := widget.Border{
				Color:        th.Palette.Fg,
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
					editor := material.Editor(th, &i.Editor, hint)
					editor.TextSize = unit.Sp(14)
					return editor.Layout(gtx)
				})
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(th, &i.PasteBtn, "📋 Вставить")
				btn.TextSize = unit.Sp(13)
				btn.Inset = layout.Inset{
					Top:    unit.Dp(6),
					Bottom: unit.Dp(6),
					Left:   unit.Dp(12),
					Right:  unit.Dp(12),
				}
				btn.CornerRadius = unit.Dp(4)
				return btn.Layout(gtx)
			})
		}),
	)
}

// Text returns the current text in the editor
func (i *InputWithPaste) Text() string {
	return i.Editor.Text()
}

// SetText sets the text in the editor
func (i *InputWithPaste) SetText(text string) {
	i.Editor.SetText(text)
}

// OutputWithCopy is a text display component with copy button and feedback
type OutputWithCopy struct {
	text         string
	CopyBtn      widget.Clickable
	copied       bool
	copiedAt     time.Time
	feedbackDone bool
	copyError    string
	errorAt      time.Time
}

// NewOutputWithCopy creates a new OutputWithCopy component
func NewOutputWithCopy() *OutputWithCopy {
	return &OutputWithCopy{}
}

// SetText sets the text to display
func (o *OutputWithCopy) SetText(text string) {
	o.text = text
}

// Text returns the current text
func (o *OutputWithCopy) Text() string {
	return o.text
}

// Layout renders the output area with copy button
func (o *OutputWithCopy) Layout(gtx layout.Context, th *material.Theme, w *app.Window) layout.Dimensions {
	// Handle copy button click
	if o.CopyBtn.Clicked(gtx) {
		// Clear previous error
		o.copyError = ""

		if o.text == "" {
			o.copyError = "Ошибка: нет текста для копирования"
			o.errorAt = time.Now()
		} else {
			err := writeClipboard(w, gtx, o.text)
			if err == nil {
				o.copied = true
				o.copiedAt = time.Now()
				o.feedbackDone = false
			} else {
				o.copyError = "Ошибка: не удалось скопировать в буфер обмена"
				o.errorAt = time.Now()
			}
		}
	}

	// Check if feedback period has expired (2 seconds)
	if o.copied && !o.feedbackDone {
		if time.Since(o.copiedAt) > 2*time.Second {
			o.copied = false
			o.feedbackDone = true
		} else {
			// Request another frame to update the UI
			w.Invalidate()
		}
	}

	// Check if error display period has expired (3 seconds)
	if o.copyError != "" && time.Since(o.errorAt) > 3*time.Second {
		o.copyError = ""
	} else if o.copyError != "" {
		// Request another frame to update the UI
		w.Invalidate()
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// Text display area with border
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(4),
				Bottom: unit.Dp(4),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				// Add border around the result field
				border := widget.Border{
					Color:        th.Palette.Fg,
					Width:        unit.Dp(1),
					CornerRadius: unit.Dp(4),
				}
				return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top:    unit.Dp(8),
						Bottom: unit.Dp(8),
						Left:   unit.Dp(8),
						Right:  unit.Dp(8),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						// Create a read-only editor for displaying text
						editor := widget.Editor{
							SingleLine: false,
							ReadOnly:   true,
						}
						editor.SetText(o.text)

						// Set a reasonable height for the text area
						gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(100))
						gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(150))

						ed := material.Editor(th, &editor, "")
						ed.TextSize = unit.Sp(14)
						return ed.Layout(gtx)
					})
				})
			})
		}),
		// Copy button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					btnText := "📋 Скопировать в буфер обмена"
					if o.copied {
						btnText = "✅ Скопировано!"
					}
					btn := material.Button(th, &o.CopyBtn, btnText)
					btn.TextSize = unit.Sp(14)
					btn.Inset = layout.Inset{
						Top:    unit.Dp(8),
						Bottom: unit.Dp(8),
						Left:   unit.Dp(16),
						Right:  unit.Dp(16),
					}
					btn.CornerRadius = unit.Dp(4)
					return btn.Layout(gtx)
				})
			})
		}),
		// Copy error message (if any)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if o.copyError != "" {
				return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(th, o.copyError)
						label.TextSize = unit.Sp(12)
						return label.Layout(gtx)
					})
				})
			}
			return layout.Dimensions{}
		}),
	)
}

// Dropdown is a selectable list component for choosing from options
type Dropdown struct {
	List     widget.List
	options  []string
	selected int
	buttons  []widget.Clickable
}

// NewDropdown creates a new Dropdown component with the given options
func NewDropdown(options []string, defaultIndex int) *Dropdown {
	if defaultIndex < 0 || defaultIndex >= len(options) {
		defaultIndex = 0
	}

	return &Dropdown{
		List: widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		options:  options,
		selected: defaultIndex,
		buttons:  make([]widget.Clickable, len(options)),
	}
}

// Layout renders the dropdown list
func (d *Dropdown) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return d.List.Layout(gtx, len(d.options), func(gtx layout.Context, index int) layout.Dimensions {
		// Check if this option was clicked
		if d.buttons[index].Clicked(gtx) {
			d.selected = index
		}

		// Determine button style based on selection
		btn := material.Button(th, &d.buttons[index], d.options[index])
		btn.TextSize = unit.Sp(13)
		btn.Inset = layout.Inset{
			Top:    unit.Dp(6),
			Bottom: unit.Dp(6),
			Left:   unit.Dp(12),
			Right:  unit.Dp(12),
		}
		btn.CornerRadius = unit.Dp(4)

		// Visual feedback for selected item
		if index == d.selected {
			// Make selected item more prominent
			btn.Background = th.Palette.ContrastBg
			btn.Color = th.Palette.ContrastFg
		}

		return layout.Inset{
			Top:    unit.Dp(2),
			Bottom: unit.Dp(2),
		}.Layout(gtx, btn.Layout)
	})
}

// Selected returns the index of the currently selected option
func (d *Dropdown) Selected() int {
	return d.selected
}

// SelectedText returns the text of the currently selected option
func (d *Dropdown) SelectedText() string {
	if d.selected >= 0 && d.selected < len(d.options) {
		return d.options[d.selected]
	}
	return ""
}

// SetSelected sets the selected option by index
func (d *Dropdown) SetSelected(index int) {
	if index >= 0 && index < len(d.options) {
		d.selected = index
	}
}

// handleClipboardData extracts text from a transfer.DataEvent
func handleClipboardData(de transfer.DataEvent) string {
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
