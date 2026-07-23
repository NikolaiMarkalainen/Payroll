package ui

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func TestShiftsTabHasMonthGrid(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newShiftsTab(w)
	content := s.canvas()
	w.SetContent(content)

	if s.grid == nil {
		t.Fatal("grid missing")
	}
	if len(s.grid.Objects) == 0 {
		t.Fatal("expected calendar cells")
	}
	if len(s.grid.Objects)%7 != 0 {
		t.Fatalf("cells=%d not multiple of 7", len(s.grid.Objects))
	}
	wantTitle := finnishMonthName(s.month.Month())
	if !strings.Contains(s.title.Text, wantTitle) {
		t.Fatalf("title=%q want month %q", s.title.Text, wantTitle)
	}
}

func TestAddShiftShowsInCell(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newShiftsTab(w)
	_ = s.canvas()
	day := time.Date(s.month.Year(), s.month.Month(), 15, 0, 0, 0, 0, s.month.Location())
	if err := s.addShift(calendarShift{
		Date: day,
		Start: "06:00",
		End: "14:00",
		Callout: true,
	}); err != nil {
		t.Fatal(err)
	}

	found := s.shiftsOn(day)
	if len(found) != 1 {
		t.Fatalf("shifts=%d", len(found))
	}
	if found[0].Start != "06:00" || found[0].End != "14:00" || !found[0].Callout {
		t.Fatalf("shift=%+v", found[0])
	}
	if found[0].ID == 0 {
		t.Fatal("expected shift id")
	}
}

func TestRemoveShift(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newShiftsTab(w)
	_ = s.canvas()
	day := time.Date(s.month.Year(), s.month.Month(), 10, 0, 0, 0, 0, s.month.Location())
	if err := s.addShift(calendarShift{Date: day, Start: "06:00", End: "14:00"}); err != nil {
		t.Fatal(err)
	}
	if err := s.addShift(calendarShift{Date: day, Start: "18:00", End: "22:00"}); err != nil {
		t.Fatal(err)
	}
	if len(s.shiftsOn(day)) != 2 {
		t.Fatalf("want 2 shifts, got %d", len(s.shiftsOn(day)))
	}
	id := s.shiftsOn(day)[0].ID
	s.removeShift(id)
	left := s.shiftsOn(day)
	if len(left) != 1 || left[0].Start != "18:00" {
		t.Fatalf("after remove: %+v", left)
	}
}

func TestUpdateShift(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newShiftsTab(w)
	_ = s.canvas()
	day := time.Date(s.month.Year(), s.month.Month(), 12, 0, 0, 0, 0, s.month.Location())
	if err := s.addShift(calendarShift{Date: day, Start: "06:00", End: "14:00"}); err != nil {
		t.Fatal(err)
	}
	id := s.shiftsOn(day)[0].ID
	if err := s.updateShift(calendarShift{
		ID: id, Date: day, Start: "07:00", End: "15:00", Callout: true,
	}); err != nil {
		t.Fatal(err)
	}
	got := s.shiftsOn(day)[0]
	if got.Start != "07:00" || got.End != "15:00" || !got.Callout {
		t.Fatalf("updated=%+v", got)
	}
}

func TestUpdateShiftRejectsOverlap(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newShiftsTab(w)
	_ = s.canvas()
	day := time.Date(s.month.Year(), s.month.Month(), 12, 0, 0, 0, 0, s.month.Location())
	if err := s.addShift(calendarShift{Date: day, Start: "06:00", End: "14:00"}); err != nil {
		t.Fatal(err)
	}
	if err := s.addShift(calendarShift{Date: day, Start: "18:00", End: "22:00"}); err != nil {
		t.Fatal(err)
	}
	id := s.shiftsOn(day)[1].ID
	err := s.updateShift(calendarShift{ID: id, Date: day, Start: "10:00", End: "12:00"})
	if err == nil {
		t.Fatal("expected overlap on edit")
	}
}

func TestDayCellIsTappableForAdd(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newShiftsTab(w)
	_ = s.canvas()

	foundTap := false
	var minH float32
	for _, obj := range s.grid.Objects {
		if tb, ok := obj.(*tapBox); ok {
			foundTap = true
			if tb.MinSize().Height > minH {
				minH = tb.MinSize().Height
			}
		}
	}
	if !foundTap {
		t.Fatal("expected tappable day cell")
	}
	if minH < 96 {
		t.Fatalf("day cell min height=%v want >= 96", minH)
	}
}

func TestFinnishWeekdaysAndMonths(t *testing.T) {
	want := []string{"Ma", "Ti", "Ke", "To", "Pe", "La", "Su"}
	for i := range want {
		if finnishWeekdays[i] != want[i] {
			t.Fatalf("weekday[%d]=%q", i, finnishWeekdays[i])
		}
	}
	if finnishMonthName(time.July) != "heinäkuu" {
		t.Fatalf("july=%q", finnishMonthName(time.July))
	}
}

func TestShiftsTabInMainUI(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	content, tabs, _, _, _ := buildUI(w)
	w.SetContent(content)

	tabs.SelectIndex(1)
	if tabs.Selected().Text != "Vuorot" {
		t.Fatalf("selected=%q", tabs.Selected().Text)
	}
}
