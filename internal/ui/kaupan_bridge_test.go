package ui

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"payroll/internal/calc"
)

func TestVartiointiSaturdayWholeDayAfterApply(t *testing.T) {
	// Regression: time picker 00:00->24:00 must not zero out Saturday window.
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.applyDemoTES()
	rules := s.calcRules()
	if rules.SaturdayStartMin != 0 {
		t.Fatalf("sat start=%d want 0", rules.SaturdayStartMin)
	}
	day := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)
	got := calc.Calculate(calc.PeriodInput{
		From: day, To: day,
		Shifts: []calc.Shift{{
			Start: time.Date(2026, 7, 18, 8, 0, 0, 0, time.UTC),
			End: time.Date(2026, 7, 18, 16, 0, 0, 0, time.UTC),
		}},
		Rates: s.rates(),
		Rules: rules,
	})
	if got.SaturdayHours != 8 {
		t.Fatalf("saturday=%v want 8 (whole day)", got.SaturdayHours)
	}
}

func TestKaupanPayBPKS2y(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.applyKaupanPay("B", true, kaupanService2y)

	if s.tesFamily.Selected != tesFamilyKauppa {
		t.Fatalf("family=%q", s.tesFamily.Selected)
	}
	if s.tesLevel.Selected != "B" || s.tesRegion.Selected != tesRegionPKS {
		t.Fatalf("group/region=%q/%q", s.tesLevel.Selected, s.tesRegion.Selected)
	}
	if s.hourlyWage.Text != "13.33" || s.levelPay.Text != "2132.00" {
		t.Fatalf("pay=%q/%q", s.hourlyWage.Text, s.levelPay.Text)
	}
	eve, night, sat := calc.KaupanAllowances(true)
	if s.eveningAllowance.Text != formatEuro(eve) ||
		s.nightAllowance.Text != formatEuro(night) ||
		s.saturdayAllowance.Text != formatEuro(sat) {
		t.Fatalf("allowances=%q/%q/%q", s.eveningAllowance.Text, s.nightAllowance.Text, s.saturdayAllowance.Text)
	}
	if s.eveningDoubleAllowance.Text != formatEuro(calc.KaupanEveningDouble(eve)) {
		t.Fatalf("2x=%q", s.eveningDoubleAllowance.Text)
	}
	if s.sundayAllowance.Text != "13.33" {
		t.Fatalf("su=%q", s.sundayAllowance.Text)
	}
	if s.dailyRestHours.Text != "11" {
		t.Fatalf("lepo=%q", s.dailyRestHours.Text)
	}
}

func TestKaupanPayDMuu9y(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.applyKaupanPay("D", false, kaupanService9y)

	// 2 v -perus 14.36; 9 v taulukko 16.92 -> kokemuslisä 2.56 e/h
	if s.hourlyWage.Text != "14.36" || s.levelPay.Text != "2707.00" {
		t.Fatalf("pay=%q/%q", s.hourlyWage.Text, s.levelPay.Text)
	}
	if s.experienceAllowance.Text != "2.56" {
		t.Fatalf("exp=%q want 2.56", s.experienceAllowance.Text)
	}
	if s.sundayAllowance.Text != "16.92" || s.holidayAllowance.Text != "16.92" {
		t.Fatalf("su/pyhä=%q/%q want full table hourly", s.sundayAllowance.Text, s.holidayAllowance.Text)
	}
	r := s.rates()
	if math.Abs(r.EffectiveHourly()-16.92) > 0.001 {
		t.Fatalf("effective=%v want 16.92", r.EffectiveHourly())
	}
	eve, night, sat := calc.KaupanAllowances(false)
	if s.eveningAllowance.Text != formatEuro(eve) ||
		s.nightAllowance.Text != formatEuro(night) ||
		s.saturdayAllowance.Text != formatEuro(sat) {
		t.Fatalf("ilta/yö/la Muu=%q/%q/%q", s.eveningAllowance.Text, s.nightAllowance.Text, s.saturdayAllowance.Text)
	}
}

func TestKaupanExperienceSeparatedFromHourly(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()

	s.applyKaupanPay("B", true, kaupanService2y)
	if s.hourlyWage.Text != "13.33" || s.experienceAllowance.Text != "0.00" {
		t.Fatalf("2v: hourly=%q exp=%q", s.hourlyWage.Text, s.experienceAllowance.Text)
	}

	s.applyKaupanPay("B", true, kaupanService9y)
	// PKS B: 2v 13.33, 9v 15.26 -> exp 1.93
	if s.hourlyWage.Text != "13.33" {
		t.Fatalf("9v hourly must stay base, got %q", s.hourlyWage.Text)
	}
	if s.experienceAllowance.Text != "1.93" {
		t.Fatalf("9v exp=%q want 1.93", s.experienceAllowance.Text)
	}
	if math.Abs(s.rates().EffectiveHourly()-15.26) > 0.001 {
		t.Fatalf("effective=%v", s.rates().EffectiveHourly())
	}
}

