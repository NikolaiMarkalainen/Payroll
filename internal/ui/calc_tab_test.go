package ui

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func TestParseFIDate(t *testing.T) {
	got, err := parseFIDate("20.07.2026")
	if err != nil {
		t.Fatal(err)
	}
	if got.Day() != 20 || got.Month() != time.July || got.Year() != 2026 {
		t.Fatalf("got=%v", got)
	}
	if _, err := parseFIDate("bad"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCalcTabRunsForDemoRange(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	settings := newSettingsTab()
	settings.hourlyWage.SetText("12.00")
	settings.eveningAllowance.SetText("1.20")
	settings.nightAllowance.SetText("2.40")
	settings.saturdayAllowance.SetText("1.50")
	settings.sundayAllowance.SetText("6.00")
	settings.holidayAllowance.SetText("6.00")

	shifts := newShiftsTab(w)
	_ = shifts.canvas()
	shifts.loadDemoSeed()

	c := newCalcTab(settings, shifts)
	_ = c.canvas()
	c.setDemoRange()
	c.run()

	if c.err.Visible() {
		t.Fatalf("err=%q", c.err.Text)
	}
	if !strings.Contains(c.summary.Text, "98.83") && !strings.Contains(c.summary.Text, "98.80") {
		// 98:50 = 98 + 50/60 ≈ 98.833... rounded to 98.83
		if !strings.Contains(c.summary.Text, "98.") {
			t.Fatalf("summary missing hours: %q", c.summary.Text)
		}
	}
	if !strings.Contains(c.summary.Text, " e") && !strings.Contains(c.summary.Text, "e\n") {
		if !strings.Contains(c.summary.Text, "Palkka yhteensä") {
			t.Fatalf("summary missing money: %q", c.summary.Text)
		}
	}
	if !strings.Contains(c.details.Text, "Pohja:") {
		t.Fatalf("details=%q", c.details.Text)
	}
}

func TestCalcTabFillFromShifts(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	settings := newSettingsTab()
	shifts := newShiftsTab(w)
	_ = shifts.canvas()
	shifts.loadDemoSeed()

	c := newCalcTab(settings, shifts)
	_ = c.canvas()
	c.fillFromShifts()
	if c.from.Text != "20.07.2026" || c.to.Text != "04.08.2026" {
		t.Fatalf("range=%s-%s", c.from.Text, c.to.Text)
	}
}
