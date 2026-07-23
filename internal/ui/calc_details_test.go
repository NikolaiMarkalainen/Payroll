package ui

import (
	"strings"
	"testing"

	"payroll/internal/calc"
)

func TestFormatCalcDetailsTableHasRateColumn(t *testing.T) {
	out := calc.Breakdown{
		BaseHours:     10,
		BasePay:       139.70,
		EveningHours:  4,
		EveningPay:    4.44,
		SaturdayHours: 8,
		SaturdayPay:   17.44,
		SundayHours:   8,
		SundayPay:     115.84,
		NightHours:    2,
		NightPay:      4.90,
		TotalPay:      282.32,
	}
	rates := calc.Rates{
		Hourly:   13.97,
		Evening:  1.11,
		Night:    2.45,
		Saturday: 2.18,
		Sunday:   14.48,
	}
	got := formatCalcDetails(out, calc.VartiointiRules(), rates)
	for _, needle := range []string{
		"Nimike", "Tunnit", "e/h", "Summa e",
		"Pohja", "13.97",
		"Iltatyölisä", "1.11",
		"Yötyölisä", "2.45",
		"Lauantailisä", "2.18", "17.44",
		"Sunnuntaikorotus 100%", "14.48",
		"Yhteensä",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("missing %q in:\n%s", needle, got)
		}
	}
	if strings.Contains(got, "Iltatyölisä 2x") {
		t.Fatalf("Vartiointi details must not show Iltatyölisä 2x:\n%s", got)
	}
	if issue := uiTextIssue(got); issue != "" {
		t.Fatalf("%s in details", issue)
	}
}

func TestFormatCalcDetailsPerehdytysRow(t *testing.T) {
	out := calc.Breakdown{PerehdytysHours: 0.58, PerehdytysPay: 0.58}
	rates := calc.Rates{Perehdytys: 1.00}
	got := formatCalcDetails(out, calc.VartiointiRules(), rates)
	for _, needle := range []string{"Perehdytyslisä", "0.58", "1.00"} {
		if !strings.Contains(got, needle) {
			t.Fatalf("missing %q in:\n%s", needle, got)
		}
	}
}

func TestFormatCalcDetailsShowsEveningDoubleForKauppa(t *testing.T) {
	out := calc.Breakdown{EveningHours: 4, EveningDoubleHours: 2, EveningPay: 20}
	rates := calc.Rates{Evening: 4.18, EveningDouble: 8.36}
	got := formatCalcDetails(out, calc.KaupanMyyjaRules(), rates)
	if !strings.Contains(got, "Iltatyölisä 2x") {
		t.Fatalf("Kaupan details should show Iltatyölisä 2x:\n%s", got)
	}
}
