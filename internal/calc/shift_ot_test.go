package calc

import (
	"math"
	"testing"
	"time"
)

func TestShiftExtensionExcludedFromPeriodOT(t *testing.T) {
	// 130h worked of which 10h are shift-extension OT → period sees 120 → no period OT at 120 threshold.
	loc := time.UTC
	from := time.Date(2026, 7, 6, 0, 0, 0, 0, loc)
	to := from.AddDate(0, 0, 20)

	var shifts []Shift
	// 12 days * 10h = 120h (no extension) + one 22h day = 10 extension → 142 total worked?
	// Better: 11*10 + one 22 = 110+22=132; extension=10; period base=122 → 2 OT over 120.
	for i := 0; i < 11; i++ {
		day := from.AddDate(0, 0, i)
		shifts = append(shifts, Shift{Start: at(day, 6, 0), End: at(day, 16, 0)})
	}
	long := from.AddDate(0, 0, 11)
	shifts = append(shifts, Shift{Start: at(long, 0, 0), End: at(long, 22, 0)}) // 22h → 10 ext

	rules := DefaultRules()
	rules.PeriodOTEnabled = true
	rules.PeriodThresholdH = PeriodThreshold120
	rules.PeriodOT50AfterH = 18

	got := Calculate(PeriodInput{
		From: from, To: to,
		Shifts: shifts,
		Rates:  Rates{Hourly: 10},
		Rules:  rules,
	})
	if got.BaseHours != 132 {
		t.Fatalf("base=%v", got.BaseHours)
	}
	if got.ShiftOTHours != 10 {
		t.Fatalf("shift ot=%v", got.ShiftOTHours)
	}
	if got.PeriodWorkedHours != 122 {
		t.Fatalf("period worked=%v want 122", got.PeriodWorkedHours)
	}
	if got.PeriodOT50Hours != 2 || got.PeriodOT100Hours != 0 {
		t.Fatalf("period ot=%v/%v", got.PeriodOT50Hours, got.PeriodOT100Hours)
	}
	if got.Overtime50Hours != 10 {
		t.Fatalf("ext ot50=%v", got.Overtime50Hours)
	}
	_ = math.Abs
}

func TestCollectShiftExtensionOvernight(t *testing.T) {
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	next := day.AddDate(0, 0, 1)
	// 18:00–08:00 = 14h → 2h extension on next morning 06–08.
	ext := collectShiftExtensionHours(
		[]Shift{{Start: at(day, 18, 0), End: at(next, 8, 0)}},
		day, next.Add(24*time.Hour), 12,
	)
	if math.Abs(ext.total-2) > 0.001 {
		t.Fatalf("total=%v", ext.total)
	}
	if len(ext.slices) != 1 {
		t.Fatalf("slices=%d", len(ext.slices))
	}
	if ext.slices[0].start.Hour() != 6 || ext.slices[0].end.Hour() != 8 {
		t.Fatalf("slice=%v–%v", ext.slices[0].start, ext.slices[0].end)
	}
}