func TestKaupanCalcRulesBridge(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.applyKaupanPay("B", true, kaupanService2y)

	got := s.calcRules()
	want := calc.KaupanMyyjaRules()
	if got.EveningStartMin != want.EveningStartMin || got.EveningEndMin != want.EveningEndMin {
		t.Fatalf("evening=%d-%d want %d-%d", got.EveningStartMin, got.EveningEndMin, want.EveningStartMin, want.EveningEndMin)
	}
	if got.NightStartMin != want.NightStartMin || got.NightEndMin != want.NightEndMin {
		t.Fatalf("night=%d-%d want %d-%d (00:00 must bridge to 0)", got.NightStartMin, got.NightEndMin, want.NightStartMin, want.NightEndMin)
	}
	if got.SaturdayStartMin != want.SaturdayStartMin || got.SaturdayEndMin != want.SaturdayEndMin {
		t.Fatalf("sat=%d-%d want %d-%d", got.SaturdayStartMin, got.SaturdayEndMin, want.SaturdayStartMin, want.SaturdayEndMin)
	}
	if got.EveningExcludeSaturday != want.EveningExcludeSaturday ||
		got.NightExcludeSunday != want.NightExcludeSunday ||
		got.NightExcludeHoliday != want.NightExcludeHoliday {
		t.Fatalf("exclusions got sat=%v sun=%v hol=%v", got.EveningExcludeSaturday, got.NightExcludeSunday, got.NightExcludeHoliday)
	}
	if got.EveningDoubleMonthFrom != want.EveningDoubleMonthFrom ||
		got.EveningDoubleMonthTo != want.EveningDoubleMonthTo ||
		got.EveningDoubleSundayOnly != want.EveningDoubleSundayOnly {
		t.Fatalf("double window %+v", got)
	}
	if got.Overtime50AfterH != want.Overtime50AfterH || got.Overtime100AfterH != want.Overtime100AfterH {
		t.Fatalf("daily OT=%v/%v", got.Overtime50AfterH, got.Overtime100AfterH)
	}
	if got.ShiftOTAfterH != want.ShiftOTAfterH {
		t.Fatalf("ShiftOT=%v want %v", got.ShiftOTAfterH, want.ShiftOTAfterH)
	}
	if got.WeeklyOTEnabled != want.WeeklyOTEnabled || math.Abs(got.WeeklyOTThresholdH-want.WeeklyOTThresholdH) > 0.001 {
		t.Fatalf("weekly=%v/%v", got.WeeklyOTEnabled, got.WeeklyOTThresholdH)
	}
	if got.PeriodOTEnabled != want.PeriodOTEnabled {
		t.Fatalf("period OT=%v", got.PeriodOTEnabled)
	}

	rates := s.rates()
	eve, night, sat := calc.KaupanAllowances(true)
	if math.Abs(rates.Evening-eve) > 0.001 || math.Abs(rates.Night-night) > 0.001 || math.Abs(rates.Saturday-sat) > 0.001 {
		t.Fatalf("rates ilta/yö/la=%v/%v/%v", rates.Evening, rates.Night, rates.Saturday)
	}
	if math.Abs(rates.EveningDouble-calc.KaupanEveningDouble(eve)) > 0.001 {
		t.Fatalf("evening double=%v", rates.EveningDouble)
	}
}

func TestKaupanRatesBridgeToCalculateSaturday(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.applyKaupanPay("B", true, kaupanService2y)

	day := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC) // Sat
	got := calc.Calculate(calc.PeriodInput{
		From: day, To: day,
		Shifts: []calc.Shift{{
			Start: time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC),
			End: time.Date(2026, 7, 18, 20, 0, 0, 0, time.UTC),
		}},
		Rates: s.rates(),
		Rules: s.calcRules(),
	})
	if math.Abs(got.SaturdayHours-7) > 0.001 || got.EveningHours != 0 {
		t.Fatalf("sat=%v evening=%v", got.SaturdayHours, got.EveningHours)
	}
	want := 8*13.33 + 7*calc.KaupanSaturdayPKS
	if math.Abs(got.TotalPay-want) > 0.05 {
		t.Fatalf("pay=%v want %v", got.TotalPay, want)
	}
}

