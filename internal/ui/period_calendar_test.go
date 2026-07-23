package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"payroll/internal/calc"
)

func TestPeriodYearAnchorAndNumber(t *testing.T) {
	loc := time.Local
	// J1 for 2026: 12.01 - 01.02
	d1 := time.Date(2026, 1, 12, 0, 0, 0, 0, loc)
	d2 := time.Date(2026, 2, 1, 0, 0, 0, 0, loc)
	d3 := time.Date(2026, 2, 2, 0, 0, 0, 0, loc)
	if periodYearNumber(d1) != 1 || periodYearNumber(d2) != 1 {
		t.Fatalf("J1 days: %d %d", periodYearNumber(d1), periodYearNumber(d2))
	}
	if periodYearNumber(d3) != 2 {
		t.Fatalf("J2 start=%d", periodYearNumber(d3))
	}
	// Before 12.01.2026 → previous year grid (still positive)
	early := time.Date(2026, 1, 5, 0, 0, 0, 0, loc)
	n := periodYearNumber(early)
	if n < 1 {
		t.Fatalf("early jan must be positive, got %d", n)
	}
	anchor := periodYearAnchorFor(early)
	wantAnchor := time.Date(2025, 1, 12, 0, 0, 0, 0, loc)
	if !anchor.Equal(wantAnchor) {
		t.Fatalf("anchor=%v want %v", anchor, wantAnchor)
	}
	from, to := calc.PeriodIndexWindow(periodYearAnchorFor(d1), 0)
	if !from.Equal(d1) || !to.Equal(d2) {
		t.Fatalf("J1 window %s-%s", from.Format("2.1.2006"), to.Format("2.1.2006"))
	}
}

func TestPeriodBadgeColorStablePerIndex(t *testing.T) {
	a := periodBadgeColor(1)
	b := periodBadgeColor(2)
	if a == b {
		t.Fatal("J1 and J2 should differ")
	}
	if periodBadgeColor(1) != periodBadgeColor(7) {
		t.Fatal("wrap should reuse color")
	}
}

func TestShiftsPeriodHighlightYearNumbers(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newShiftsTab(w)
	yearAnchor := time.Date(2026, 1, 12, 0, 0, 0, 0, time.Local)
	s.periodThreshold = func(day time.Time) (float64, bool) {
		return calc.PeriodThresholdAt(yearAnchor, day, calc.PeriodThreshold120), true
	}
	selFrom, selTo := calc.PeriodIndexWindow(yearAnchor, 0) // J1
	s.selectedPeriod = func() (time.Time, time.Time, bool) {
		return selFrom, selTo, true
	}

	dayInJ1 := time.Date(2026, 1, 20, 0, 0, 0, 0, time.Local)
	num, label, selected, ok := s.periodInfo(dayInJ1)
	if !ok || num != 1 || !selected || label != "J1" {
		t.Fatalf("J1 mid: num=%d label=%q selected=%v", num, label, selected)
	}

	_, startLabel, _, ok := s.periodInfo(selFrom)
	if !ok || startLabel != "J1 / 120h" {
		t.Fatalf("J1 start label=%q", startLabel)
	}

	dayInJ2 := time.Date(2026, 2, 5, 0, 0, 0, 0, time.Local)
	num2, label2, sel2, ok := s.periodInfo(dayInJ2)
	if !ok || num2 != 2 || sel2 || label2 != "J2" {
		t.Fatalf("J2: num=%d label=%q selected=%v", num2, label2, sel2)
	}
}
