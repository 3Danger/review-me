package gui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// UI Constants for consistent styling

const (
	// Padding constants
	StandardPadding = unit.Dp(8)
	SmallPadding    = unit.Dp(4)
	LargePadding    = unit.Dp(16)

	// Font size constants
	TitleSize = unit.Sp(16)
	BodySize  = unit.Sp(14)
	SmallSize = unit.Sp(12)

	// Corner radius constants
	StandardCornerRadius = unit.Dp(4)
	ButtonCornerRadius   = unit.Dp(6)
)

// ButtonSize represents the size of a button
type ButtonSize int

const (
	ButtonSmall ButtonSize = iota
	ButtonMedium
	ButtonLarge
)

// CreateBorderedInput creates a text input field with a border
func CreateBorderedInput(gtx layout.Context, th *material.Theme, editor *widget.Editor, hint string) layout.Dimensions {
	border := widget.Border{
		Color:        th.Palette.Fg,
		Width:        unit.Dp(1),
		CornerRadius: StandardCornerRadius,
	}
	return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(6),
			Bottom: unit.Dp(6),
			Left:   StandardPadding,
			Right:  StandardPadding,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			ed := material.Editor(th, editor, hint)
			ed.TextSize = BodySize
			return ed.Layout(gtx)
		})
	})
}

// CreateLabel creates a text label with the specified size
func CreateLabel(gtx layout.Context, th *material.Theme, text string, size unit.Sp) layout.Dimensions {
	label := material.Body2(th, text)
	label.TextSize = size
	return label.Layout(gtx)
}

// CreateSectionTitle creates a section title using H6 style
func CreateSectionTitle(gtx layout.Context, th *material.Theme, title string) layout.Dimensions {
	label := material.H6(th, title)
	return label.Layout(gtx)
}

// CreateButton creates a button with consistent styling based on size
func CreateButton(gtx layout.Context, th *material.Theme, btn *widget.Clickable, text string, size ButtonSize) layout.Dimensions {
	button := material.Button(th, btn, text)
	button.CornerRadius = ButtonCornerRadius

	switch size {
	case ButtonSmall:
		button.TextSize = unit.Sp(13)
		button.Inset = layout.Inset{
			Top:    unit.Dp(6),
			Bottom: unit.Dp(6),
			Left:   unit.Dp(12),
			Right:  unit.Dp(12),
		}
	case ButtonMedium:
		button.TextSize = BodySize
		button.Inset = layout.Inset{
			Top:    StandardPadding,
			Bottom: StandardPadding,
			Left:   LargePadding,
			Right:  LargePadding,
		}
	case ButtonLarge:
		button.TextSize = TitleSize
		button.Inset = layout.Inset{
			Top:    unit.Dp(10),
			Bottom: unit.Dp(10),
			Left:   unit.Dp(24),
			Right:  unit.Dp(24),
		}
	}

	return button.Layout(gtx)
}

// StandardInset returns a standard layout.Inset
func StandardInset() layout.Inset {
	return layout.Inset{
		Top:    StandardPadding,
		Bottom: StandardPadding,
		Left:   StandardPadding,
		Right:  StandardPadding,
	}
}

// SmallInset returns a small layout.Inset
func SmallInset() layout.Inset {
	return layout.Inset{
		Top:    SmallPadding,
		Bottom: SmallPadding,
		Left:   SmallPadding,
		Right:  SmallPadding,
	}
}

// LargeInset returns a large layout.Inset
func LargeInset() layout.Inset {
	return layout.Inset{
		Top:    LargePadding,
		Bottom: LargePadding,
		Left:   LargePadding,
		Right:  LargePadding,
	}
}
