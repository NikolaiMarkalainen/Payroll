package ui

import (
	"os"
	"strings"
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestSettingsDefaults(t *testing.T) {
	test.NewApp()

	s := newSettingsTab()
	_ = s.canvas()

	if s.dailyRestHours.Text != "10" {
		t.Fatalf("daily rest=%q", s.dailyRestHours.Text)
	}
	if s.restViolationPercent.Text != "50" {
		t.Fatalf("rest %%=%q", s.restViolationPercent.Text)
	}
	if s.eveningStart.value() != "18:00" || s.eveningEnd.value() != "22:00" {
		t.Fatalf("evening=%s-%s", s.eveningStart.value(), s.eveningEnd.value())
	}
	if s.nightStart.value() != "22:00" || s.nightEnd.value() != "06:00" {
		t.Fatalf("night=%s-%s", s.nightStart.value(), s.nightEnd.value())
	}
	if s.overtime50After.Text != "12" || s.overtime100After.Text != "18" {
		t.Fatalf("overtime thresholds=%q/%q", s.overtime50After.Text, s.overtime100After.Text)
	}
	if s.colorShiftTitles == nil || !s.colorShiftTitles.Checked {
		t.Fatal("shift title coloring should default on")
	}
	rules := s.allowanceRules()
	if rules.overtime50AfterH != 12 || rules.overtime100AfterH != 18 {
		t.Fatalf("rules OT=%v/%v", rules.overtime50AfterH, rules.overtime100AfterH)
	}
}

func TestSettingsPayFieldsEditable(t *testing.T) {
	test.NewApp()

	s := newSettingsTab()
	s.hourlyWage.SetText("12.74")
	s.eveningAllowance.SetText("1.11")
	s.nightAllowance.SetText("2.50")

	if s.hourlyWage.Text != "12.74" {
		t.Fatalf("hourly=%q", s.hourlyWage.Text)
	}
	if s.eveningAllowance.Text != "1.11" || s.nightAllowance.Text != "2.50" {
		t.Fatalf("allowances evening=%q night=%q", s.eveningAllowance.Text, s.nightAllowance.Text)
	}
}

func TestSettingsSaveUpdatesStatus(t *testing.T) {
	test.NewApp()

	s := newSettingsTab()
	content := s.canvas()
	w := test.NewWindow(content)
	defer w.Close()

	dir := t.TempDir()
	shifts := newShiftsTab(nil)
	calc := newCalcTab(s, shifts)
	p := newAppPersister(dir, s, shifts, calc)
	wirePersistence(s, shifts, calc, p)

	if s.saveBtn == nil {
		t.Fatal("save button missing")
	}
	test.Tap(s.saveBtn)

	if !strings.Contains(s.status.Text, "Tallennettu:") {
		t.Fatalf("status=%q", s.status.Text)
	}
	if _, err := os.Stat(statePath(dir)); err != nil {
		t.Fatalf("state file missing: %v", err)
	}
}

func TestPersistRestoresTESFamily(t *testing.T) {
	test.NewApp()
	dir := t.TempDir()

	s := newSettingsTab()
	_ = s.canvas()
	s.tesFamily.SetSelected(tesFamilyVartio)
	s.applyVartiointiPay("Taso III", true, tesService7y)

	shifts := newShiftsTab(nil)
	p := newAppPersister(dir, s, shifts, nil)
	if err := p.saveNow(); err != nil {
		t.Fatal(err)
	}

	s2 := newSettingsTab()
	_ = s2.canvas()
	if s2.tesFamily.Selected != tesFamilyCustom {
		t.Fatalf("precondition: default family=%q", s2.tesFamily.Selected)
	}
	p2 := newAppPersister(dir, s2, newShiftsTab(nil), nil)
	if err := p2.load(); err != nil {
		t.Fatal(err)
	}
	if s2.tesFamily.Selected != tesFamilyVartio {
		t.Fatalf("family=%q want Vartio", s2.tesFamily.Selected)
	}
	if s2.tesLevel.Selected != "Taso III" {
		t.Fatalf("level=%q", s2.tesLevel.Selected)
	}
	if s2.experienceTier.Selected != tesService7y {
		t.Fatalf("service=%q", s2.experienceTier.Selected)
	}
}

func TestSettingsShiftColorSection(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	s := newSettingsTab()
	s.window = w
	_ = s.canvas()

	if s.colorShiftTitles == nil || !s.colorShiftTitles.Checked {
		t.Fatal("coloring toggle missing/off")
	}
	if s.colorRows == nil {
		t.Fatal("color rows missing")
	}
	if len(s.colorRows.Objects) != 0 {
		t.Fatalf("no calendar shifts → 0 color groups, got %d", len(s.colorRows.Objects))
	}
}
