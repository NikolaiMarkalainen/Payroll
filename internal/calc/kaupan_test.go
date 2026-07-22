package calc

import (
	"math"
	"testing"
	"time"
)

func TestKaupanSaturdayWindowNoEvening(t *testing.T) {
	// Arkilauantai 12–20: la-lisä vain 13–20 (7h), ei iltalisää.
	day := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC) // Saturday
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 12, 0), End: at(day, 20, 0)}},
		Rates:  Rates{Hourly: 13.33, Evening: KaupanEveningPKS, Saturday: KaupanSaturdayPKS},
		Rules:  KaupanMyyjaRules(),
	})
	if got.BaseHours != 8 {
		t.Fatalf("base=%v", got.BaseHours)
	}
	if math.Abs(got.SaturdayHours-7) > 0.001 {
		t.Fatalf("sat=%v want 7", got.SaturdayHours)
	}
	if got.EveningHours != 0 {
		t.Fatalf("evening=%v want 0", got.EveningHours)
	}
	want := 8*13.33 + 7*KaupanSaturdayPKS
	if math.Abs(got.TotalPay-want) > 0.02 {
		t.Fatalf("pay=%v want %v", got.TotalPay, want)
	}
}

func TestKaupanSundayNightExcludedEveningDouble(t *testing.T) {
	// Sun 2026-12-06 (also holiday): 22–06 overnight → no night on Sunday;
	// evening 22–24 on Sunday in Dec → double evening.
	sun := time.Date(2026, 12, 6, 0, 0, 0, 0, time.UTC)
	mon := sun.AddDate(0, 0, 1)
	got := Calculate(PeriodInput{
		From: sun, To: mon,
		Shifts: []Shift{{Start: at(sun, 22, 0), End: at(mon, 6, 0)}},
		Rates: Rates{
			Hourly: 13.33, Evening: KaupanEveningPKS, EveningDouble: KaupanEveningDouble(KaupanEveningPKS),
			Night: KaupanNightPKS, Sunday: 13.33, Holiday: 13.33,
		},
		Rules: KaupanMyyjaRules(),
	})
	if got.BaseHours != 8 {
		t.Fatalf("base=%v", got.BaseHours)
	}
	// Sunday slice 22–24: 2h evening double; Monday 00–06: 6h night (Mon ok).
	if math.Abs(got.EveningDoubleHours-2) > 0.001 {
		t.Fatalf("eveDouble=%v want 2", got.EveningDoubleHours)
	}
	if math.Abs(got.NightHours-6) > 0.001 {
		t.Fatalf("night=%v want 6 (Mon only)", got.NightHours)
	}
	if got.Days[0].Night != 0 {
		t.Fatalf("Sunday night=%v want 0", got.Days[0].Night)
	}
}

func TestKaupanDailyOT50After10(t *testing.T) {
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC) // Wed
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 8, 0), End: at(day, 20, 0)}}, // 12h → 2h @50%
		Rates:  Rates{Hourly: 13.33},
		Rules:  KaupanMyyjaRules(),
	})
	if math.Abs(got.Overtime50Hours-2) > 0.001 || got.Overtime100Hours != 0 {
		t.Fatalf("ot=%v/%v", got.Overtime50Hours, got.Overtime100Hours)
	}
	want := 12*13.33 + 2*13.33*0.5
	if math.Abs(got.TotalPay-want) > 0.02 {
		t.Fatalf("pay=%v want %v", got.TotalPay, want)
	}
}

func TestKaupanWeeklyOT50After375(t *testing.T) {
	// Mon–Fri 8h = 40h → weekly OT 2.5h @50% (no daily OT).
	loc := time.UTC
	from := time.Date(2026, 7, 13, 0, 0, 0, 0, loc) // Mon
	to := time.Date(2026, 7, 17, 0, 0, 0, 0, loc)   // Fri
	var shifts []Shift
	for i := 0; i < 5; i++ {
		day := from.AddDate(0, 0, i)
		shifts = append(shifts, Shift{Start: at(day, 8, 0), End: at(day, 16, 0)})
	}
	got := Calculate(PeriodInput{
		From: from, To: to,
		Shifts: shifts,
		Rates:  Rates{Hourly: 10},
		Rules:  KaupanMyyjaRules(),
	})
	if math.Abs(got.BaseHours-40) > 0.001 {
		t.Fatalf("base=%v", got.BaseHours)
	}
	if math.Abs(got.WeeklyOT50Hours-2.5) > 0.001 {
		t.Fatalf("weeklyOT=%v want 2.5", got.WeeklyOT50Hours)
	}
	if math.Abs(got.Overtime50Hours-2.5) > 0.001 {
		t.Fatalf("ot50=%v", got.Overtime50Hours)
	}
	want := 40*10 + 2.5*10*0.5
	if math.Abs(got.TotalPay-want) > 0.02 {
		t.Fatalf("pay=%v want %v", got.TotalPay, want)
	}
}

