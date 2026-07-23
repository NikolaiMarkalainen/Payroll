package ui

import (
	"math"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"payroll/internal/calc"
)

func TestVartiointiRatesBridgeToCalculate(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.applyVartiointiPay("Taso IV", true, tesServicePerus)
	s.trainingEnabled.SetChecked(true)

	r := s.rates()
	if math.Abs(r.Hourly-13.97) > 0.001 {
		t.Fatalf("hourly=%v", r.Hourly)
	}
	if math.Abs(r.Evening-1.11) > 0.001 || math.Abs(r.Night-2.45) > 0.001 {
		t.Fatalf("ilta/yö=%v/%v", r.Evening, r.Night)
	}
	if math.Abs(r.Saturday-2.18) > 0.001 {
		t.Fatalf("lauantai=%v", r.Saturday)
	}
	if math.Abs(r.Sunday-13.97) > 0.001 || math.Abs(r.Holiday-13.97) > 0.001 {
		t.Fatalf("sun/pyhä=%v/%v", r.Sunday, r.Holiday)
	}
	if math.Abs(r.Training-0.25) > 0.001 {
		t.Fatalf("koulutus=%v", r.Training)
	}

	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	got := calc.Calculate(calc.PeriodInput{
		From: day, To: day,
		Shifts: []calc.Shift{{
			Start: time.Date(2026, 7, 15, 14, 0, 0, 0, time.UTC),
			End: time.Date(2026, 7, 15, 22, 0, 0, 0, time.UTC),
		}},
		Rates: r,
		Rules: s.calcRules(),
	})
	// 8h base + 4h evening; training on all hours
	want := 8*13.97 + 4*1.11 + 8*0.25
	if math.Abs(got.TotalPay-want) > 0.05 {
		t.Fatalf("pay=%v want %v", got.TotalPay, want)
	}
}

func TestVartiointiCalcRulesDefaults(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.applyDemoTES()

	rules := s.calcRules()
	if rules.EveningStartMin != 18*60 || rules.EveningEndMin != 22*60 {
		t.Fatalf("evening window=%d-%d", rules.EveningStartMin, rules.EveningEndMin)
	}
	if rules.NightStartMin != 22*60 || rules.NightEndMin != 6*60 {
		t.Fatalf("night window=%d-%d", rules.NightStartMin, rules.NightEndMin)
	}
	if rules.ShiftOTAfterH != 12 || rules.ShiftOT50CapH != 18 {
		t.Fatalf("shift OT=%v/%v", rules.ShiftOTAfterH, rules.ShiftOT50CapH)
	}
	if !rules.PeriodOTEnabled {
		t.Fatal("period OT should be enabled for Vartio demo")
	}
	if s.periodMode.Selected != periodMode128_112 {
		t.Fatalf("period mode=%q", s.periodMode.Selected)
	}
	if s.dailyRestHours.Text != "10" || s.restViolationPercent.Text != "50" {
		t.Fatalf("lepo=%q / %q", s.dailyRestHours.Text, s.restViolationPercent.Text)
	}
	// Lepokentät ovat asetuksissa; Calculate ei vielä käytä niitä.
}

func TestVartiointiService7yExperience(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.applyVartiointiPay("Taso IV", true, tesService7y)

	// perus 13.97, 7v 14.97 -> exp 1.00
	if s.hourlyWage.Text != "13.97" {
		t.Fatalf("base=%q", s.hourlyWage.Text)
	}
	if s.experienceAllowance.Text != "1.00" {
		t.Fatalf("exp=%q", s.experienceAllowance.Text)
	}
	if s.levelPay.Text != "2425.00" {
		t.Fatalf("kk=%q", s.levelPay.Text)
	}
	if s.sundayAllowance.Text != "14.97" {
		t.Fatalf("sunday=%q", s.sundayAllowance.Text)
	}
	r := s.rates()
	if math.Abs(r.EffectiveHourly()-14.97) > 0.001 {
		t.Fatalf("effective=%v", r.EffectiveHourly())
	}
}

func TestVartiointiPeriodThreshold128FromAnchor(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.applyDemoTES() // anchor 06.07.2026, mode 128/112

	from := time.Date(2026, 7, 20, 0, 0, 0, 0, time.Local)
	th := s.periodThresholdForRange(from)
	if th != calc.PeriodThreshold128 {
		t.Fatalf("threshold=%v want 128 (first period from anchor)", th)
	}
	next := time.Date(2026, 7, 27, 0, 0, 0, 0, time.Local)
	th2 := s.periodThresholdForRange(next)
	if th2 != calc.PeriodThreshold112 {
		t.Fatalf("threshold2=%v want 112", th2)
	}
}

func TestCalcTabDemoTESEndToEnd(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	settings := newSettingsTab()
	_ = settings.canvas()
	settings.applyDemoTES()
	settings.trainingEnabled.SetChecked(true)

	shifts := newShiftsTab(w)
	_ = shifts.canvas()
	shifts.loadDemoSeed()

	c := newCalcTab(settings, shifts)
	_ = c.canvas()
	c.setDemoRange()
	c.run()

	if c.err.Visible() {
		t.Fatalf("err=%q", c.err.Text)
	}
	if !strings.Contains(c.summary.Text, "Palkka yhteensä") {
		t.Fatalf("summary=%q", c.summary.Text)
	}
	for _, needle := range []string{"Iltalisä", "Yölisä", "Lauantai", "Pidennysylityö"} {
		if !strings.Contains(c.details.Text, needle) {
			t.Fatalf("details missing %q: %q", needle, c.details.Text)
		}
	}
	if !strings.Contains(c.summary.Text, "Jakso:") {
		t.Fatalf("period line missing: %q", c.summary.Text)
	}
}
