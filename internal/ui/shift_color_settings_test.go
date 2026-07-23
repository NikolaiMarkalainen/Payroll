package ui

import (
	"image/color"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func TestCollectShiftColorGroupsFromCalendar(t *testing.T) {
	day := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	groups := collectShiftColorGroups([]calendarShift{
		{Date: day, Code: "3AAA"},
		{Date: day, Code: "3AAB"},
		{Date: day, Code: "2BBB"},
		{Date: day, Code: "ALPHA"},
		{Date: day, Code: ""}, // ignored
	})
	if len(groups) != 3 {
		t.Fatalf("groups=%d want 3: %+v", len(groups), groups)
	}
	if groups[0].Key != "2" || len(groups[0].Codes) != 1 || groups[0].Codes[0] != "2BBB" {
		t.Fatalf("first=%+v", groups[0])
	}
	if groups[1].Key != "3" || len(groups[1].Codes) != 2 {
		t.Fatalf("second=%+v", groups[1])
	}
	if groups[2].Key != "ALPHA" {
		t.Fatalf("third=%+v", groups[2])
	}
	label := shiftColorGroupLabel(groups[1])
	if label != "Ryhmä 3 (*3AAA, *3AAB)" {
		t.Fatalf("label=%q", label)
	}
}

func TestSettingsShiftColorRowsFollowCalendar(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	shifts := newShiftsTab(w)
	s := newSettingsTab()
	s.window = w
	s.shiftsSource = func() []calendarShift { return shifts.shifts }
	_ = s.canvas()

	if len(s.colorRows.Objects) != 0 {
		t.Fatalf("empty calendar should have 0 color rows, got %d", len(s.colorRows.Objects))
	}

	day := time.Date(shifts.month.Year(), shifts.month.Month(), 10, 0, 0, 0, 0, shifts.month.Location())
	if err := shifts.addShift(calendarShift{Date: day, Start: "06:00", End: "14:00", Code: "3AAA"}); err != nil {
		t.Fatal(err)
	}
	if err := shifts.addShift(calendarShift{Date: day.AddDate(0, 0, 1), Start: "06:00", End: "14:00", Code: "3AAB"}); err != nil {
		t.Fatal(err)
	}
	s.refreshColorRows()
	if len(s.colorRows.Objects) != 1 {
		t.Fatalf("rows=%d want 1 group for 3*", len(s.colorRows.Objects))
	}

	s.setShiftColor("3", color.NRGBA{R: 1, G: 2, B: 3, A: 255})
	if s.colorForShiftCode("3AAB") != (color.NRGBA{R: 1, G: 2, B: 3, A: 255}) {
		t.Fatal("override not applied to group")
	}
}

func TestSettingsManualShiftColorAdd(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newSettingsTab()
	s.window = w
	_ = s.canvas()

	if !s.addManualColorKey("4") {
		t.Fatal("add failed")
	}
	if len(s.colorRows.Objects) != 1 {
		t.Fatalf("rows=%d want 1", len(s.colorRows.Objects))
	}
	groups := s.allShiftColorGroups()
	if len(groups) != 1 || groups[0].Key != "4" || !groups[0].ManualOnly {
		t.Fatalf("groups=%+v", groups)
	}
	s.setShiftColor("4", color.NRGBA{R: 9, G: 8, B: 7, A: 255})
	if s.colorForShiftCode("4ZZZ") != (color.NRGBA{R: 9, G: 8, B: 7, A: 255}) {
		t.Fatal("manual group color not used for matching codes")
	}
	if !s.addManualColorKey("ALPHA") {
		t.Fatal("add letter key failed")
	}
	if len(s.allShiftColorGroups()) != 2 {
		t.Fatalf("want 2 groups")
	}
	s.removeManualColorKey("4")
	got := s.allShiftColorGroups()
	if len(got) != 1 || got[0].Key != "ALPHA" {
		t.Fatalf("after remove=%+v", got)
	}
}