func TestKaupanVartioDefaultUnaffected(t *testing.T) {
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 6, 0), End: at(day, 20, 0)}}, // 14h
		Rates:  Rates{Hourly: 10},
		Rules:  VartiointiRules(),
	})
	if got.Overtime50Hours != 2 || got.ShiftOTHours != 2 {
		t.Fatalf("vartio ot=%v shift=%v", got.Overtime50Hours, got.ShiftOTHours)
	}
}

func TestKaupanWeekdayEvening18to24(t *testing.T) {
	// Wed 17–24: 1h plain + 6h evening (18–24); yö alkaa vasta 00.
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 17, 0), End: at(day, 24, 0)}},
		Rates:  Rates{Hourly: 13.33, Evening: KaupanEveningPKS, Night: KaupanNightPKS},
		Rules:  KaupanMyyjaRules(),
	})
	if got.BaseHours != 7 {
		t.Fatalf("base=%v", got.BaseHours)
	}
	if math.Abs(got.EveningHours-6) > 0.001 {
		t.Fatalf("evening=%v want 6", got.EveningHours)
	}
	if got.NightHours != 0 || got.EveningDoubleHours != 0 {
		t.Fatalf("night=%v eveDouble=%v", got.NightHours, got.EveningDoubleHours)
	}
}

func TestKaupanSaturdayExactly13Boundary(t *testing.T) {
	day := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 13, 0), End: at(day, 14, 0)}},
		Rates:  Rates{Hourly: 10, Saturday: KaupanSaturdayPKS, Evening: KaupanEveningPKS},
		Rules:  KaupanMyyjaRules(),
	})
	if math.Abs(got.SaturdayHours-1) > 0.001 {
		t.Fatalf("sat=%v want 1", got.SaturdayHours)
	}
	if got.EveningHours != 0 {
		t.Fatalf("evening=%v", got.EveningHours)
	}
}

func TestKaupanHolidayNightExcluded(t *testing.T) {
	// Midsummer Eve 2026-06-19 is Friday holiday — night 00–06 excluded.
	day := time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC)
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 0, 0), End: at(day, 6, 0)}},
		Rates:  Rates{Hourly: 13.33, Night: KaupanNightPKS, Holiday: 13.33},
		Rules:  KaupanMyyjaRules(),
	})
	if got.NightHours != 0 {
		t.Fatalf("night=%v want 0 on holiday", got.NightHours)
	}
	if got.HolidayHours != 6 {
		t.Fatalf("holiday=%v", got.HolidayHours)
	}
}

func TestKaupanJulySundayNoEveningDouble(t *testing.T) {
	// Double evening only Nov–Dec Sundays.
	day := time.Date(2026, 7, 19, 0, 0, 0, 0, time.UTC) // Sunday
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{{Start: at(day, 18, 0), End: at(day, 22, 0)}},
		Rates: Rates{
			Hourly: 13.33, Evening: KaupanEveningPKS,
			EveningDouble: KaupanEveningDouble(KaupanEveningPKS), Sunday: 13.33,
		},
		Rules: KaupanMyyjaRules(),
	})
	if math.Abs(got.EveningHours-4) > 0.001 {
		t.Fatalf("evening=%v", got.EveningHours)
	}
	if got.EveningDoubleHours != 0 {
		t.Fatalf("eveDouble=%v want 0 in July", got.EveningDoubleHours)
	}
}

func TestKaupanDailyAndWeeklyOTNoDoubleCount(t *testing.T) {
	// One 12h day (2h daily OT) + four 8h = 44h; weekly excess over 37.5
	// must exclude the 2h already paid as daily OT.
	loc := time.UTC
	from := time.Date(2026, 7, 13, 0, 0, 0, 0, loc) // Mon
	to := time.Date(2026, 7, 17, 0, 0, 0, 0, loc)   // Fri
	var shifts []Shift
	shifts = append(shifts, Shift{Start: at(from, 8, 0), End: at(from, 20, 0)}) // 12h
	for i := 1; i < 5; i++ {
		day := from.AddDate(0, 0, i)
		shifts = append(shifts, Shift{Start: at(day, 8, 0), End: at(day, 16, 0)})
	}
	got := Calculate(PeriodInput{
		From: from, To: to,
		Shifts: shifts,
		Rates:  Rates{Hourly: 10},
		Rules:  KaupanMyyjaRules(),
	})
	if math.Abs(got.BaseHours-44) > 0.001 {
		t.Fatalf("base=%v", got.BaseHours)
	}
	// daily OT 2 + weekly (44-2-37.5)=4.5 → ot50 total 6.5
	if math.Abs(got.WeeklyOT50Hours-4.5) > 0.001 {
		t.Fatalf("weekly=%v want 4.5", got.WeeklyOT50Hours)
	}
	if math.Abs(got.Overtime50Hours-6.5) > 0.001 {
		t.Fatalf("ot50=%v want 6.5", got.Overtime50Hours)
	}
}
