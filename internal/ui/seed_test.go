package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func TestDemoRosterTotalMatchesSheet(t *testing.T) {
	shifts := demoRosterShifts(time.UTC)
	if len(shifts) != 10 {
		t.Fatalf("shifts=%d want 10", len(shifts))
	}
	var minutes int
	for _, sh := range shifts {
		s, err1 := clockToMinutes(sh.Start)
		e, err2 := clockToMinutes(sh.End)
		if err1 != nil || err2 != nil {
			t.Fatalf("bad time %s-%s", sh.Start, sh.End)
		}
		if e <= s {
			t.Fatalf("overnight not expected in demo: %+v", sh)
		}
		minutes += e - s
	}
	if minutes != 98*60+50 {
		t.Fatalf("total minutes=%d want %d (98:50)", minutes, 98*60+50)
	}
}

func TestLoadDemoSeed(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newShiftsTab(w)
	_ = s.canvas()
	s.loadDemoSeed()

	if s.month.Month() != time.July || s.month.Year() != 2026 {
		t.Fatalf("month=%v", s.month)
	}
	if len(s.shifts) != 10 {
		t.Fatalf("loaded=%d", len(s.shifts))
	}
}

func TestApplyDemoTESTasoIVPKS(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.applyDemoTES()

	if s.tesFamily.Selected != tesFamilyVartio {
		t.Fatalf("family=%q", s.tesFamily.Selected)
	}
	if s.tesLevel.Selected != "Taso IV" {
		t.Fatalf("level=%q", s.tesLevel.Selected)
	}
	if s.tesRegion.Selected != tesRegionPKS {
		t.Fatalf("region=%q", s.tesRegion.Selected)
	}
	if s.hourlyWage.Text != "13.97" || s.levelPay.Text != "2263.00" {
		t.Fatalf("pay=%q/%q", s.hourlyWage.Text, s.levelPay.Text)
	}
	if s.eveningAllowance.Text != "1.11" || s.nightAllowance.Text != "2.45" {
		t.Fatalf("ilta/yö=%q/%q", s.eveningAllowance.Text, s.nightAllowance.Text)
	}
	if s.trainingAllowance.Text != "0.25" {
		t.Fatalf("koulutuslisä=%q", s.trainingAllowance.Text)
	}
}

func TestVartiointiTasoIIIMuuSuomi(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.applyVartiointiPay("Taso III", false, tesServicePerus)
	if s.hourlyWage.Text != "12.67" || s.levelPay.Text != "2052.00" {
		t.Fatalf("got %s / %s", s.hourlyWage.Text, s.levelPay.Text)
	}
	if s.tesRegion.Selected != tesRegionMuu {
		t.Fatalf("region=%q", s.tesRegion.Selected)
	}
}

func TestOtherAllowanceModesInRates(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.otherAllowance.SetText("1.50")
	s.otherMode.SetSelected(otherModeNone)
	if s.rates().OtherHourly != 0 || s.rates().OtherFixed != 0 {
		t.Fatalf("none: %+v", s.rates())
	}
	s.otherMode.SetSelected(otherModeHourly)
	if s.rates().OtherHourly != 1.5 || s.rates().OtherFixed != 0 {
		t.Fatalf("hourly: %+v", s.rates())
	}
	s.otherMode.SetSelected(otherModeFixed)
	s.otherAllowance.SetText("40")
	if s.rates().OtherHourly != 0 || s.rates().OtherFixed != 40 {
		t.Fatalf("fixed: %+v", s.rates())
	}
}

func TestTrainingAllowanceInRates(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	// Koulutuslisä is Vartio-only; default Oma hides the section -> Training stays 0.
	s.applyVartiointiPay("Taso IV", true, tesServicePerus)
	s.hourlyWage.SetText("10.00")
	s.trainingAllowance.SetText("0.25")
	s.trainingEnabled.SetChecked(false)
	if s.rates().Training != 0 {
		t.Fatalf("disabled training=%v", s.rates().Training)
	}
	s.trainingEnabled.SetChecked(true)
	if s.rates().Training != 0.25 {
		t.Fatalf("enabled training=%v", s.rates().Training)
	}
}

func TestVartiointiExperience2yPKS(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.applyVartiointiPay("Taso IV", true, tesService2y)
	if s.hourlyWage.Text != "13.97" {
		t.Fatalf("base hourly=%q", s.hourlyWage.Text)
	}
	if s.experienceAllowance.Text != "0.51" {
		t.Fatalf("exp=%q", s.experienceAllowance.Text)
	}
	if s.levelPay.Text != "2345.00" {
		t.Fatalf("level=%q", s.levelPay.Text)
	}
	if s.sundayAllowance.Text != "14.48" {
		t.Fatalf("sunday=%q", s.sundayAllowance.Text)
	}
}
