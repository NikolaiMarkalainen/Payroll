package calc

import (
	"math"
	"testing"
	"time"
)

func TestPerehdytysPayFromShiftHours(t *testing.T) {
	loc := time.UTC
	day := time.Date(2026, 6, 20, 0, 0, 0, 0, loc)
	got := Calculate(PeriodInput{
		From: day, To: day,
		Shifts: []Shift{
			{
				Start:           time.Date(2026, 6, 20, 6, 0, 0, 0, loc),
				End:             time.Date(2026, 6, 20, 14, 0, 0, 0, loc),
				PerehdytysHours: 0.58,
			},
		},
		Rates: Rates{Hourly: 13.97, Perehdytys: 1.00},
		Rules: VartiointiRules(),
	})
	if math.Abs(got.PerehdytysHours-0.58) > 0.001 {
		t.Fatalf("hours=%v want 0.58", got.PerehdytysHours)
	}
	if math.Abs(got.PerehdytysPay-0.58) > 0.001 {
		t.Fatalf("pay=%v want 0.58", got.PerehdytysPay)
	}
	if got.TotalPay < got.BasePay+got.PerehdytysPay-0.01 {
		t.Fatalf("total missing perehdytys: total=%v base=%v pere=%v", got.TotalPay, got.BasePay, got.PerehdytysPay)
	}
}

func TestPerehdytysNotInEffectiveHourly(t *testing.T) {
	r := Rates{Hourly: 10, Experience: 1, Personal: 0.5, Perehdytys: 1}
	if math.Abs(r.EffectiveHourly()-11.5) > 0.001 {
		t.Fatalf("eff=%v", r.EffectiveHourly())
	}
}
