package calc

// vartiointi.go — Vartiointialan TES (jaksotyö) default profile.
// Keep Vartio-specific thresholds here; Calculate stays TES-agnostic.

// VartiointiRules matches Turvallisuusala / Vartiointiala TES.
// TES 31 §: shift over 12 h → like OT at 50%; over 18 such hours in period → 100%.
// TES 31 § hälytystyö: kiinteä 2 h palkka per hälytysvuoro.
// TES 29 §: jaksoylityö defaults off until enabled in UI.
const VartiointiCalloutFixedH = 2.0

func VartiointiRules() Rules {
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
		CalloutFixedH:     VartiointiCalloutFixedH,
	}
}

// DefaultRules is an alias for VartiointiRules (historical default profile).
func DefaultRules() Rules {
	return VartiointiRules()
}
