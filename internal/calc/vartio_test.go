package calc

import (
	"math"
	"testing"
	"time"
)

// Vartiointiala TES 1.8.2025 fixed allowances + Taso IV PK-seutu perus hourly.
const (
	vartioHourlyIV = 13.97
	vartioEvening  = 1.11
	vartioNight    = 2.45
	vartioSaturday = 2.18
)

func TestCalculateVartioEveningNightSaturday(t *testing.T) {
	// Wed 15:00–23:00 → 8h base, 4h evening (18–22), 1h night (22–23)
	// Sat 08:00–16:00 → 8h saturday
	loc := time.UTC
	wed := time.Date(2026, 7, 15, 0, 0, 0, 0, loc)
	sat := time.Date(2026, 7, 18, 0, 0, 0, 0, loc)

	got := Calculate(PeriodInput{
		From: wed, To: sat,
		Shifts: []Shift{
			{Start: at(wed, 15, 0), End: at(wed, 23, 0)},
			{Start: at(sat, 8, 0), End: at(sat, 16, 0)},
		},
		Rates: Rates{
			Hourly:   vartioHourlyIV,
			Evening:  vartioEvening,
			Night:    vartioNight,
			Saturday: vartioSaturday,
		},
		Rules: DefaultRules(),
	})

	if got.BaseHours != 16 {
		t.Fatalf("base=%v", got.BaseHours)
	}
	if math.Abs(got.EveningHours-4) > 0.001 || math.Abs(got.NightHours-1) > 0.001 {
		t.Fatalf("evening=%v night=%v", got.EveningHours, got.NightHours)
	}
	if math.Abs(got.SaturdayHours-8) > 0.001 {
		t.Fatalf("saturday=%v", got.SaturdayHours)
	}

	want := 16*vartioHourlyIV + 4*vartioEvening + 1*vartioNight + 8*vartioSaturday
	if math.Abs(got.TotalPay-want) > 0.02 {
		t.Fatalf("pay=%v want %v", got.TotalPay, want)
	}
}

func TestCalculateEveningNightWindowBoundaries(t *testing.T) {
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	rules := DefaultRules()

	// Exactly on evening start → evening counts; ends exactly at night start → no night.
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 18, 0), End: at(day, 22, 0)}},
		Rates:  Rates{Hourly: 10, Evening: 1, Night: 2},
		Rules:  rules,
	})
	if got.EveningHours != 4 || got.NightHours != 0 {
		t.Fatalf("18–22: evening=%v night=%v", got.EveningHours, got.NightHours)
	}

	// Starts exactly at night start.
	got = Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 22, 0), End: at(day, 23, 0)}},
		Rates:  Rates{Hourly: 10, Evening: 1, Night: 2},
		Rules:  rules,
	})
	if got.EveningHours != 0 || got.NightHours != 1 {
		t.Fatalf("22–23: evening=%v night=%v", got.EveningHours, got.NightHours)
	}

	// Before evening window → no evening.
	got = Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 16, 0), End: at(day, 18, 0)}},
		Rates:  Rates{Hourly: 10, Evening: 1},
		Rules:  rules,
	})
	if got.EveningHours != 0 {
		t.Fatalf("16–18 evening=%v want 0", got.EveningHours)
	}
}

func TestCalculateZeroRatesStillCountsHours(t *testing.T) {
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 14, 0), End: at(day, 22, 0)}},
		Rates:  Rates{}, // all zero
		Rules:  DefaultRules(),
	})
	if got.BaseHours != 8 || got.EveningHours != 4 {
		t.Fatalf("hours base=%v evening=%v", got.BaseHours, got.EveningHours)
	}
	if got.TotalPay != 0 {
		t.Fatalf("pay=%v want 0", got.TotalPay)
	}
}

func TestCalculateVartioSundayEqualsHourly(t *testing.T) {
	// TES: sunnuntai-/pyhälisä = tehtäväkohtainen tuntipalkka (taso+palvelusaika).
	day := time.Date(2026, 7, 19, 0, 0, 0, 0, time.UTC) // Sunday
	hourly := 14.48                                     // Taso IV PK 2v
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 8, 0), End: at(day, 16, 0)}},
		Rates:  Rates{Hourly: 13.97, Experience: 0.51, Sunday: hourly},
		Rules:  DefaultRules(),
	})
	if got.SundayHours != 8 {
		t.Fatalf("sunday hours=%v", got.SundayHours)
	}
	wantSun := 8 * hourly
	if math.Abs(got.SundayPay-wantSun) > 0.01 {
		t.Fatalf("sunday pay=%v want %v", got.SundayPay, wantSun)
	}
}

func TestCalculateNoShiftsEmpty(t *testing.T) {
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: nil,
		Rates:  Rates{Hourly: vartioHourlyIV, Evening: vartioEvening},
		Rules:  DefaultRules(),
	})
	if got.BaseHours != 0 || got.TotalPay != 0 {
		t.Fatalf("empty: base=%v pay=%v", got.BaseHours, got.TotalPay)
	}
}
