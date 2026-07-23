package calc

import (
	"testing"
	"time"
)

func TestCalloutFixedTwoHoursPayVartio(t *testing.T) {
	day := time.Date(2026, 7, 6, 0, 0, 0, 0, time.Local)
	in := PeriodInput{
		From: day,
		To:   day,
		Shifts: []Shift{{
			Start:   time.Date(2026, 7, 6, 8, 0, 0, 0, time.Local),
			End:     time.Date(2026, 7, 6, 16, 0, 0, 0, time.Local),
			Callout: true,
		}},
		Rates: Rates{Hourly: 10, Experience: 1, Personal: 0.5},
		Rules: VartiointiRules(),
	}
	in.Rules.PeriodOTEnabled = false
	out := Calculate(in)

	if out.CalloutHours != 2 {
		t.Fatalf("callout hours=%v want 2 (not full 8h shift)", out.CalloutHours)
	}
	wantPay := 2 * in.Rates.EffectiveHourly() // 2 * 11.5
	if out.CalloutPay != roundMoney(wantPay) {
		t.Fatalf("callout pay=%v want %v", out.CalloutPay, roundMoney(wantPay))
	}
	if out.BaseHours != 8 {
		t.Fatalf("base hours still worked=%v", out.BaseHours)
	}
	if out.TotalPay < out.BasePay+out.CalloutPay-0.01 {
		t.Fatalf("total should include callout: total=%v base=%v callout=%v", out.TotalPay, out.BasePay, out.CalloutPay)
	}
}

func TestCalloutFixedDisabledWhenZero(t *testing.T) {
	day := time.Date(2026, 7, 6, 0, 0, 0, 0, time.Local)
	rules := VartiointiRules()
	rules.CalloutFixedH = 0
	rules.PeriodOTEnabled = false
	out := Calculate(PeriodInput{
		From: day,
		To:   day,
		Shifts: []Shift{{
			Start:   time.Date(2026, 7, 6, 8, 0, 0, 0, time.Local),
			End:     time.Date(2026, 7, 6, 16, 0, 0, 0, time.Local),
			Callout: true,
		}},
		Rates: Rates{Hourly: 10},
		Rules: rules,
	})
	if out.CalloutHours != 0 || out.CalloutPay != 0 {
		t.Fatalf("no fixed callout: h=%v pay=%v", out.CalloutHours, out.CalloutPay)
	}
}
