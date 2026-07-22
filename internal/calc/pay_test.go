package calc

import (
	"math"
	"testing"
	"time"
)

func at(day time.Time, hh, mm int) time.Time {
	y, m, d := day.Date()
	return time.Date(y, m, d, hh, mm, 0, 0, day.Location())
}

func TestCalculateWeekdayShiftBaseAndEvening(t *testing.T) {
	// Wednesday 2026-07-15 14:00–22:00 → 8h base, 4h evening (18–22)
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	rates := Rates{Hourly: 10, Evening: 1.5}
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 14, 0), End: at(day, 22, 0)}},
		Rates:  rates,
		Rules:  DefaultRules(),
	})
	if got.BaseHours != 8 || got.EveningHours != 4 {
		t.Fatalf("hours base=%v evening=%v", got.BaseHours, got.EveningHours)
	}
	wantTotal := 8*10 + 4*1.5 // 86
	if math.Abs(got.TotalPay-wantTotal) > 0.001 {
		t.Fatalf("pay=%v want %v", got.TotalPay, wantTotal)
	}
}

func TestCalculateOvernightNightSplit(t *testing.T) {
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	next := day.AddDate(0, 0, 1)
	got := Calculate(PeriodInput{
		From: day, To: next,
		Shifts: []Shift{{Start: at(day, 22, 0), End: at(next, 6, 0)}},
		Rates:  Rates{Hourly: 10, Night: 2},
		Rules:  DefaultRules(),
	})
	if got.BaseHours != 8 || got.NightHours != 8 {
		t.Fatalf("base=%v night=%v", got.BaseHours, got.NightHours)
	}
	if len(got.Days) != 2 {
		t.Fatalf("days=%d", len(got.Days))
	}
	if got.Days[0].Night != 2 || got.Days[1].Night != 6 {
		t.Fatalf("day nights=%v/%v", got.Days[0].Night, got.Days[1].Night)
	}
}

func TestCalculateHolidayAndSunday(t *testing.T) {
	// 2026-12-06 is Sunday + Itsenäisyyspäivä
	day := time.Date(2026, 12, 6, 0, 0, 0, 0, time.UTC)
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 8, 0), End: at(day, 16, 0)}},
		Rates:  Rates{Hourly: 10, Sunday: 3, Holiday: 5},
		Rules:  DefaultRules(),
	})
	if got.SundayHours != 8 || got.HolidayHours != 8 {
		t.Fatalf("sun=%v hol=%v", got.SundayHours, got.HolidayHours)
	}
	if got.Days[0].HolidayName != "Itsenäisyyspäivä" {
		t.Fatalf("name=%q", got.Days[0].HolidayName)
	}
	want := 8*10 + 8*3 + 8*5 // 144
	if math.Abs(got.TotalPay-float64(want)) > 0.001 {
		t.Fatalf("pay=%v want %v", got.TotalPay, want)
	}
}

func TestCalculateOvertime50And100(t *testing.T) {
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	// TES 31 §: 14h shift → 2h over 12 at 50% (under 18h cap).
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 6, 0), End: at(day, 20, 0)}}, // 14h
		Rates:  Rates{Hourly: 10},
		Rules:  DefaultRules(),
	})
	if got.Overtime50Hours != 2 || got.Overtime100Hours != 0 {
		t.Fatalf("ot=%v/%v want 2/0", got.Overtime50Hours, got.Overtime100Hours)
	}
	// base 14*10 + ot50 2*10*0.5 = 140+10 = 150
	if math.Abs(got.TotalPay-150) > 0.001 {
		t.Fatalf("pay=%v", got.TotalPay)
	}
}

func TestCalculateShiftOT100After18(t *testing.T) {
	// Two 22h shifts → 10+10=20h extension → 18@50 + 2@100.
	loc := time.UTC
	d1 := time.Date(2026, 7, 15, 0, 0, 0, 0, loc)
	d2 := time.Date(2026, 7, 16, 0, 0, 0, 0, loc)
	got := Calculate(PeriodInput{
		From: d1, To: d2,
		Shifts: []Shift{
			{Start: at(d1, 0, 0), End: at(d1, 22, 0)},
			{Start: at(d2, 0, 0), End: at(d2, 22, 0)},
		},
		Rates: Rates{Hourly: 10},
		Rules: DefaultRules(),
	})
	if got.ShiftOTHours != 20 {
		t.Fatalf("shift ot hours=%v", got.ShiftOTHours)
	}
	if got.Overtime50Hours != 18 || got.Overtime100Hours != 2 {
		t.Fatalf("ot=%v/%v want 18/2", got.Overtime50Hours, got.Overtime100Hours)
	}
}

