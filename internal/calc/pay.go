package calc

import (
	"math"
	"time"
)

// Calculate computes TES-style pay for shifts overlapping [From, To] (calendar days).
//
// Model (session math, not a full official TES parser):
//   - Every worked minute on the period earns base hourly wage.
//   - Evening / night / Sat / Sun / holiday allowances are additive €/h on matching minutes.
//   - Night wins over evening on overlap.
//   - Shift extension OT (TES 31 §): when ShiftOTAfterH > 0, continuous shift hours over
//     that threshold (default 12) are paid like overtime — first ShiftOT50CapH (18) at +50%,
//     rest at +100%. Those hours are excluded from jaksoylityö.
//   - Legacy calendar-day OT: used only when ShiftOTAfterH == 0.
//   - Period overtime (TES 29 §): when enabled, (worked − shift-OT + credited absences) over
//     PeriodThresholdH; first PeriodOT50AfterH hours at +50%, rest at +100%.
//   - Kokemuslisä and henkilökohtainen lisä are paid on all worked hours (€/h).
//   - Koulutuslisä (TES 25 §) is paid on all worked hours but is not part of OT base.
//   - Overtime %-premium uses effective hourly (base + kokemus + henkilökohtainen).
//   - Callout hours are tracked but do not add a separate € rate yet (no setting).
func Calculate(in PeriodInput) Breakdown {
	rules := in.Rules
	if rules.Overtime50AfterH == 0 && rules.Overtime100AfterH == 0 &&
		rules.EveningStartMin == 0 && rules.EveningEndMin == 0 &&
		rules.ShiftOTAfterH == 0 {
		rules = DefaultRules()
	}
	if rules.PeriodOTEnabled {
		if rules.PeriodThresholdH <= 0 {
			rules.PeriodThresholdH = PeriodThreshold120
		}
		if rules.PeriodOT50AfterH <= 0 {
			rules.PeriodOT50AfterH = DefaultPeriodOT50H
		}
	}
	if rules.ShiftOTAfterH > 0 && rules.ShiftOT50CapH <= 0 {
		rules.ShiftOT50CapH = ShiftOT50CapHDefault
	}
	holidays := in.Holidays
	if holidays == nil {
		holidays = HolidaySetRange(in.From, in.To)
	}

	fromDay := truncateDay(in.From)
	toDay := truncateDay(in.To)
	periodStart := fromDay
	periodEnd := toDay.Add(24 * time.Hour)

	// Accumulate per calendar day first.
	byDay := map[string]*DayHours{}

	for _, sh := range in.Shifts {
		if !sh.Start.Before(sh.End) {
			continue
		}
		// Clip shift to period.
		start, end := sh.Start, sh.End
		if start.Before(periodStart) {
			start = periodStart
		}
		if end.After(periodEnd) {
			end = periodEnd
		}
		if !start.Before(end) {
			continue
		}

		for day := truncateDay(start); day.Before(end); day = day.AddDate(0, 0, 1) {
			dayEnd := day.Add(24 * time.Hour)
			cs, ce := start, end
			if cs.Before(day) {
				cs = day
			}
			if ce.After(dayEnd) {
				ce = dayEnd
			}
			if !cs.Before(ce) {
				continue
			}
			key := Key(day)
			dh := byDay[key]
			if dh == nil {
				dh = &DayHours{Date: day}
				if name, ok := holidays[key]; ok {
					dh.HolidayName = name
				}
				byDay[key] = dh
			}
			addDaySlice(dh, cs, ce, day, sh.Callout, rules, holidays)
		}
	}

	var out Breakdown
	for d := fromDay; !d.After(toDay); d = d.AddDate(0, 0, 1) {
		dh := byDay[Key(d)]
		if dh == nil {
			continue
		}
		if rules.ShiftOTAfterH <= 0 {
			dh.Overtime50, dh.Overtime100 = splitOvertime(dh.Total, rules.Overtime50AfterH, rules.Overtime100AfterH)
		}
		roundDay(dh)
		out.Days = append(out.Days, *dh)

		out.BaseHours += dh.Total
		out.EveningHours += dh.Evening
		out.NightHours += dh.Night
		out.SaturdayHours += dh.Saturday
		out.SundayHours += dh.Sunday
		out.HolidayHours += dh.Holiday
		out.CalloutHours += dh.Callout
		out.Overtime50Hours += dh.Overtime50
		out.Overtime100Hours += dh.Overtime100
	}

	if rules.ShiftOTAfterH > 0 {
		ext := collectShiftExtensionHours(in.Shifts, periodStart, periodEnd, rules.ShiftOTAfterH)
		out.ShiftOTHours = roundHours(ext.total)
		out.Overtime50Hours, out.Overtime100Hours = splitPeriodOvertime(ext.total, 0, rules.ShiftOT50CapH)
		out.Overtime50Hours = roundHours(out.Overtime50Hours)
		out.Overtime100Hours = roundHours(out.Overtime100Hours)
		applyExtensionOTToDays(out.Days, ext.slices, rules.ShiftOT50CapH)
	}

	if rules.PeriodOTEnabled {
		credited := in.CreditedAbsenceH
		if credited < 0 {
			credited = 0
		}
		// TES 31 §: hours paid like OT (shift extension) are not counted in jaksoylityö.
		workedForPeriod := out.BaseHours - out.ShiftOTHours
		if workedForPeriod < 0 {
			workedForPeriod = 0
		}
		out.PeriodWorkedHours = roundHours(workedForPeriod)
		out.PeriodCreditedHours = roundHours(credited)
		out.PeriodTotalHours = roundHours(out.PeriodWorkedHours + out.PeriodCreditedHours)
		out.PeriodThresholdH = rules.PeriodThresholdH
		out.PeriodOT50Hours, out.PeriodOT100Hours = splitPeriodOvertime(
			out.PeriodTotalHours, rules.PeriodThresholdH, rules.PeriodOT50AfterH,
		)
		out.PeriodOTHours = roundHours(out.PeriodOT50Hours + out.PeriodOT100Hours)
		out.PeriodOT50Hours = roundHours(out.PeriodOT50Hours)
		out.PeriodOT100Hours = roundHours(out.PeriodOT100Hours)
	}

	r := in.Rates
	eff := r.EffectiveHourly()
	out.BasePay = out.BaseHours * r.Hourly
	out.ExperiencePay = out.BaseHours * r.Experience
	out.PersonalPay = out.BaseHours * r.Personal
	out.TrainingPay = out.BaseHours * r.Training
	out.OtherPay = out.BaseHours*r.OtherHourly + r.OtherFixed
	out.EveningPay = out.EveningHours * r.Evening
	out.NightPay = out.NightHours * r.Night
	out.SaturdayPay = out.SaturdayHours * r.Saturday
	out.SundayPay = out.SundayHours * r.Sunday
	out.HolidayPay = out.HolidayHours * r.Holiday
	out.Overtime50Pay = out.Overtime50Hours * eff * 0.5
	out.Overtime100Pay = out.Overtime100Hours * eff * 1.0
	out.PeriodOT50Pay = out.PeriodOT50Hours * eff * 0.5
	out.PeriodOT100Pay = out.PeriodOT100Hours * eff * 1.0
	out.TotalPay = roundMoney(
		out.BasePay + out.ExperiencePay + out.PersonalPay + out.TrainingPay + out.OtherPay +
			out.EveningPay + out.NightPay + out.SaturdayPay +
			out.SundayPay + out.HolidayPay + out.Overtime50Pay + out.Overtime100Pay +
			out.PeriodOT50Pay + out.PeriodOT100Pay,
	)

	out.BaseHours = roundHours(out.BaseHours)
	out.EveningHours = roundHours(out.EveningHours)
	out.NightHours = roundHours(out.NightHours)
	out.SaturdayHours = roundHours(out.SaturdayHours)
	out.SundayHours = roundHours(out.SundayHours)
	out.HolidayHours = roundHours(out.HolidayHours)
	out.CalloutHours = roundHours(out.CalloutHours)
	out.Overtime50Hours = roundHours(out.Overtime50Hours)
	out.Overtime100Hours = roundHours(out.Overtime100Hours)

	out.BasePay = roundMoney(out.BasePay)
	out.ExperiencePay = roundMoney(out.ExperiencePay)
	out.PersonalPay = roundMoney(out.PersonalPay)
	out.TrainingPay = roundMoney(out.TrainingPay)
	out.OtherPay = roundMoney(out.OtherPay)
	out.EveningPay = roundMoney(out.EveningPay)
	out.NightPay = roundMoney(out.NightPay)
	out.SaturdayPay = roundMoney(out.SaturdayPay)
	out.SundayPay = roundMoney(out.SundayPay)
	out.HolidayPay = roundMoney(out.HolidayPay)
	out.Overtime50Pay = roundMoney(out.Overtime50Pay)
	out.Overtime100Pay = roundMoney(out.Overtime100Pay)
	out.PeriodOT50Pay = roundMoney(out.PeriodOT50Pay)
	out.PeriodOT100Pay = roundMoney(out.PeriodOT100Pay)

	return out
}

