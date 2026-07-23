package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// optionSelect is a dropdown where the popup lists only the other choices -
// the current selection stays on the button (not repeated in the menu).
type optionSelect struct {
	widget.BaseWidget

	Options []string
	Selected string
	PlaceHolder string
	OnChanged func(string)

	button *widget.Button
	popUp *widget.PopUpMenu
}

func newOptionSelect(options []string, changed func(string)) *optionSelect {
	o := &optionSelect{
		Options: append([]string(nil), options...),
		PlaceHolder: "(Valitse)",
		OnChanged: changed,
	}
	o.ExtendBaseWidget(o)
	o.button = widget.NewButtonWithIcon(o.displayText(), theme.MenuDropDownIcon(), func() {
		o.showMenu()
	})
	o.button.Alignment = widget.ButtonAlignLeading
	return o
}

func (o *optionSelect) CreateRenderer() fyne.WidgetRenderer {
	o.ExtendBaseWidget(o)
	return widget.NewSimpleRenderer(o.button)
}

func (o *optionSelect) displayText() string {
	if o.Selected != "" {
		return o.Selected
	}
	if o.PlaceHolder != "" {
		return o.PlaceHolder
	}
	return "(Valitse)"
}

func (o *optionSelect) SetSelected(text string) {
	if text != "" && !containsString(o.Options, text) {
		return
	}
	prev := o.Selected
	o.Selected = text
	if o.button != nil {
		o.button.SetText(o.displayText())
	}
	if text != prev && o.OnChanged != nil {
		o.OnChanged(text)
	}
	o.Refresh()
}

// forceSelected sets the value without OnChanged and without requiring it
// to already be in Options (used when restoring persisted TES state).
func (o *optionSelect) forceSelected(text string) {
	if o == nil {
		return
	}
	o.Selected = text
	if o.button != nil {
		o.button.SetText(o.displayText())
	}
	o.Refresh()
}

func (o *optionSelect) SetOptions(options []string) {
	o.Options = append([]string(nil), options...)
	o.Refresh()
}

func (o *optionSelect) Refresh() {
	if o.button != nil {
		o.button.SetText(o.displayText())
		o.button.Refresh()
	}
	o.BaseWidget.Refresh()
}

func (o *optionSelect) showMenu() {
	if o.popUp != nil {
		o.popUp.Hide()
		o.popUp = nil
	}
	var items []*fyne.MenuItem
	for _, opt := range o.Options {
		if opt == o.Selected {
			continue // selected already on the button - do not repeat in menu
		}
		text := opt
		items = append(items, fyne.NewMenuItem(text, func() {
			o.SetSelected(text)
			o.popUp = nil
		}))
	}
	if len(items) == 0 {
		return
	}
	c := fyne.CurrentApp().Driver().CanvasForObject(o)
	if c == nil {
		c = fyne.CurrentApp().Driver().CanvasForObject(o.button)
	}
	if c == nil {
		return
	}
	pop := widget.NewPopUpMenu(fyne.NewMenu("", items...), c)
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(o)
	pop.ShowAtPosition(fyne.NewPos(pos.X, pos.Y+o.Size().Height))
	pop.Resize(fyne.NewSize(o.Size().Width, pop.MinSize().Height))
	pop.OnDismiss = func() {
		pop.Hide()
		if o.popUp == pop {
			o.popUp = nil
		}
	}
	o.popUp = pop
}
