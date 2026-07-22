package calc

import (
	"math"
	"testing"
	"time"
)

func TestPeriodThresholdAlternating(t *testing.T) {
	anchor := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC) // period 0 start
	if got := PeriodThresholdAt(anchor, anchor, PeriodThreshold128); got != 128 {
		t.Fatalf("period0=%v", got)
	}
	p1 := anchor.AddDate(0, 0, 21)
	if got := PeriodThresholdAt(anchor, p1, PeriodThreshold128); got != 112 {
		t.Fatalf("period1=%v", got)
	}
	p2 := anchor.AddDate(0, 0, 42)
	if got := PeriodThresholdAt(anchor, p2, PeriodThreshold128); got != 128 {
		t.Fatalf("period2=%v", got)
	}
	if got := PeriodThresholdAt(anchor, anchor, PeriodThreshold120); got != 120 {
		t.Fatalf("fixed120=%v", got)
	}
}

func TestPeriodWindow(t *testing.T) {
	anchor := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	from, to := PeriodWindow(anchor, time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC))
	if from.Day() != 6 || to.Day() != 26 {
		t.Fatalf("window=%v–%v", from, to)
	}
	from2, to2 := PeriodWindow(anchor, time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC))
	if from2.Day() != 27 || to2.Month() != time.August || to2.Day() != 16 {
		t.Fatalf("window2=%v–%v", from2, to2)
	}
}

func TestSplitPeriodOvertime(t *testing.T) {
	ot50, ot100 := splitPeriodOvertime(140, 128, 18)
	if math.Abs(ot50-12) > 0.001 || math.Abs(ot100) > 0.001 {
		t.Fatalf("140/128 -> %v/%v want 12/0", ot50, ot100)
	}
	ot50, ot100 = splitPeriodOvertime(160, 128, 18)
	if math.Abs(ot50-18) > 0.001 || math.Abs(ot100-14) > 0.001 {
		t.Fatalf("160/128 -> %v/%v want 18/14", ot50, ot100)
	}
	ot50, ot100 = splitPeriodOvertime(120, 112, 18)
	if math.Abs(ot50-8) > 0.001 || ot100 != 0 {
		t.Fatalf("120/112 -> %v/%v want 8/0", ot50, ot100)
	}
}

func TestCalculatePeriodOT128(t *testing.T) {
	// One long day isn't enough — build 130h of work with short shifts.
	loc := time.UTC
	from := time.Date(2026, 7, 6, 0, 0, 0, 0, loc)
	to := from.AddDate(0, 0, 20) // 21 days

	var shifts []Shift
	// 13 days * 10h = 130h
	for i := 0; i < 13; i++ {
		day := from.AddDate(0, 0, i)
		shifts = append(shifts, Shift{Start: at(day, 6, 0), End: at(day, 16, 0)})
	}
	rules := DefaultRules()
	rules.PeriodOTEnabled = true
	rules.PeriodThresholdH = PeriodThreshold128
	rules.PeriodOT50AfterH = 18
	rules.ShiftOTAfterH = 0 // isolate period OT
	rules.Overtime50AfterH = 24
	rules.Overtime100AfterH = 24

	got := Calculate(PeriodInput{
		From: from, To: to,
		Shifts: shifts,
		Rates:  Rates{Hourly: 10},
		Rules:  rules,
	})
	if got.PeriodTotalHours != 130 {
		t.Fatalf("period total=%v", got.PeriodTotalHours)
	}
	if got.PeriodOT50Hours != 2 || got.PeriodOT100Hours != 0 {
		t.Fatalf("period ot=%v/%v want 2/0", got.PeriodOT50Hours, got.PeriodOT100Hours)
	}
	// base 1300 + period OT50 2*10*0.5 = 10 → 1310
	if math.Abs(got.TotalPay-1310) > 0.01 {
		t.Fatalf("pay=%v", got.TotalPay)
	}
}

func TestCalculatePeriodOT112WithAbsences(t *testing.T) {
	// TES liite 3 example 2 style: 96 work + 40 vacation = 136 → 16 OT over 120.
	// Here use 112 threshold: 96 + 40 = 136 → 24 OT → 18@50 + 6@100.
	loc := time.UTC
	from := time.Date(2026, 7, 6, 0, 0, 0, 0, loc)
	to := from.AddDate(0, 0, 20)

	var shifts []Shift
	for i := 0; i < 12; i++ { // 12*8 = 96h
		day := from.AddDate(0, 0, i)
		shifts = append(shifts, Shift{Start: at(day, 8, 0), End: at(day, 16, 0)})
	}
	rules := DefaultRules()
	rules.PeriodOTEnabled = true
	rules.PeriodThresholdH = PeriodThreshold112
	rules.PeriodOT50AfterH = 18
	rules.ShiftOTAfterH = 0
	rules.Overtime50AfterH = 24
	rules.Overtime100AfterH = 24

	got := Calculate(PeriodInput{
		From: from, To: to,
		Shifts:           shifts,
		Rates:            Rates{Hourly: 10},
		Rules:            rules,
		CreditedAbsenceH: 40, // 6 lomapaivaa * 6.7 ≈ 40
	})
	if got.PeriodWorkedHours != 96 || got.PeriodCreditedHours != 40 {
		t.Fatalf("worked=%v credited=%v", got.PeriodWorkedHours, got.PeriodCreditedHours)
	}
	if got.PeriodOT50Hours != 18 || got.PeriodOT100Hours != 6 {
		t.Fatalf("ot=%v/%v want 18/6", got.PeriodOT50Hours, got.PeriodOT100Hours)
	}
	// base 960 + ot50 18*5 + ot100 6*10 = 960+90+60 = 1110
	if math.Abs(got.TotalPay-1110) > 0.01 {
		t.Fatalf("pay=%v", got.TotalPay)
	}
}

func TestVacationDayHours(t *testing.T) {
	if math.Abs(6*VacationDayHours-40.2) > 0.001 && math.Abs(6*VacationDayHours-40) > 1 {
		// TES example uses 40 for 6 days (= 6.666… rounded as 6.7*6=40.2; example says 40)
		t.Fatalf("vacation day hours=%v", VacationDayHours)
	}
}
