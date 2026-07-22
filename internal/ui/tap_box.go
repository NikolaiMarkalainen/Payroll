package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// tapBox makes arbitrary content clickable, with hover/press feedback.
type tapBox struct {
	widget.BaseWidget
	content   fyne.CanvasObject
	bg        *canvas.Rectangle
	onTap     func()
	minWidth  float32
	minHeight float32
	hovered   bool
	pressed   bool
}

func newTapBox(content fyne.CanvasObject, onTap func()) *tapBox {
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = 4
	t := &tapBox{
		content: container.NewStack(bg, content),
		bg:      bg,
		onTap:   onTap,
	}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tapBox) withMinSize(w, h float32) *tapBox {
	t.minWidth = w
	t.minHeight = h
	return t
}

func (t *tapBox) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.content)
}

func (t *tapBox) Tapped(_ *fyne.PointEvent) {
	t.pressed = false
	t.refreshBg()
	if t.onTap != nil {
		t.onTap()
	}
}

func (t *tapBox) TappedSecondary(_ *fyne.PointEvent) {}

func (t *tapBox) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (t *tapBox) MouseIn(_ *desktop.MouseEvent) {
	t.hovered = true
	t.refreshBg()
}

func (t *tapBox) MouseOut() {
	t.hovered = false
	t.pressed = false
	t.refreshBg()
}

func (t *tapBox) MouseMoved(_ *desktop.MouseEvent) {}

func (t *tapBox) MouseDown(_ *desktop.MouseEvent) {
	t.pressed = true
	t.refreshBg()
}

func (t *tapBox) MouseUp(_ *desktop.MouseEvent) {
	t.pressed = false
	t.refreshBg()
}

func (t *tapBox) refreshBg() {
	if t.bg == nil {
		return
	}
	switch {
	case t.pressed:
		t.bg.FillColor = theme.Color(theme.ColorNamePressed)
	case t.hovered:
		t.bg.FillColor = theme.Color(theme.ColorNameHover)
	default:
		t.bg.FillColor = color.Transparent
	}
	t.bg.Refresh()
}

func (t *tapBox) MinSize() fyne.Size {
	min := fyne.NewSize(0, 0)
	if t.content != nil {
		min = t.content.MinSize()
	}
	if t.minWidth > min.Width {
		min.Width = t.minWidth
	}
	if t.minHeight > min.Height {
		min.Height = t.minHeight
	}
	return min
}
