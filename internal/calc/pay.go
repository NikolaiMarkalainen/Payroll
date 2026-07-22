package calc

import "time"

// Calculate computes TES-style pay for shifts overlapping [From, To] (calendar days).
//
// Orchestrator only: allowance classification → overtime → money.
// Components: allowance.go, overtime.go, period.go, holiday.go, round.go.
// TES profiles (Rules/Rates factories): kaupan.go, vartiointi.go — no TES-name
// switches in this file.
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
	if rules.WeeklyOTEnabled && rules.WeeklyOTThresholdH <= 0 {
		rules.WeeklyOTThresholdH = 37.5
	}
	holidays := in.Holidays
	if holidays == nil {
		holidays = HolidaySetRange(in.From, in.To)
	}

	fromDay := truncateDay(in.From)
	toDay := truncateDay(in.To)
	periodStart := fromDay
	periodEnd := toDay.Add(24 * time.Hour)

	byDay := map[string]*DayHours{}

	for _, sh := range in.Shifts {
		if !sh.Start.Before(sh.End) {
			continue
		}
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
		out.EveningDoubleHours += dh.EveningDouble
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

	if rules.WeeklyOTEnabled {
		weekly := weeklyOT50Hours(out.Days, rules.WeeklyOTThresholdH)
		out.WeeklyOT50Hours = roundHours(weekly)
		out.Overtime50Hours = roundHours(out.Overtime50Hours + weekly)
	}

	if rules.PeriodOTEnabled {
		credited := in.CreditedAbsenceH
		if credited < 0 {
			credited = 0
		}
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

	applyMoney(&out, in.Rates)
	return out
}

func applyMoney(out *Breakdown, r Rates) {
	eff := r.EffectiveHourly()
	eveDoubleRate := r.EveningDouble
	if eveDoubleRate == 0 && out.EveningDoubleHours > 0 {
		eveDoubleRate = 2 * r.Evening
	}
	normalEve := out.EveningHours - out.EveningDoubleHours
	if normalEve < 0 {
		normalEve = 0
	}
	out.BasePay = out.BaseHours * r.Hourly
	out.ExperiencePay = out.BaseHours * r.Experience
	out.PersonalPay = out.BaseHours * r.Personal
	out.TrainingPay = out.BaseHours * r.Training
	out.OtherPay = out.BaseHours*r.OtherHourly + r.OtherFixed
	out.EveningPay = normalEve*r.Evening + out.EveningDoubleHours*eveDoubleRate
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
	out.EveningDoubleHours = roundHours(out.EveningDoubleHours)
	out.NightHours = roundHours(out.NightHours)
	out.SaturdayHours = roundHours(out.SaturdayHours)
	out.SundayHours = roundHours(out.SundayHours)
	out.HolidayHours = roundHours(out.HolidayHours)
	out.CalloutHours = roundHours(out.CalloutHours)
	out.Overtime50Hours = roundHours(out.Overtime50Hours)
	out.Overtime100Hours = roundHours(out.Overtime100Hours)
	out.WeeklyOT50Hours = roundHours(out.WeeklyOT50Hours)

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
}
