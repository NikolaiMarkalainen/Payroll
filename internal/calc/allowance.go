package calc

import "time"

// allowance.go — day-level allowance classification (ilta / yö / la / su / pyhä).
// TES-specific quirks are configured via Rules (see kaupan.go, vartiointi.go), not if-branches here.

func addDaySlice(dh *DayHours, start, end, dayStart time.Time, callout bool, rules Rules, holidays map[string]string) {
	hours := end.Sub(start).Hours()
	dh.Total += hours
	if callout {
		dh.Callout += hours
	}
	switch dayStart.Weekday() {
	case time.Saturday:
		dh.Saturday += saturdayAllowanceHours(start, end, dayStart, rules)
	case time.Sunday:
		dh.Sunday += hours
	}
	if IsHoliday(dayStart, holidays) {
		dh.Holiday += hours
	}
	eve, night := splitEveningNight(start, end, dayStart, rules)
	if rules.EveningExcludeSaturday && dayStart.Weekday() == time.Saturday {
		eve = 0
	}
	if rules.NightExcludeSunday && dayStart.Weekday() == time.Sunday {
		night = 0
	}
	if rules.NightExcludeHoliday && IsHoliday(dayStart, holidays) {
		night = 0
	}
	dh.Evening += eve
	dh.Night += night
	if eve > 0 && eveningDoubleApplies(dayStart, rules) {
		dh.EveningDouble += eve
	}
}

func saturdayAllowanceHours(start, end, dayStart time.Time, rules Rules) float64 {
	if rules.SaturdayStartMin == 0 && rules.SaturdayEndMin == 0 {
		return end.Sub(start).Hours()
	}
	span0 := dayStart.Add(time.Duration(rules.SaturdayStartMin) * time.Minute)
	span1 := dayStart.Add(time.Duration(rules.SaturdayEndMin) * time.Minute)
	return overlapHours(start, end, [2]time.Time{span0, span1})
}

func eveningDoubleApplies(dayStart time.Time, rules Rules) bool {
	from, to := rules.EveningDoubleMonthFrom, rules.EveningDoubleMonthTo
	if from == 0 && to == 0 {
		return false
	}
	if from <= 0 {
		from = 1
	}
	if to <= 0 {
		to = 12
	}
	m := int(dayStart.Month())
	inMonth := false
	if from <= to {
		inMonth = m >= from && m <= to
	} else {
		inMonth = m >= from || m <= to
	}
	if !inMonth {
		return false
	}
	if rules.EveningDoubleSundayOnly && dayStart.Weekday() != time.Sunday {
		return false
	}
	return true
}

func splitEveningNight(start, end, dayStart time.Time, rules Rules) (evening, night float64) {
	eve0 := dayStart.Add(time.Duration(rules.EveningStartMin) * time.Minute)
	eve1 := dayStart.Add(time.Duration(rules.EveningEndMin) * time.Minute)

	var nightSpans [][2]time.Time
	if rules.NightEndMin <= rules.NightStartMin {
		nightSpans = [][2]time.Time{
			{dayStart, dayStart.Add(time.Duration(rules.NightEndMin) * time.Minute)},
			{dayStart.Add(time.Duration(rules.NightStartMin) * time.Minute), dayStart.Add(24 * time.Hour)},
		}
	} else {
		nightSpans = [][2]time.Time{{
			dayStart.Add(time.Duration(rules.NightStartMin) * time.Minute),
			dayStart.Add(time.Duration(rules.NightEndMin) * time.Minute),
		}}
	}

	night = overlapHours(start, end, nightSpans...)
	eveSpans := [][2]time.Time{{eve0, eve1}}
	eveningRaw := overlapHours(start, end, eveSpans...)
	evening = eveningRaw - overlapHours(start, end, intersectSpans(eveSpans, nightSpans)...)
	if evening < 0 {
		evening = 0
	}
	return evening, night
}

func overlapHours(start, end time.Time, spans ...[2]time.Time) float64 {
	var total float64
	for _, sp := range spans {
		a, b := sp[0], sp[1]
		if !a.Before(b) {
			continue
		}
		cs, ce := start, end
		if cs.Before(a) {
			cs = a
		}
		if ce.After(b) {
			ce = b
		}
		if cs.Before(ce) {
			total += ce.Sub(cs).Hours()
		}
	}
	return total
}

func intersectSpans(a, b [][2]time.Time) [][2]time.Time {
	var out [][2]time.Time
	for _, x := range a {
		for _, y := range b {
			cs, ce := x[0], x[1]
			if cs.Before(y[0]) {
				cs = y[0]
			}
			if ce.After(y[1]) {
				ce = y[1]
			}
			if cs.Before(ce) {
				out = append(out, [2]time.Time{cs, ce})
			}
		}
	}
	return out
}
