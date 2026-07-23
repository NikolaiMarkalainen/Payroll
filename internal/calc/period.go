package calc

import "time"

const (
	// PeriodDays is one kolmiviikkoisjakso.
	PeriodDays = 21

	// Standard period regular hours (TES 9 § / 29 §).
	PeriodThreshold120 = 120.0
	PeriodThreshold128 = 128.0
	PeriodThreshold112 = 112.0

	// DefaultPeriodOT50H: first N period OT hours at +50%.
	DefaultPeriodOT50H = 18.0

	// ShiftOTAfterHDefault: TES 31 § — over 12 h on a shift → like OT.
	ShiftOTAfterHDefault = 12.0

	// ShiftOT50CapHDefault: first N shift-extension OT hours in period at +50%, rest +100%.
	ShiftOT50CapHDefault = 18.0

	// VacationDayHours is TES vuosilomapäivän pituus (liite 3).
	VacationDayHours = 6.7
)

// PeriodThresholdAt returns the regular-hours threshold for the 3-week period
// that contains day, given an anchor start and which threshold the first period uses.
// Alternating 128↔112 averages 120 h over six weeks (TES tasoittumisjakso).
// If firstThreshold is 120 (or anything other than 112/128), every period is 120.
func PeriodThresholdAt(anchor, day time.Time, firstThreshold float64) float64 {
	a := truncateDay(anchor)
	d := truncateDay(day)
	if d.Before(a) {
		// Walk backwards in 21-day steps.
		days := int(a.Sub(d).Hours() / 24)
		idx := days / PeriodDays
		if days%PeriodDays != 0 {
			idx++
		}
		return thresholdForIndex(-(idx), firstThreshold)
	}
	days := int(d.Sub(a).Hours() / 24)
	idx := days / PeriodDays
	return thresholdForIndex(idx, firstThreshold)
}

func thresholdForIndex(idx int, first float64) float64 {
	if first != PeriodThreshold112 && first != PeriodThreshold128 {
		return PeriodThreshold120
	}
	// Normalize negative indices into alternating pair.
	mod := idx % 2
	if mod < 0 {
		mod += 2
	}
	if mod == 0 {
		return first
	}
	if first == PeriodThreshold128 {
		return PeriodThreshold112
	}
	return PeriodThreshold128
}

// PeriodWindow containing day, anchored at anchor (period 0 start).
func PeriodWindow(anchor, day time.Time) (from, to time.Time) {
	idx := PeriodIndexContaining(anchor, day)
	return PeriodIndexWindow(anchor, idx)
}

// PeriodIndexContaining returns the 0-based period index for day (0 = anchor's first period).
func PeriodIndexContaining(anchor, day time.Time) int {
	a := truncateDay(anchor)
	d := truncateDay(day)
	if d.Before(a) {
		days := int(a.Sub(d).Hours() / 24)
		idx := -(days / PeriodDays)
		if days%PeriodDays != 0 {
			idx--
		}
		return idx
	}
	days := int(d.Sub(a).Hours() / 24)
	return days / PeriodDays
}

// PeriodIndexWindow returns the [from,to] calendar days for period index
// (0 = first 21 days starting at anchor).
func PeriodIndexWindow(anchor time.Time, index int) (from, to time.Time) {
	a := truncateDay(anchor)
	from = a.AddDate(0, 0, index*PeriodDays)
	to = from.AddDate(0, 0, PeriodDays-1)
	return from, to
}

// splitPeriodOvertime: hours over threshold; first after50 at 50%, rest at 100%.
func splitPeriodOvertime(total, threshold, after50 float64) (ot50, ot100 float64) {
	if threshold < 0 {
		threshold = 0
	}
	if after50 < 0 {
		after50 = 0
	}
	if total <= threshold {
		return 0, 0
	}
	ot := total - threshold
	if ot <= after50 {
		return ot, 0
	}
	return after50, ot - after50
}
