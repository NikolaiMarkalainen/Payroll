package calc

import "time"

// Rates are €/h inputs used in pay math.
type Rates struct {
	Hourly      float64
	Experience  float64 // kokemus- / palvelusaikalisä €/h
	Personal    float64 // henkilökohtainen palkanosa €/h
	Training    float64 // koulutuslisä €/h (TES 25 §; not in OT base)
	OtherHourly float64 // muu lisä €/h (not in OT base)
	OtherFixed  float64 // muu lisä kiinteä € / laskentajakso (not in OT base)
	Evening     float64
	Night       float64
	Saturday    float64
	Sunday      float64
	Holiday     float64
}

// EffectiveHourly is tehtäväkohtainen + kokemus + henkilökohtainen.
// Koulutuslisä is excluded (TES 35 §: not part of perustuntipalkka for OT).
func (r Rates) EffectiveHourly() float64 {
	return r.Hourly + r.Experience + r.Personal
}

// Rules are time windows and overtime thresholds.
type Rules struct {
	EveningStartMin   int     // e.g. 18*60
	EveningEndMin     int     // e.g. 22*60
	NightStartMin     int     // e.g. 22*60
	NightEndMin       int     // e.g. 6*60 (crosses midnight when <= NightStartMin)
	Overtime50AfterH  float64 // legacy calendar-day OT 50% band (unused when ShiftOTAfterH > 0)
	Overtime100AfterH float64 // legacy calendar-day OT 100% after (unused when ShiftOTAfterH > 0)

	// ShiftOTAfterH (TES 31 §): continuous shift hours over this earn "kuten ylityöstä".
	// Typically 12. First ShiftOT50CapH of those hours in the window get +50%, rest +100%.
	// Those hours are excluded from jaksoylityö (TES 31 §).
	ShiftOTAfterH float64
	ShiftOT50CapH float64 // typically 18 (same as period OT 50% band)

	// Period OT (TES 29 § jaksoylityö). When PeriodOTEnabled, hours over PeriodThresholdH
	// in the calculation window are period overtime. First PeriodOT50AfterH of those
	// get +50%, the rest +100%. Absences credited via PeriodInput.CreditedAbsenceH.
	PeriodOTEnabled  bool
	PeriodThresholdH float64 // 120, 128 or 112
	PeriodOT50AfterH float64 // typically 18
}

// DefaultRules matches Vartiointialan TES (jaksotyö).
// TES 31 §: shift over 12 h → like OT at 50%; over 18 such hours in period → 100%.
// TES 29 §: jaksoylityö defaults off until enabled.
func DefaultRules() Rules {
	return Rules{
		EveningStartMin:   18 * 60,
		EveningEndMin:     22 * 60,
		NightStartMin:     22 * 60,
		NightEndMin:       6 * 60,
		Overtime50AfterH:  12,
		Overtime100AfterH: 12,
		ShiftOTAfterH:     12,
		ShiftOT50CapH:     18,
		PeriodOTEnabled:   false,
		PeriodThresholdH:  120,
		PeriodOT50AfterH:  18,
	}
}

// Shift is one worked interval in absolute time.
type Shift struct {
	Start   time.Time
	End     time.Time
	Callout bool
}

// DayHours is classified hours for one calendar day.
type DayHours struct {
	Date        time.Time
	Total       float64
	Evening     float64
	Night       float64
	Saturday    float64
	Sunday      float64
	Holiday     float64
	Callout     float64
	Overtime50  float64
	Overtime100 float64
	HolidayName string
}

// Breakdown is the money result for a period.
type Breakdown struct {
	Days []DayHours

	BaseHours        float64
	EveningHours     float64
	NightHours       float64
	SaturdayHours    float64
	SundayHours      float64
	HolidayHours     float64
	CalloutHours     float64
	Overtime50Hours  float64 // pidennys / shift OT 50 %
	Overtime100Hours float64 // pidennys / shift OT 100 %
	ShiftOTHours     float64 // hours over ShiftOTAfterH (excluded from jaksoylityö)

	// Period OT (jaksoylityö) — separate from daily OT.
	PeriodWorkedHours   float64 // worked hours in window
	PeriodCreditedHours float64 // absences counted as work
	PeriodTotalHours    float64 // worked + credited
	PeriodThresholdH    float64
	PeriodOTHours       float64 // total period OT
	PeriodOT50Hours     float64
	PeriodOT100Hours    float64

	BasePay        float64
	ExperiencePay  float64
	PersonalPay    float64
	TrainingPay    float64
	OtherPay       float64
	EveningPay     float64
	NightPay       float64
	SaturdayPay    float64
	SundayPay      float64
	HolidayPay     float64
	Overtime50Pay  float64
	Overtime100Pay float64
	PeriodOT50Pay  float64
	PeriodOT100Pay float64
	TotalPay       float64
}

// PeriodInput is everything needed to compute pay for a date range.
type PeriodInput struct {
	From     time.Time // inclusive calendar day
	To       time.Time // inclusive calendar day
	Shifts   []Shift
	Rates    Rates
	Rules    Rules
	Holidays map[string]string // optional; built from From–To if nil

	// CreditedAbsenceH is absence time equated to work for period OT
	// (vuosiloma 6.7 h/day, arkipyhä, vuosivapaa, sairaus, …).
	CreditedAbsenceH float64
}
