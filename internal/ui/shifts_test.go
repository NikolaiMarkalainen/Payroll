package ui

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func TestShiftsTabHasMonthGrid(t *testing.T) {
	test.NewApp()

	s := newShiftsTab()
	content := s.canvas()
	w := test.NewWindow(content)
	defer w.Close()

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

func TestShiftsTabShowsShiftTimes(t *testing.T) {
	test.NewApp()

	s := newShiftsTab()
	_ = s.canvas()
	day := time.Date(s.month.Year(), s.month.Month(), 15, 0, 0, 0, 0, s.month.Location())
	s.shifts = []calendarShift{{
		Date:  day,
		Start: "06:00",
		End:   "14:00",
	}}
	s.refresh()

	found := s.shiftsOn(day)
	if len(found) != 1 || found[0].Start != "06:00" || found[0].End != "14:00" {
		t.Fatalf("shiftsOn=%+v", found)
	}
	if len(s.grid.Objects) == 0 {
		t.Fatal("no cells after refresh")
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
	if finnishMonthName(time.January) != "tammikuu" {
		t.Fatalf("january=%q", finnishMonthName(time.January))
	}
}

func TestShiftsTabInMainUI(t *testing.T) {
	test.NewApp()

	_, tabs := buildUI()
	w := test.NewWindow(tabs)
	defer w.Close()

	tabs.SelectIndex(1)
	if tabs.Selected().Text != "Vuorot" {
		t.Fatalf("selected=%q", tabs.Selected().Text)
	}
}
