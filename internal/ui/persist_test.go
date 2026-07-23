package ui

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func TestPersistRoundTripSettingsShiftsCalc(t *testing.T) {
	test.NewApp()
	testDir := t.TempDir()

	s := newSettingsTab()
	_ = s.canvas()
	s.applyVartiointiPay("Taso IV", true, tesServicePerus)
	s.trainingEnabled.SetChecked(true)
	s.hourlyWage.SetText("13.97")
	s.perehdytysAllowance.SetText("1.00")
	s.colorShiftTitles.SetChecked(true)
	s.initShiftColors()
	s.shiftColorOverrides["3"] = colorNRGBA(0x11, 0x22, 0x33)
	s.shiftColorManual["X"] = struct{}{}

	w := newShiftsTab(nil)
	day := time.Date(2026, 7, 3, 0, 0, 0, 0, time.Local)
	if err := w.addShift(calendarShift{
		Date: day, Start: "06:00", End: "14:00", Code: "3AAA",
		PerehdytysStart: "06:00", PerehdytysEnd: "07:00",
	}); err != nil {
		t.Fatal(err)
	}

	c := newCalcTab(s, w)
	_ = c.canvas()
	c.setRange(day, day.AddDate(0, 0, 20))
	c.absence.SetText("2.5")

	p := newAppPersister(testDir, s, w, c)
	if err := p.saveNow(); err != nil {
		t.Fatal(err)
	}
	path := statePath(testDir)
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}

	s2 := newSettingsTab()
	_ = s2.canvas()
	w2 := newShiftsTab(nil)
	c2 := newCalcTab(s2, w2)
	_ = c2.canvas()
	p2 := newAppPersister(testDir, s2, w2, c2)
	if err := p2.load(); err != nil {
		t.Fatal(err)
	}

	if s2.tesFamily.Selected != tesFamilyVartio {
		t.Fatalf("family=%q", s2.tesFamily.Selected)
	}
	if s2.hourlyWage.Text != "13.97" {
		t.Fatalf("hourly=%q", s2.hourlyWage.Text)
	}
	if s2.perehdytysAllowance.Text != "1.00" {
		t.Fatalf("pere=%q", s2.perehdytysAllowance.Text)
	}
	if !s2.trainingEnabled.Checked {
		t.Fatal("training not restored")
	}
	if len(w2.shifts) != 1 {
		t.Fatalf("shifts=%d", len(w2.shifts))
	}
	got := w2.shifts[0]
	if got.Code != "3AAA" || got.PerehdytysStart != "06:00" || !got.Date.Equal(day) {
		t.Fatalf("shift=%+v", got)
	}
	if c2.from.Text != formatFIDate(day) || c2.absence.Text != "2.5" {
		t.Fatalf("calc from=%q absence=%q", c2.from.Text, c2.absence.Text)
	}
	if hex, ok := s2.shiftColorOverrides["3"]; !ok || colorToHex(hex) != "#112233" {
		t.Fatalf("color overrides=%v", s2.shiftColorOverrides)
	}
}

func colorNRGBA(r, g, b uint8) color.NRGBA {
	return color.NRGBA{R: r, G: g, B: b, A: 0xff}
}

func TestLoadMissingStateIsNil(t *testing.T) {
	st, err := loadPersistedState(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatal(err)
	}
	if st != nil {
		t.Fatalf("want nil, got %+v", st)
	}
}
