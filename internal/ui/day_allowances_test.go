package ui

import (
	"image/color"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func TestSummarizeDayEveningAndNight(t *testing.T) {
	rules := defaultAllowanceRules()
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC) // Wednesday

	shifts := []calendarShift{{
		Date: day, Start: "18:00", End: "22:00",
	}}
	got := summarizeDay(day, shifts, rules)
	if got.Evening != 4 || got.Night != 0 {
		t.Fatalf("evening/night=%v/%v", got.Evening, got.Night)
	}

	shifts = []calendarShift{{
		Date: day, Start: "20:00", End: "23:00",
	}}
	got = summarizeDay(day, shifts, rules)
	if got.Evening != 2 || got.Night != 1 {
		t.Fatalf("mixed evening/night=%v/%v want 2/1", got.Evening, got.Night)
	}
}

func TestSummarizeDayOvernightNightSplit(t *testing.T) {
	rules := defaultAllowanceRules()
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	next := day.AddDate(0, 0, 1)
	shifts := []calendarShift{{
		Date: day, Start: "22:00", End: "06:00",
	}}

	startDay := summarizeDay(day, shifts, rules)
	if startDay.Night != 2 || startDay.Total != 2 {
		t.Fatalf("start day=%+v", startDay)
	}
	endDay := summarizeDay(next, shifts, rules)
	if endDay.Night != 6 || endDay.Total != 6 {
		t.Fatalf("end day=%+v", endDay)
	}
	chips := endDay.chips()
	if len(chips) == 0 || chips[0].Code != codeNight {
		t.Fatalf("chips=%+v", chips)
	}
}

func TestSummarizeDaySundaySaturdayHoliday(t *testing.T) {
	rules := defaultAllowanceRules()
	sunday := time.Date(2026, 7, 19, 0, 0, 0, 0, time.UTC)
	saturday := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)
	holiday := time.Date(2026, 12, 6, 0, 0, 0, 0, time.UTC) // Sunday independence day 2026? Dec 6 2026 is Sunday
	rules.holidays = map[string]bool{holidayKey(holiday): true}

	sShifts := []calendarShift{{Date: sunday, Start: "06:00", End: "14:00"}}
	got := summarizeDay(sunday, sShifts, rules)
	if got.Sunday != 8 || got.Saturday != 0 {
		t.Fatalf("sunday=%+v", got)
	}

	lShifts := []calendarShift{{Date: saturday, Start: "06:00", End: "14:00"}}
	got = summarizeDay(saturday, lShifts, rules)
	if got.Saturday != 8 {
		t.Fatalf("saturday=%+v", got)
	}

	pShifts := []calendarShift{{Date: holiday, Start: "08:00", End: "12:00"}}
	got = summarizeDay(holiday, pShifts, rules)
	if got.Holiday != 4 || got.Sunday != 4 {
		t.Fatalf("holiday sunday=%+v", got)
	}
}

func TestSummarizeDayCalloutAndOvertime(t *testing.T) {
	rules := defaultAllowanceRules()
	rules.calloutFixedH = 2 // Vartiointi TES 31 §
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	shifts := []calendarShift{{
		Date: day, Start: "06:00", End: "16:00", Callout: true,
	}}
	got := summarizeDay(day, shifts, rules)
	// 10h total -> under 12h shift OT; hälytys = kiinteä 2 h (ei koko vuoro)
	if got.Total != 10 || got.Callout != 2 || got.Overtime50 != 0 || got.Overtime100 != 0 {
		t.Fatalf("got=%+v", got)
	}

	shifts = []calendarShift{{
		Date: day, Start: "06:00", End: "18:00", Callout: true,
	}}
	got = summarizeDay(day, shifts, rules)
	// exactly 12h -> no extension OT
	if got.Total != 12 || got.Callout != 2 || got.Overtime50 != 0 || got.Overtime100 != 0 {
		t.Fatalf("12h day got=%+v", got)
	}

	shifts = []calendarShift{{
		Date: day, Start: "06:00", End: "20:00", Callout: true,
	}}
	got = summarizeDay(day, shifts, rules)
	// 14h -> 2h extension @ 50% chip
	if got.Total != 14 || got.Callout != 2 || got.Overtime50 != 2 || got.Overtime100 != 0 {
		t.Fatalf("14h day got=%+v", got)
	}
}

func TestSplitOvertime50And100(t *testing.T) {
	ot50, ot100 := splitOvertime(12, 8, 10)
	if ot50 != 2 || ot100 != 2 {
		t.Fatalf("12h -> 50%%=%v 100%%=%v want 2/2", ot50, ot100)
	}
	ot50, ot100 = splitOvertime(9, 8, 10)
	if ot50 != 1 || ot100 != 0 {
		t.Fatalf("9h -> %v/%v", ot50, ot100)
	}
	ot50, ot100 = splitOvertime(8, 8, 10)
	if ot50 != 0 || ot100 != 0 {
		t.Fatalf("8h -> %v/%v", ot50, ot100)
	}
}

func TestChipColorsDefined(t *testing.T) {
	for _, code := range []string{
		codeCallout, codeSunday, codeHoliday, codeEvening, codeNight, codeSaturday,
		codeOvertime50, codeOvertime100,
	} {
		c := chipColor(code)
		if c == (color.NRGBA{}) {
			t.Fatalf("missing color for %s", code)
		}
	}
}

func TestSummarizeDayNoOverlapDoubleCountEveningNight(t *testing.T) {
	rules := defaultAllowanceRules()
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	// Exactly evening window end = night start.
	shifts := []calendarShift{{Date: day, Start: "18:00", End: "22:00"}}
	got := summarizeDay(day, shifts, rules)
	if got.Evening+got.Night != got.Total {
		t.Fatalf("eve+night should equal total: %+v", got)
	}
}

func TestSummarizeDayOvernightViaShiftsTab(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newShiftsTab(w)
	_ = s.canvas()
	rules := defaultAllowanceRules()
	rules.calloutFixedH = 2
	s.rules = func() allowanceRules { return rules }
	day := time.Date(s.month.Year(), s.month.Month(), 10, 0, 0, 0, 0, s.month.Location())
	next := day.AddDate(0, 0, 1)

	if err := s.addShift(calendarShift{Date: day, Start: "22:00", End: "06:00", Callout: true}); err != nil {
		t.Fatal(err)
	}

	start := summarizeDay(day, s.shifts, s.currentRules())
	end := summarizeDay(next, s.shifts, s.currentRules())
	if start.Night != 2 || start.Callout != 2 {
		t.Fatalf("start=%+v want night=2 callout=2 (fixed on start day)", start)
	}
	if end.Night != 6 || end.Callout != 0 {
		t.Fatalf("end=%+v want night=6 callout=0 (fixed only on start day)", end)
	}
}