func formatEuro(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func TestTESFamilyDropdownIncludesKauppa(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()

	opts := s.tesFamily.Options
	if len(opts) != 3 {
		t.Fatalf("TES options len=%d want 3: %v", len(opts), opts)
	}
	for _, want := range []string{tesFamilyCustom, tesFamilyVartio, tesFamilyKauppa} {
		if !containsString(opts, want) {
			t.Fatalf("missing %q in %v", want, opts)
		}
	}

	s.tesFamily.SetSelected(tesFamilyKauppa)
	if s.tesFamily.Selected != tesFamilyKauppa {
		t.Fatalf("selected=%q", s.tesFamily.Selected)
	}
	if s.hourlyWage.Text != "13.33" { // B / PKS / 2v default from apply
		// apply may use previous level B or default B
		if s.tesLevel.Selected != "B" && s.tesLevel.Selected != "A" && s.tesLevel.Selected != "C" && s.tesLevel.Selected != "D" {
			t.Fatalf("kaupan level not applied: level=%q hourly=%q", s.tesLevel.Selected, s.hourlyWage.Text)
		}
	}
	rules := s.calcRules()
	if rules.WeeklyOTEnabled != true || rules.ShiftOTAfterH != 0 {
		t.Fatalf("kaupan rules weekly=%v shiftOT=%v", rules.WeeklyOTEnabled, rules.ShiftOTAfterH)
	}
	if s.trainingSection != nil && s.trainingSection.Visible() {
		t.Fatal("koulutuslisä should be hidden for Kauppa")
	}

	// Switching families must not drop Kauppa from the list.
	s.tesFamily.SetSelected(tesFamilyVartio)
	s.ensureTESFamilyOptions()
	if !containsString(s.tesFamily.Options, tesFamilyKauppa) {
		t.Fatalf("Kauppa dropped after switch: %v", s.tesFamily.Options)
	}
}

func TestKaupanSelectorOptionsAndSwitchBackToVartio(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.applyKaupanPay("C", true, kaupanService6y)
	if len(s.tesLevel.Options) != 4 || s.tesLevel.Options[0] != "A" {
		t.Fatalf("kaupan levels=%v", s.tesLevel.Options)
	}
	if len(s.experienceTier.Options) != 4 {
		t.Fatalf("kaupan service=%v", s.experienceTier.Options)
	}

	s.applyVartiointiPay("Taso IV", true, tesServicePerus)
	if s.tesFamily.Selected != tesFamilyVartio {
		t.Fatalf("family=%q", s.tesFamily.Selected)
	}
	if s.hourlyWage.Text != "13.97" {
		t.Fatalf("vartio hourly=%q", s.hourlyWage.Text)
	}
	rules := s.calcRules()
	if rules.ShiftOTAfterH != 12 || !rules.PeriodOTEnabled || rules.WeeklyOTEnabled {
		t.Fatalf("vartio rules shift=%v period=%v weekly=%v",
			rules.ShiftOTAfterH, rules.PeriodOTEnabled, rules.WeeklyOTEnabled)
	}
	if rules.EveningExcludeSaturday {
		t.Fatal("vartio should allow evening on Saturday")
	}
}

func TestCalcTabKaupanWeeklyOTVisible(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	settings := newSettingsTab()
	_ = settings.canvas()
	settings.applyKaupanPay("B", true, kaupanService2y)

	shifts := newShiftsTab(w)
	_ = shifts.canvas()
	// Mon-Fri 8h in July 2026 (week of 13th) -> 40h -> weekly OT 2.5
	loc := time.Local
	for i := 0; i < 5; i++ {
		day := time.Date(2026, 7, 13+i, 0, 0, 0, 0, loc)
		shifts.shifts = append(shifts.shifts, calendarShift{
			Date: day, Start: "08:00", End: "16:00",
		})
	}

	c := newCalcTab(settings, shifts)
	_ = c.canvas()
	c.from.SetText("13.07.2026")
	c.to.SetText("17.07.2026")
	c.run()

	if c.err.Visible() {
		t.Fatalf("err=%q", c.err.Text)
	}
	if !strings.Contains(c.details.Text, "Viikkoylitys") {
		t.Fatalf("details missing weekly OT: %q", c.details.Text)
	}
	if strings.Contains(c.summary.Text, "Jakso:") {
		t.Fatalf("Kaupan should not show period OT line: %q", c.summary.Text)
	}
}
