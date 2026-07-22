package calc

import (
	"math"
	"testing"
)

func TestKaupanAllowancesAndDouble(t *testing.T) {
	eve, night, sat := KaupanAllowances(true)
	if eve != KaupanEveningPKS || night != KaupanNightPKS || sat != KaupanSaturdayPKS {
		t.Fatalf("PKS=%v/%v/%v", eve, night, sat)
	}
	eve, night, sat = KaupanAllowances(false)
	if eve != KaupanEveningMuu || night != KaupanNightMuu || sat != KaupanSaturdayMuu {
		t.Fatalf("Muu=%v/%v/%v", eve, night, sat)
	}
	if math.Abs(KaupanEveningDouble(KaupanEveningPKS)-8.36) > 0.001 {
		t.Fatalf("double=%v", KaupanEveningDouble(KaupanEveningPKS))
	}
}

func TestKaupanMyyjaRulesProfile(t *testing.T) {
	r := KaupanMyyjaRules()
	if r.EveningStartMin != 18*60 || r.EveningEndMin != 24*60 {
		t.Fatalf("evening=%d–%d", r.EveningStartMin, r.EveningEndMin)
	}
	if r.NightStartMin != 0 || r.NightEndMin != 6*60 {
		t.Fatalf("night=%d–%d", r.NightStartMin, r.NightEndMin)
	}
	if r.SaturdayStartMin != 13*60 || r.SaturdayEndMin != 24*60 {
		t.Fatalf("sat=%d–%d", r.SaturdayStartMin, r.SaturdayEndMin)
	}
	if !r.EveningExcludeSaturday || !r.NightExcludeSunday || !r.NightExcludeHoliday {
		t.Fatal("exclusions missing")
	}
	if r.EveningDoubleMonthFrom != 11 || r.EveningDoubleMonthTo != 12 || !r.EveningDoubleSundayOnly {
		t.Fatal("evening double window")
	}
	if r.Overtime50AfterH != 10 || r.ShiftOTAfterH != 0 || r.PeriodOTEnabled {
		t.Fatalf("OT model daily=%v shift=%v period=%v", r.Overtime50AfterH, r.ShiftOTAfterH, r.PeriodOTEnabled)
	}
	if !r.WeeklyOTEnabled || math.Abs(r.WeeklyOTThresholdH-KaupanWeeklyOTHours) > 0.001 {
		t.Fatalf("weekly=%v/%v", r.WeeklyOTEnabled, r.WeeklyOTThresholdH)
	}
}

func TestVartiointiRulesIsDefault(t *testing.T) {
	a, b := VartiointiRules(), DefaultRules()
	if a.ShiftOTAfterH != b.ShiftOTAfterH || a.EveningEndMin != b.EveningEndMin {
		t.Fatalf("DefaultRules must alias VartiointiRules: %+v vs %+v", a, b)
	}
	if a.ShiftOTAfterH != 12 || a.ShiftOT50CapH != 18 {
		t.Fatalf("vartio shift OT=%v/%v", a.ShiftOTAfterH, a.ShiftOT50CapH)
	}
}