type extSlice struct {
	start time.Time
	end   time.Time
}

type extHours struct {
	total  float64
	slices []extSlice
}

// collectShiftExtensionHours finds the tail of each shift beyond afterH, clipped to window.
// Uses full (unclipped) shift duration to decide extension, then clips the extension span.
func collectShiftExtensionHours(shifts []Shift, windowStart, windowEnd time.Time, afterH float64) extHours {
	var out extHours
	if afterH <= 0 {
		return out
	}
	for _, sh := range shifts {
		if !sh.Start.Before(sh.End) {
			continue
		}
		dur := sh.End.Sub(sh.Start).Hours()
		if dur <= afterH {
			continue
		}
		extH := dur - afterH
		extStart := sh.End.Add(-time.Duration(extH * float64(time.Hour)))
		extEnd := sh.End
		if extStart.Before(windowStart) {
			extStart = windowStart
		}
		if extEnd.After(windowEnd) {
			extEnd = windowEnd
		}
		if !extStart.Before(extEnd) {
			continue
		}
		out.total += extEnd.Sub(extStart).Hours()
		out.slices = append(out.slices, extSlice{start: extStart, end: extEnd})
	}
	return out
}

// applyExtensionOTToDays attributes first capH of extension chronologically as 50%, rest 100%.
func applyExtensionOTToDays(days []DayHours, slices []extSlice, capH float64) {
	if len(days) == 0 || len(slices) == 0 {
		return
	}
	byKey := map[string]*DayHours{}
	for i := range days {
		days[i].Overtime50 = 0
		days[i].Overtime100 = 0
		byKey[Key(days[i].Date)] = &days[i]
	}
	remaining50 := capH
	if remaining50 < 0 {
		remaining50 = 0
	}
	for _, sp := range slices {
		for day := truncateDay(sp.start); day.Before(sp.end); day = day.AddDate(0, 0, 1) {
			dayEnd := day.Add(24 * time.Hour)
			cs, ce := sp.start, sp.end
			if cs.Before(day) {
				cs = day
			}
			if ce.After(dayEnd) {
				ce = dayEnd
			}
			if !cs.Before(ce) {
				continue
			}
			h := ce.Sub(cs).Hours()
			dh := byKey[Key(day)]
			if dh == nil {
				continue
			}
			if remaining50 > 0 {
				take := h
				if take > remaining50 {
					take = remaining50
				}
				dh.Overtime50 += take
				h -= take
				remaining50 -= take
			}
			if h > 0 {
				dh.Overtime100 += h
			}
		}
	}
	for i := range days {
		days[i].Overtime50 = roundHours(days[i].Overtime50)
		days[i].Overtime100 = roundHours(days[i].Overtime100)
	}
}