func TestCalculateHypotheticalDateRange(t *testing.T) {
	// Hypothetical week 2026-12-01 (Tue) … 2026-12-07 (Mon):
	//  - Tue–Fri: 06–14 (8h) regular
	//  - Sat 06–14 (8h) saturday
	//  - Sun 06–14 (8h) sunday + independence day (Dec 6)
	//  - Mon ignored (outside shifts)
	loc := time.UTC
	from := time.Date(2026, 12, 1, 0, 0, 0, 0, loc)
	to := time.Date(2026, 12, 7, 0, 0, 0, 0, loc)

	var shifts []Shift
	for _, d := range []int{1, 2, 3, 4} { // Tue–Fri
		day := time.Date(2026, 12, d, 0, 0, 0, 0, loc)
		shifts = append(shifts, Shift{Start: at(day, 6, 0), End: at(day, 14, 0)})
	}
	sat := time.Date(2026, 12, 5, 0, 0, 0, 0, loc)
	sun := time.Date(2026, 12, 6, 0, 0, 0, 0, loc)
	shifts = append(shifts,
		Shift{Start: at(sat, 6, 0), End: at(sat, 14, 0)},
		Shift{Start: at(sun, 6, 0), End: at(sun, 14, 0)},
	)

	rates := Rates{
		Hourly:   12.00,
		Evening:  1.20,
		Night:    2.40,
		Saturday: 1.50,
		Sunday:   6.00,
		Holiday:  6.00,
	}

	got := Calculate(PeriodInput{
		From: from, To: to,
		Shifts: shifts,
		Rates:  rates,
		Rules:  DefaultRules(),
	})

	// 6 days * 8h = 48h base
	if got.BaseHours != 48 {
		t.Fatalf("base hours=%v want 48", got.BaseHours)
	}
	if got.SaturdayHours != 8 {
		t.Fatalf("saturday=%v", got.SaturdayHours)
	}
	if got.SundayHours != 8 || got.HolidayHours != 8 {
		t.Fatalf("sunday=%v holiday=%v", got.SundayHours, got.HolidayHours)
	}
	if got.EveningHours != 0 || got.NightHours != 0 {
		t.Fatalf("unexpected evening/night %v/%v", got.EveningHours, got.NightHours)
	}
	if got.Overtime50Hours != 0 || got.Overtime100Hours != 0 {
		t.Fatalf("unexpected OT %v/%v", got.Overtime50Hours, got.Overtime100Hours)
	}

	want := 48*12 + 8*1.5 + 8*6 + 8*6 // 576 + 12 + 48 + 48 = 684
	if math.Abs(got.TotalPay-want) > 0.001 {
		t.Fatalf("total=%v want %v; breakdown=%+v", got.TotalPay, want, got)
	}

	// Ensure Dec 6 day carries holiday name in daily breakdown.
	found := false
	for _, d := range got.Days {
		if Key(d.Date) == "2026-12-06" {
			found = true
			if d.HolidayName != "Itsenäisyyspäivä" {
				t.Fatalf("dec6 name=%q", d.HolidayName)
			}
			if d.Holiday != 8 || d.Sunday != 8 {
				t.Fatalf("dec6 hours hol=%v sun=%v", d.Holiday, d.Sunday)
			}
		}
	}
	if !found {
		t.Fatal("missing 2026-12-06 in days")
	}
}

func TestCalculateRangeClipsShiftsOutsidePeriod(t *testing.T) {
	from := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	before := time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC)
	got := Calculate(PeriodInput{
		From: from, To: to,
		Shifts: []Shift{
			{Start: at(before, 6, 0), End: at(before, 14, 0)},
			{Start: at(from, 6, 0), End: at(from, 10, 0)},
		},
		Rates: Rates{Hourly: 10},
		Rules: DefaultRules(),
	})
	if got.BaseHours != 4 {
		t.Fatalf("base=%v want 4 (outside clipped)", got.BaseHours)
	}
}

func TestCalculatePersonalAndExperience(t *testing.T) {
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	rates := Rates{
		Hourly:     10,
		Experience: 0.50,
		Personal:   1.00,
		Training:   0.25,
	}
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 8, 0), End: at(day, 16, 0)}}, // 8h
		Rates:  rates,
		Rules:  DefaultRules(),
	})
	if math.Abs(got.ExperiencePay-4) > 0.001 || math.Abs(got.PersonalPay-8) > 0.001 {
		t.Fatalf("exp=%v pers=%v", got.ExperiencePay, got.PersonalPay)
	}
	if math.Abs(got.TrainingPay-2) > 0.001 {
		t.Fatalf("training=%v", got.TrainingPay)
	}
	if math.Abs(rates.EffectiveHourly()-11.5) > 0.001 {
		t.Fatalf("OT base must exclude training: %v", rates.EffectiveHourly())
	}
	// base 80 + exp 4 + pers 8 + train 2 = 94
	if math.Abs(got.TotalPay-94) > 0.001 {
		t.Fatalf("total=%v", got.TotalPay)
	}
}

func TestOtherAllowanceHourlyAndFixed(t *testing.T) {
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	hourly := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 8, 0), End: at(day, 16, 0)}}, // 8h
		Rates:  Rates{Hourly: 10, OtherHourly: 0.50},
		Rules:  DefaultRules(),
	})
	if math.Abs(hourly.OtherPay-4) > 0.001 {
		t.Fatalf("hourly other=%v", hourly.OtherPay)
	}
	fixed := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 8, 0), End: at(day, 16, 0)}},
		Rates:  Rates{Hourly: 10, OtherFixed: 25},
		Rules:  DefaultRules(),
	})
	if math.Abs(fixed.OtherPay-25) > 0.001 {
		t.Fatalf("fixed other=%v", fixed.OtherPay)
	}
	// base 80 + other 25 = 105
	if math.Abs(fixed.TotalPay-105) > 0.001 {
		t.Fatalf("total=%v", fixed.TotalPay)
	}
}

func TestTrainingNotInOvertimeBase(t *testing.T) {
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 6, 0), End: at(day, 20, 0)}}, // 14h → 2h shift OT @50%
		Rates:  Rates{Hourly: 10, Training: 0.25},
		Rules:  DefaultRules(),
	})
	// OT premium uses 10, not 10.25: 2 * 10 * 0.5 = 10
	if math.Abs(got.Overtime50Pay-10) > 0.001 {
		t.Fatalf("ot50 pay=%v want 10 (training excluded)", got.Overtime50Pay)
	}
	// base 140 + training 14*0.25=3.5 + ot 10 = 153.5
	if math.Abs(got.TotalPay-153.5) > 0.001 {
		t.Fatalf("total=%v", got.TotalPay)
	}
}
