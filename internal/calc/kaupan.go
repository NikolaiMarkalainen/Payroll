package calc

// kaupan.go — Kaupan TES (myyjät) profile and quirks.
//
// Wire UI via KaupanMyyjaRules() / allowance rates; do not add "if tes == Kauppa"
// branches into pay.go. New Kaupan-only behaviour lands in this file first
// (Rules flags or helpers), then Calculate uses them generically.

const (
	// KaupanWeeklyOTHours is regular weekly hours before +50% (myyjät).
	KaupanWeeklyOTHours = 37.5

	// Allowance €/h (1.5.2022 → 31.1.2028), myyjät.
	KaupanEveningPKS  = 4.18
	KaupanEveningMuu  = 4.00
	KaupanNightPKS    = 6.28
	KaupanNightMuu    = 6.01
	KaupanSaturdayPKS = 5.46
	KaupanSaturdayMuu = 5.27
)

// KaupanMyyjaRules is the first-cut seller Rules profile from TES-haku.
//
// Quirks encoded as Rules (not pay.go switches):
//   - la-lisä only 13–24; no evening on Saturday
//   - night 00–06, not on Sunday/holiday
//   - evening 18–24; Nov–Dec Sunday evening uses EveningDouble rate
//   - calendar-day OT >10 h @50%; weekly OT >37.5 h @50%; no jaksoylityö / shift-OT-12
func KaupanMyyjaRules() Rules {
	return Rules{
		EveningStartMin:         18 * 60,
		EveningEndMin:           24 * 60,
		NightStartMin:           0,
		NightEndMin:             6 * 60,
		SaturdayStartMin:        13 * 60,
		SaturdayEndMin:          24 * 60,
		EveningExcludeSaturday:  true,
		NightExcludeSunday:      true,
		NightExcludeHoliday:     true,
		EveningDoubleMonthFrom:  11,
		EveningDoubleMonthTo:    12,
		EveningDoubleSundayOnly: true,
		Overtime50AfterH:        10,
		Overtime100AfterH:       10, // daily excess all at 50%
		ShiftOTAfterH:           0,
		WeeklyOTEnabled:         true,
		WeeklyOTThresholdH:      KaupanWeeklyOTHours,
		PeriodOTEnabled:         false,
	}
}

// KaupanEveningDouble returns 2× evening €/h (marras–joulu su -ilta).
func KaupanEveningDouble(evening float64) float64 {
	return evening * 2
}

// KaupanAllowances returns evening/night/saturday €/h for region.
func KaupanAllowances(pks bool) (evening, night, saturday float64) {
	if pks {
		return KaupanEveningPKS, KaupanNightPKS, KaupanSaturdayPKS
	}
	return KaupanEveningMuu, KaupanNightMuu, KaupanSaturdayMuu
}
