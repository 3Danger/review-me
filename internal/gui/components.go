package gui

import (
	"time"

	"gioui.org/app"
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
		readClipboard(gtx, w)
	}

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return CreateBorderedInput(gtx, th, &i.Editor, hint)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: StandardPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return CreateButton(gtx, th, &i.PasteBtn, "📋 Вставить", ButtonSmall)
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
			err := writeClipboard(gtx, o.text)
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
				Top:    SmallPadding,
				Bottom: SmallPadding,
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

				// Use CreateBorderedInput but wrap it to handle the read-only editor
				border := widget.Border{
					Color:        th.Palette.Fg,
					Width:        unit.Dp(1),
					CornerRadius: StandardCornerRadius,
				}
				return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top:    StandardPadding,
						Bottom: StandardPadding,
						Left:   StandardPadding,
						Right:  StandardPadding,
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						ed := material.Editor(th, &editor, "")
						ed.TextSize = BodySize
						return ed.Layout(gtx)
					})
				})
			})
		}),
		// Copy button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: StandardPadding, Bottom: SmallPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					btnText := "📋 Скопировать в буфер обмена"
					if o.copied {
						btnText = "✅ Скопировано!"
					}
					return CreateButton(gtx, th, &o.CopyBtn, btnText, ButtonMedium)
				})
			})
		}),
		// Copy error message (if any)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if o.copyError != "" {
				return layout.Inset{Top: StandardPadding, Bottom: StandardPadding}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return CreateLabel(gtx, th, o.copyError, SmallSize)
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

		// Visual feedback for selected item
		// We need to customize the button for selection, so we'll use material.Button directly
		// but with constants for sizing
		btn := material.Button(th, &d.buttons[index], d.options[index])
		btn.TextSize = unit.Sp(13)
		btn.Inset = layout.Inset{
			Top:    unit.Dp(6),
			Bottom: unit.Dp(6),
			Left:   unit.Dp(12),
			Right:  unit.Dp(12),
		}
		btn.CornerRadius = StandardCornerRadius

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