func addDaySlice(dh *DayHours, start, end, dayStart time.Time, callout bool, rules Rules, holidays map[string]string) {
	hours := end.Sub(start).Hours()
	dh.Total += hours
	if callout {
		dh.Callout += hours
	}
	switch dayStart.Weekday() {
	case time.Saturday:
		dh.Saturday += hours
	case time.Sunday:
		dh.Sunday += hours
	}
	if IsHoliday(dayStart, holidays) {
		dh.Holiday += hours
	}
	eve, night := splitEveningNight(start, end, dayStart, rules)
	dh.Evening += eve
	dh.Night += night
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

func splitOvertime(total, after50, after100 float64) (ot50, ot100 float64) {
	if after50 < 0 {
		after50 = 0
	}
	if after100 < after50 {
		after100 = after50
	}
	if total <= after50 {
		return 0, 0
	}
	if total <= after100 {
		return total - after50, 0
	}
	return after100 - after50, total - after100
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

func truncateDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func roundHours(h float64) float64 {
	return math.Round(h*100) / 100
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func roundDay(dh *DayHours) {
	dh.Total = roundHours(dh.Total)
	dh.Evening = roundHours(dh.Evening)
	dh.Night = roundHours(dh.Night)
	dh.Saturday = roundHours(dh.Saturday)
	dh.Sunday = roundHours(dh.Sunday)
	dh.Holiday = roundHours(dh.Holiday)
	dh.Callout = roundHours(dh.Callout)
	dh.Overtime50 = roundHours(dh.Overtime50)
	dh.Overtime100 = roundHours(dh.Overtime100)
}
