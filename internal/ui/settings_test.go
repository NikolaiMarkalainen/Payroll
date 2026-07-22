package ui

import (
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
		t.Fatalf("evening=%s–%s", s.eveningStart.value(), s.eveningEnd.value())
	}
	if s.nightStart.value() != "22:00" || s.nightEnd.value() != "06:00" {
		t.Fatalf("night=%s–%s", s.nightStart.value(), s.nightEnd.value())
	}
	if s.overtime50After.Text != "12" || s.overtime100After.Text != "18" {
		t.Fatalf("overtime thresholds=%q/%q", s.overtime50After.Text, s.overtime100After.Text)
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

	if s.saveBtn == nil {
		t.Fatal("save button missing")
	}
	test.Tap(s.saveBtn)

	if !strings.Contains(s.status.Text, "tallennettu") {
		t.Fatalf("status=%q", s.status.Text)
	}
}
