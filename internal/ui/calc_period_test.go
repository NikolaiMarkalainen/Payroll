package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"payroll/internal/calc"
)

func TestCalcTabPeriodSelectFillsRange(t *testing.T) {
	test.NewApp()
	settings := newSettingsTab()
	_ = settings.canvas()
	settings.periodAnchor.SetText("29.06.2026")

	shifts := newShiftsTab(test.NewWindow(nil))
	c := newCalcTab(settings, shifts)
	_ = c.canvas()

	c.periodAnchor.SetText("29.06.2026")
	c.refreshPeriodOptions()
	if len(c.periodOpts) == 0 {
		t.Fatal("no period options")
	}
	// Pick period #1 (index 0)
	var first periodOpt
	for _, o := range c.periodOpts {
		if o.index == 0 {
			first = o
			break
		}
	}
	if first.label == "" {
		t.Fatal("missing index 0")
	}
	c.applyPeriodSelection(first.label)
	wantFrom, wantTo := calc.PeriodIndexWindow(time.Date(2026, 6, 29, 0, 0, 0, 0, time.Local), 0)
	if c.from.Text != formatFIDate(wantFrom) || c.to.Text != formatFIDate(wantTo) {
		t.Fatalf("range=%s-%s want %s-%s", c.from.Text, c.to.Text, formatFIDate(wantFrom), formatFIDate(wantTo))
	}
}
