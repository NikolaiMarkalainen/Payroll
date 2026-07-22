package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestTapBoxHoverAndPress(t *testing.T) {
	test.NewApp()

	tb := newTapBox(widget.NewLabel("x"), func() {})
	tb.MouseIn(nil)
	if !tb.hovered {
		t.Fatal("expected hovered after MouseIn")
	}
	tb.MouseDown(nil)
	if !tb.pressed {
		t.Fatal("expected pressed after MouseDown")
	}
	tb.MouseUp(nil)
	if tb.pressed {
		t.Fatal("expected not pressed after MouseUp")
	}
	tb.MouseOut()
	if tb.hovered || tb.pressed {
		t.Fatal("expected clear state after MouseOut")
	}
}
