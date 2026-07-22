package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func TestOvernightSegmentsSplitAcrossDays(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newShiftsTab(w)
	_ = s.canvas()
	day := time.Date(s.month.Year(), s.month.Month(), 10, 0, 0, 0, 0, s.month.Location())
	next := day.AddDate(0, 0, 1)

	if err := s.addShift(calendarShift{Date: day, Start: "22:00", End: "06:00"}); err != nil {
		t.Fatal(err)
	}

	startSegs := s.segmentsOn(day)
	if len(startSegs) != 1 || startSegs[0].Label != "22:00–24:00" || !startSegs[0].Continues {
		t.Fatalf("start day segments=%+v", startSegs)
	}
	endSegs := s.segmentsOn(next)
	if len(endSegs) != 1 || endSegs[0].Label != "00:00–06:00" || !endSegs[0].Continuation {
		t.Fatalf("next day segments=%+v", endSegs)
	}
	if startSegs[0].Shift.ID != endSegs[0].Shift.ID {
		t.Fatal("segments should share shift id")
	}
}

func TestOverlapSameDayRejected(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newShiftsTab(w)
	_ = s.canvas()
	day := time.Date(s.month.Year(), s.month.Month(), 12, 0, 0, 0, 0, s.month.Location())

	if err := s.addShift(calendarShift{Date: day, Start: "06:00", End: "14:00"}); err != nil {
		t.Fatal(err)
	}
	err := s.addShift(calendarShift{Date: day, Start: "10:00", End: "12:00"})
	if err == nil {
		t.Fatal("expected overlap error")
	}
	if len(s.shiftsOn(day)) != 1 {
		t.Fatalf("shifts=%d", len(s.shiftsOn(day)))
	}
}

func TestAdjacentShiftsAllowed(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newShiftsTab(w)
	_ = s.canvas()
	day := time.Date(s.month.Year(), s.month.Month(), 12, 0, 0, 0, 0, s.month.Location())

	if err := s.addShift(calendarShift{Date: day, Start: "06:00", End: "14:00"}); err != nil {
		t.Fatal(err)
	}
	if err := s.addShift(calendarShift{Date: day, Start: "14:00", End: "22:00"}); err != nil {
		t.Fatal(err)
	}
	if len(s.shiftsOn(day)) != 2 {
		t.Fatalf("shifts=%d", len(s.shiftsOn(day)))
	}
}

func TestOvernightOverlapsNextMorning(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newShiftsTab(w)
	_ = s.canvas()
	day := time.Date(s.month.Year(), s.month.Month(), 15, 0, 0, 0, 0, s.month.Location())
	next := day.AddDate(0, 0, 1)

	if err := s.addShift(calendarShift{Date: day, Start: "22:00", End: "06:00"}); err != nil {
		t.Fatal(err)
	}
	err := s.addShift(calendarShift{Date: next, Start: "05:00", End: "12:00"})
	if err == nil {
		t.Fatal("expected overlap with overnight tail")
	}
	// Touching at 06:00 is OK.
	if err := s.addShift(calendarShift{Date: next, Start: "06:00", End: "12:00"}); err != nil {
		t.Fatal(err)
	}
}

func TestAbsoluteRangeOvernight(t *testing.T) {
	day := time.Date(2026, 7, 10, 0, 0, 0, 0, time.Local)
	sh := calendarShift{Date: day, Start: "22:00", End: "06:00"}
	start, end, err := sh.absoluteRange()
	if err != nil {
		t.Fatal(err)
	}
	if start.Hour() != 22 || start.Day() != 10 {
		t.Fatalf("start=%v", start)
	}
	if end.Hour() != 6 || end.Day() != 11 {
		t.Fatalf("end=%v", end)
	}
}
