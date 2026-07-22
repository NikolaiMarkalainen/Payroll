package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestTESFamilyMenuContainsAllAgreements(t *testing.T) {
	// Regressio: uusi TES koodissa ei riitä - sen pitää olla TES-valikossa.
	names := tesFamilyNames()
	for _, want := range []string{tesFamilyCustom, tesFamilyVartio, tesFamilyKauppa} {
		if !containsString(names, want) {
			t.Fatalf("tesFamilyNames missing %q; got %v", want, names)
		}
	}

	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	if s.tesFamily == nil {
		t.Fatal("tesFamily select missing")
	}
	for _, want := range []string{tesFamilyCustom, tesFamilyVartio, tesFamilyKauppa} {
		if !containsString(s.tesFamily.Options, want) {
			t.Fatalf("TES-valikko missing %q; options=%v", want, s.tesFamily.Options)
		}
	}
}

func TestApplyTESFamilyViaMenuSwitchesKaupan(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	// Real menu path: Select updates Selected then fires OnChanged -> applyTESFamily.
	s.tesFamily.SetSelected(tesFamilyKauppa)
	if s.tesFamily.Selected != tesFamilyKauppa {
		t.Fatalf("selected=%q", s.tesFamily.Selected)
	}
	if !containsString(s.tesLevel.Options, "B") || len(s.tesLevel.Options) != 4 {
		t.Fatalf("kaupan groups not in level menu: %v", s.tesLevel.Options)
	}
	if s.levelFormItem == nil || s.levelFormItem.Text != "Palkkaryhmä" {
		t.Fatalf("label=%v", s.levelFormItem)
	}
	if s.hourlyWage.Text == "" || s.hourlyWage.Text == "0.00" {
		t.Fatalf("Kaupan pay not applied via menu: hourly=%q", s.hourlyWage.Text)
	}
}

func TestFeaturesForFamily(t *testing.T) {
	v := featuresForFamily(tesFamilyVartio)
	if v.LevelLabel != "Taso" || !v.Training || !v.PeriodOT || !v.ShiftOT {
		t.Fatalf("vartio features=%+v", v)
	}
	if v.EveningDouble || v.WeeklyOT || v.KaupanFlags {
		t.Fatalf("vartio should hide Kaupan-only UI: %+v", v)
	}

	k := featuresForFamily(tesFamilyKauppa)
	if k.LevelLabel != "Palkkaryhmä" || !k.EveningDouble || !k.WeeklyOT || !k.KaupanFlags {
		t.Fatalf("kauppa features=%+v", k)
	}
	if k.Training || k.PeriodOT || k.ShiftOT {
		t.Fatalf("kauppa should hide Vartio-only UI: %+v", k)
	}

	o := featuresForFamily(tesFamilyCustom)
	// Oma is barebones: fixed 120 h period only, no TES profile selectors / Kaupan flags.
	if o.Training || o.EveningDouble || o.WeeklyOT || o.ShiftOT || o.KaupanFlags || o.ProfileSelectors {
		t.Fatalf("oma should be barebones: %+v", o)
	}
	if !o.PeriodOT || !o.PeriodFixed120 {
		t.Fatalf("oma should keep fixed period OT: %+v", o)
	}
}

func TestProfileSelectorsForFamily(t *testing.T) {
	k := profileSelectorsForFamily(tesFamilyKauppa)
	if len(k.Levels) != 4 || k.DefaultLevel != "B" || k.DefaultService != kaupanService2y {
		t.Fatalf("kauppa selectors=%+v", k)
	}
	v := profileSelectorsForFamily(tesFamilyVartio)
	if len(v.Levels) != 7 || v.DefaultLevel != "Taso IV" || v.DefaultService != tesServicePerus {
		t.Fatalf("vartio selectors=%+v", v)
	}
	c := profileSelectorsForFamily(tesFamilyCustom)
	if c.KeepCurrentOpts || len(c.Levels) != 0 || len(c.Services) != 0 {
		t.Fatalf("oma should clear selectors: %+v", c)
	}
}

func TestSyncTESVisibilityKaupanVsVartio(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()

	s.applyKaupanPay("B", true, kaupanService2y)
	if s.levelFormItem == nil || s.levelFormItem.Text != "Palkkaryhmä" {
		t.Fatalf("label=%v", s.levelFormItem)
	}
	assertSectionVisible(t, "training", s.trainingSection, false)
	assertSectionVisible(t, "eveningDouble", s.eveningDoubleSection, true)
	assertSectionVisible(t, "weeklyOT", s.weeklyOTSection, true)
	assertSectionVisible(t, "periodOT", s.periodOTSection, false)
	assertSectionVisible(t, "shiftOT", s.shiftOTSection, false)
	assertSectionVisible(t, "kaupanFlags", s.kaupanFlagsSection, true)
	if s.personalHeading != nil && s.personalHeading.Text != "Kokemus- ja henkilökohtainen lisä" {
		t.Fatalf("heading=%q", s.personalHeading.Text)
	}

	s.applyVartiointiPay("Taso IV", true, tesServicePerus)
	if s.levelFormItem.Text != "Taso" {
		t.Fatalf("label=%q", s.levelFormItem.Text)
	}
	assertSectionVisible(t, "training", s.trainingSection, true)
	assertSectionVisible(t, "eveningDouble", s.eveningDoubleSection, false)
	assertSectionVisible(t, "weeklyOT", s.weeklyOTSection, false)
	assertSectionVisible(t, "periodOT", s.periodOTSection, true)
	assertSectionVisible(t, "shiftOT", s.shiftOTSection, true)
	assertSectionVisible(t, "kaupanFlags", s.kaupanFlagsSection, false)
}

func TestSyncTESVisibilityOmaBarebones(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()
	s.tesFamily.SetSelected(tesFamilyCustom)

	assertSectionVisible(t, "profileSelectors", s.profileSelectorsSection, false)
	assertSectionVisible(t, "training", s.trainingSection, false)
	assertSectionVisible(t, "eveningDouble", s.eveningDoubleSection, false)
	assertSectionVisible(t, "weeklyOT", s.weeklyOTSection, false)
	assertSectionVisible(t, "shiftOT", s.shiftOTSection, false)
	assertSectionVisible(t, "kaupanFlags", s.kaupanFlagsSection, false)
	assertSectionVisible(t, "periodOT", s.periodOTSection, true)
	assertSectionVisible(t, "periodAdvanced", s.periodAdvancedSection, false)
	if s.periodMode.Selected != periodMode120 {
		t.Fatalf("period mode=%q want %q", s.periodMode.Selected, periodMode120)
	}
	if !s.periodOTEnabled.Checked {
		t.Fatal("period OT should be on for Oma")
	}
}

func assertSectionVisible(t *testing.T, name string, o interface{ Visible() bool }, want bool) {
	t.Helper()
	if o == nil {
		t.Fatalf("%s section nil", name)
	}
	if o.Visible() != want {
		t.Fatalf("%s visible=%v want %v", name, o.Visible(), want)
	}
}
