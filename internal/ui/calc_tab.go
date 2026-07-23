package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"payroll/internal/calc"
)

type calcTab struct {
	settings *settingsTab
	shifts   *shiftsTab
	from     *widget.Entry
	to       *widget.Entry
	absence  *widget.Entry
	err      *widget.Label
	summary  *widget.Label
	details  *widget.Label
	calcBtn  *widget.Button

	periodAnchor *widget.Entry
	periodSelect *widget.Select
	periodOpts   []periodOpt // parallel to periodSelect.Options
	suppressPeriod bool
	onPeriodRangeChanged func()
	onPersist            func() // disk save (range / absence / anchor)
}

type periodOpt struct {
	index int
	from  time.Time
	to    time.Time
	label string
}

func newCalcTab(settings *settingsTab, shifts *shiftsTab) *calcTab {
	c := &calcTab{
		settings: settings,
		shifts:   shifts,
		from:     widget.NewEntry(),
		to:       widget.NewEntry(),
		absence:  widget.NewEntry(),
		err:      widget.NewLabel(""),
		summary:  widget.NewLabel("Laskentaa ei ole vielä ajettu."),
		details:  widget.NewLabel(""),
		periodAnchor: widget.NewEntry(),
	}
	c.err.Importance = widget.DangerImportance
	c.err.Hide()
	c.summary.Wrapping = fyne.TextWrapWord
	c.details.Wrapping = fyne.TextWrapOff
	c.details.TextStyle = fyne.TextStyle{Monospace: true}
	c.from.SetPlaceHolder("PP.KK.VVVV")
	c.to.SetPlaceHolder("PP.KK.VVVV")
	c.absence.SetPlaceHolder("0.00")
	c.absence.SetText("0")
	c.periodAnchor.SetPlaceHolder("PP.KK.VVVV")

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, -1)
	c.from.SetText(formatFIDate(start))
	c.to.SetText(formatFIDate(end))

	if settings != nil {
		if a, ok := settings.periodAnchorDate(); ok {
			c.periodAnchor.SetText(formatFIDate(a))
		} else if settings.periodAnchor != nil {
			c.periodAnchor.SetText(settings.periodAnchor.Text)
		}
	}

	c.periodSelect = widget.NewSelect(nil, func(label string) {
		if c.suppressPeriod {
			return
		}
		c.applyPeriodSelection(label)
	})
	c.periodSelect.PlaceHolder = "Valitse 3 vk jakso"

	c.periodAnchor.OnChanged = func(string) {
		if c.suppressPeriod {
			return
		}
		c.syncAnchorToSettings()
		c.refreshPeriodOptions()
		if c.onPeriodRangeChanged != nil {
			c.onPeriodRangeChanged()
		}
	}

	c.calcBtn = widget.NewButton("Laske palkka", func() {
		c.run()
	})
	c.calcBtn.Importance = widget.HighImportance
	c.refreshPeriodOptions()
	return c
}

func (c *calcTab) canvas() fyne.CanvasObject {
	hint := widget.NewLabel("Aseta ankkuri (oletus 12.01 = vuoden J1). Valitse 3 vk jakso (J1, J2...) ja laske palkka. Poissaolotunnit (loma 6,7 h/pv, sairaus, arkipyhä, vuosivapaa...) lasketaan mukaan jaksoylityöhön.")
	hint.Wrapping = fyne.TextWrapWord

	heading := widget.NewLabel("Laskelma")
	heading.TextStyle = fyne.TextStyle{Bold: true}

	rangeHeading := widget.NewLabel("Aikaväli")
	rangeHeading.TextStyle = fyne.TextStyle{Bold: true}

	periodForm := widget.NewForm(
		widget.NewFormItem("Vuoden J1 alkaa (ankkuri)", c.periodAnchor),
		widget.NewFormItem("Jakso (3 vk)", c.periodSelect),
	)

	form := widget.NewForm(
		widget.NewFormItem("Alkaa", c.from),
		widget.NewFormItem("Loppuu", c.to),
		widget.NewFormItem("Poissaolotunnit (jakso)", c.absence),
	)

	fillBtn := widget.NewButton("Täytä vuorojen mukaan", func() {
		c.fillFromShifts()
	})
	periodBtn := widget.NewButton("Valitse jakso vuorojen mukaan", func() {
		c.selectPeriodCoveringShifts()
	})

	resultHeading := widget.NewLabel("Tulos")
	resultHeading.TextStyle = fyne.TextStyle{Bold: true}

	body := container.NewVBox(
		heading,
		hint,
		widget.NewSeparator(),
		rangeHeading,
		periodForm,
		form,
		container.NewHBox(c.calcBtn, fillBtn, periodBtn),
		c.err,
		widget.NewSeparator(),
		resultHeading,
		c.summary,
		c.details,
	)
	return container.NewPadded(container.NewVScroll(body))
}

func (c *calcTab) syncAnchorToSettings() {
	if c.settings == nil || c.settings.periodAnchor == nil {
		return
	}
	a, err := parseFIDate(c.periodAnchor.Text)
	if err != nil {
		return
	}
	c.settings.periodAnchor.SetText(formatFIDate(a))
}

func (c *calcTab) anchorDate() (time.Time, bool) {
	if a, err := parseFIDate(c.periodAnchor.Text); err == nil {
		return a, true
	}
	if c.settings != nil {
		return c.settings.periodAnchorDate()
	}
	return time.Time{}, false
}

func (c *calcTab) refreshPeriodOptions() {
	anchor, ok := c.anchorDate()
	if !ok {
		c.periodOpts = nil
		if c.periodSelect != nil {
			c.periodSelect.Options = nil
			c.periodSelect.ClearSelected()
			c.periodSelect.Refresh()
		}
		return
	}
	minIdx, maxIdx := c.periodIndexBounds(anchor)
	opts := make([]periodOpt, 0, maxIdx-minIdx+1)
	labels := make([]string, 0, maxIdx-minIdx+1)
	for i := minIdx; i <= maxIdx; i++ {
		from, to := calc.PeriodIndexWindow(anchor, i)
		jn := periodYearNumber(from)
		label := fmt.Sprintf("%s - %s  (J%d)", formatFIDate(from), formatFIDate(to), jn)
		opts = append(opts, periodOpt{index: i, from: from, to: to, label: label})
		labels = append(labels, label)
	}
	c.periodOpts = opts
	was := c.suppressPeriod
	c.suppressPeriod = true
	c.periodSelect.Options = labels
	// Keep selection if still in list; else pick period containing current from-date.
	sel := ""
	if from, err := parseFIDate(c.from.Text); err == nil {
		idx := calc.PeriodIndexContaining(anchor, from)
		for _, o := range opts {
			if o.index == idx {
				sel = o.label
				break
			}
		}
	}
	if sel != "" {
		c.periodSelect.SetSelected(sel)
	} else if len(labels) > 0 {
		c.periodSelect.SetSelected(labels[0])
		c.setRange(opts[0].from, opts[0].to)
	}
	c.periodSelect.Refresh()
	c.suppressPeriod = was
}

func (c *calcTab) periodIndexBounds(anchor time.Time) (minIdx, maxIdx int) {
	coverFrom, coverTo := time.Time{}, time.Time{}
	if c.shifts != nil {
		for _, sh := range c.shifts.shifts {
			d := sh.Date
			if coverFrom.IsZero() || d.Before(coverFrom) {
				coverFrom = d
			}
			if coverTo.IsZero() || d.After(coverTo) {
				coverTo = d
			}
			if isOvernight(sh.Start, sh.End) {
				next := d.AddDate(0, 0, 1)
				if next.After(coverTo) {
					coverTo = next
				}
			}
		}
	}
	if coverFrom.IsZero() {
		today := time.Now()
		idx := calc.PeriodIndexContaining(anchor, today)
		return idx - 2, idx + 10
	}
	minIdx = calc.PeriodIndexContaining(anchor, coverFrom)
	maxIdx = calc.PeriodIndexContaining(anchor, coverTo)
	// Pad one period on each side.
	minIdx--
	maxIdx++
	if maxIdx < minIdx {
		maxIdx = minIdx
	}
	return minIdx, maxIdx
}

func (c *calcTab) applyPeriodSelection(label string) {
	for _, o := range c.periodOpts {
		if o.label == label {
			c.setRange(o.from, o.to)
			c.err.Hide()
			return
		}
	}
}

func (c *calcTab) selectPeriodCoveringShifts() {
	anchor, ok := c.anchorDate()
	if !ok {
		c.showErr("Aseta 1. jakson alkupäivä (ankkuri), esim. 29.06.2026.")
		return
	}
	c.syncAnchorToSettings()
	c.refreshPeriodOptions()
	if c.shifts == nil || len(c.shifts.shifts) == 0 {
		c.showErr("Ei vuoroja kalenterissa.")
		return
	}
	day := c.shifts.shifts[0].Date
	for _, sh := range c.shifts.shifts {
		if sh.Date.Before(day) {
			day = sh.Date
		}
	}
	idx := calc.PeriodIndexContaining(anchor, day)
	for _, o := range c.periodOpts {
		if o.index == idx {
			c.suppressPeriod = true
			c.periodSelect.SetSelected(o.label)
			c.suppressPeriod = false
			c.setRange(o.from, o.to)
			c.err.Hide()
			return
		}
	}
	from, to := calc.PeriodIndexWindow(anchor, idx)
	c.setRange(from, to)
	c.err.Hide()
}

func (c *calcTab) fillFromShifts() {
	if c.shifts == nil || len(c.shifts.shifts) == 0 {
		c.showErr("Ei vuoroja kalenterissa.")
		return
	}
	minD, maxD := c.shifts.shifts[0].Date, c.shifts.shifts[0].Date
	for _, sh := range c.shifts.shifts {
		if sh.Date.Before(minD) {
			minD = sh.Date
		}
		if sh.Date.After(maxD) {
			maxD = sh.Date
		}
		// Overnight continuation may end next calendar day.
		if isOvernight(sh.Start, sh.End) {
			next := sh.Date.AddDate(0, 0, 1)
			if next.After(maxD) {
				maxD = next
			}
		}
	}
	c.setRange(minD, maxD)
	c.refreshPeriodOptions()
	c.err.Hide()
}

func (c *calcTab) fillPeriodFromAnchor() {
	c.selectPeriodCoveringShifts()
}

func (c *calcTab) setDemoRange() {
	c.setRange(
		time.Date(2026, 7, 20, 0, 0, 0, 0, time.Local),
		time.Date(2026, 8, 4, 0, 0, 0, 0, time.Local),
	)
}

func (c *calcTab) setRange(from, to time.Time) {
	c.from.SetText(formatFIDate(from))
	c.to.SetText(formatFIDate(to))
	c.syncPeriodSelectToRange(from)
	if c.onPeriodRangeChanged != nil {
		c.onPeriodRangeChanged()
	}
	if c.onPersist != nil {
		c.onPersist()
	}
}

func (c *calcTab) syncPeriodSelectToRange(day time.Time) {
	if c.periodSelect == nil || len(c.periodOpts) == 0 {
		return
	}
	anchor, ok := c.anchorDate()
	if !ok {
		return
	}
	idx := calc.PeriodIndexContaining(anchor, day)
	for _, o := range c.periodOpts {
		if o.index == idx {
			was := c.suppressPeriod
			c.suppressPeriod = true
			c.periodSelect.SetSelected(o.label)
			c.suppressPeriod = was
			return
		}
	}
}

func (c *calcTab) showErr(msg string) {
	c.err.SetText(msg)
	c.err.Show()
}

func (c *calcTab) run() {
	c.err.Hide()
	from, err1 := parseFIDate(c.from.Text)
	to, err2 := parseFIDate(c.to.Text)
	if err1 != nil || err2 != nil {
		c.showErr("Virheellinen päivämäärä. Käytä muotoa PP.KK.VVVV.")
		return
	}
	if to.Before(from) {
		c.showErr("Loppupäivä ei voi olla ennen alkupäivää.")
		return
	}
	if c.settings == nil || c.shifts == nil {
		c.showErr("Laskentaa ei voi ajaa.")
		return
	}

	rules := c.settings.calcRules()
	if rules.PeriodOTEnabled {
		rules.PeriodThresholdH = c.settings.periodThresholdForRange(from)
	}
	absence := entryFloatOr(c.absence, 0)

	in := calc.PeriodInput{
		From: from,
		To: to,
		Shifts: c.shifts.toCalcShifts(),
		Rates: c.settings.rates(),
		Rules: rules,
		CreditedAbsenceH: absence,
	}
	out := calc.Calculate(in)
	c.summary.SetText(formatCalcSummary(from, to, out, rules))
	c.details.SetText(formatCalcDetails(out, rules, in.Rates))
}

func formatCalcSummary(from, to time.Time, out calc.Breakdown, rules calc.Rules) string {
	s := fmt.Sprintf(
		"Aikaväli %s-%s\n"+
			"Tunnit yhteensä: %.2f h\n"+
			"Palkka yhteensä: %.2f e",
		formatFIDate(from), formatFIDate(to),
		out.BaseHours,
		out.TotalPay,
	)
	if rules.PeriodOTEnabled {
		s += fmt.Sprintf(
			"\nJakso: %.2f h työtä + %.2f h poissaoloa = %.2f h (kynnys %.0f h) / jaksoylityö %.2f h",
			out.PeriodWorkedHours, out.PeriodCreditedHours, out.PeriodTotalHours,
			out.PeriodThresholdH, out.PeriodOTHours,
		)
	}
	return s
}

func formatCalcDetails(out calc.Breakdown, rules calc.Rules, rates calc.Rates) string {
	eff := rates.EffectiveHourly()
	eveDoubleRate := rates.EveningDouble
	if eveDoubleRate == 0 && out.EveningDoubleHours > 0 {
		eveDoubleRate = 2 * rates.Evening
	}
	normalEveH := out.EveningHours - out.EveningDoubleHours
	if normalEveH < 0 {
		normalEveH = 0
	}
	normalEvePay := normalEveH * rates.Evening
	eveDoublePay := out.EveningDoubleHours * eveDoubleRate

	ot50Label := "Ylityö 50 %"
	ot100Label := "Ylityö 100 %"
	if rules.ShiftOTAfterH > 0 {
		ot50Label = "Pidennysylityö 50 % (TES 31)"
		ot100Label = "Pidennysylityö 100 % (yli 18 h)"
	}

	type row struct {
		name  string
		hours float64
		rate  float64
		sum   float64
		hasH  bool // show hours+rate columns
	}
	rows := []row{
		{"Pohja", out.BaseHours, rates.Hourly, out.BasePay, true},
		{"Kokemuslisä", out.BaseHours, rates.Experience, out.ExperiencePay, true},
		{"Henkilökohtainen", out.BaseHours, rates.Personal, out.PersonalPay, true},
		{"Koulutuslisä", out.BaseHours, rates.Training, out.TrainingPay, true},
	}
	if rates.OtherHourly > 0 || out.OtherPay != 0 {
		if rates.OtherHourly > 0 {
			rows = append(rows, row{"Muu lisä (e/h)", out.BaseHours, rates.OtherHourly, out.BaseHours * rates.OtherHourly, true})
		}
		if rates.OtherFixed > 0 {
			rows = append(rows, row{"Muu lisä (kiinteä)", 0, 0, rates.OtherFixed, false})
		}
	} else {
		rows = append(rows, row{"Muu lisä", 0, 0, out.OtherPay, false})
	}
	rows = append(rows, row{"Iltatyölisä", normalEveH, rates.Evening, normalEvePay, true})
	// Iltalisä 2x is Kaupan (marras–joulu su), not used in Vartiointi.
	if eveningDoubleEnabled(rules) {
		rows = append(rows, row{"Iltatyölisä 2x", out.EveningDoubleHours, eveDoubleRate, eveDoublePay, true})
	}
	rows = append(rows,
		row{"Yötyölisä", out.NightHours, rates.Night, out.NightPay, true},
		row{"Lauantailisä", out.SaturdayHours, rates.Saturday, out.SaturdayPay, true},
		row{"Sunnuntaikorotus 100%", out.SundayHours, rates.Sunday, out.SundayPay, true},
		row{"Pyhäkorotus", out.HolidayHours, rates.Holiday, out.HolidayPay, true},
		row{"Perehdytyslisä", out.PerehdytysHours, rates.Perehdytys, out.PerehdytysPay, true},
		row{ot50Label, out.Overtime50Hours, eff * 0.5, out.Overtime50Pay, true},
		row{ot100Label, out.Overtime100Hours, eff, out.Overtime100Pay, true},
	)
	if rules.WeeklyOTEnabled && out.WeeklyOT50Hours > 0 {
		rows = append(rows, row{"Viikkoylitys 50 % (sis. ylityöhön)", out.WeeklyOT50Hours, eff * 0.5, out.WeeklyOT50Hours * eff * 0.5, true})
	}
	if rules.PeriodOTEnabled {
		rows = append(rows,
			row{"Jaksoylityö 50 %", out.PeriodOT50Hours, eff * 0.5, out.PeriodOT50Pay, true},
			row{"Jaksoylityö 100 %", out.PeriodOT100Hours, eff, out.PeriodOT100Pay, true},
		)
	}
	rows = append(rows, row{"Hälytystyö (kiinteä)", out.CalloutHours, eff, out.CalloutPay, true})

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%-32s %8s %8s %10s\n", "Nimike", "Tunnit", "e/h", "Summa e"))
	b.WriteString(strings.Repeat("-", 62) + "\n")
	for _, r := range rows {
		if r.hasH {
			b.WriteString(fmt.Sprintf("%-32s %8.2f %8.2f %10.2f\n", truncatePad(r.name, 32), r.hours, r.rate, r.sum))
		} else {
			b.WriteString(fmt.Sprintf("%-32s %8s %8s %10.2f\n", truncatePad(r.name, 32), "-", "-", r.sum))
		}
	}
	b.WriteString(strings.Repeat("-", 62) + "\n")
	b.WriteString(fmt.Sprintf("%-32s %8s %8s %10.2f\n", "Yhteensä", "", "", out.TotalPay))

	if rules.PeriodOTEnabled {
		b.WriteString(fmt.Sprintf(
			"\nJakson tunnit: %.2f h työtä + %.2f h poissaoloa = %.2f h (kynnys %.0f h)\n",
			out.PeriodWorkedHours, out.PeriodCreditedHours, out.PeriodTotalHours, out.PeriodThresholdH,
		))
	}
	if len(out.Days) > 0 {
		b.WriteString("\nPäiväkohtaisesti:\n")
		b.WriteString(fmt.Sprintf("%-12s %8s\n", "Päivä", "Tunnit"))
		b.WriteString(strings.Repeat("-", 22) + "\n")
		for _, d := range out.Days {
			name := formatFIDate(d.Date)
			if d.HolidayName != "" {
				name += " " + d.HolidayName
			}
			b.WriteString(fmt.Sprintf("%-12s %8.2f\n", truncatePad(name, 12), d.Total))
		}
	}
	return b.String()
}

func eveningDoubleEnabled(rules calc.Rules) bool {
	return rules.EveningDoubleMonthFrom != 0 || rules.EveningDoubleMonthTo != 0
}

func truncatePad(s string, width int) string {
	runes := []rune(s)
	if len(runes) > width {
		return string(runes[:width-1]) + "."
	}
	return s
}

func formatFIDate(t time.Time) string {
	return t.Format("02.01.2006")
}

func parseFIDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	return time.ParseInLocation("02.01.2006", s, time.Local)
}

func (s *shiftsTab) toCalcShifts() []calc.Shift {
	out := make([]calc.Shift, 0, len(s.shifts))
	for _, sh := range s.shifts {
		start, end, err := sh.absoluteRange()
		if err != nil {
			continue
		}
		out = append(out, calc.Shift{
			Start:           start,
			End:             end,
			Callout:         sh.Callout,
			PerehdytysHours: sh.perehdytysHours(),
		})
	}
	return out
}

func (s *settingsTab) rates() calc.Rates {
	training := 0.0
	if s.trainingSection != nil && !s.trainingSection.Visible() {
		training = 0
	} else if s.trainingEnabled != nil && s.trainingEnabled.Checked {
		training = entryFloatOr(s.trainingAllowance, 0)
	}
	var otherH, otherF float64
	if s.otherMode != nil {
		amt := entryFloatOr(s.otherAllowance, 0)
		switch s.otherMode.Selected {
		case otherModeHourly:
			otherH = amt
		case otherModeFixed:
			otherF = amt
		}
	}
	return calc.Rates{
		Hourly: entryFloatOr(s.hourlyWage, 0),
		Experience: entryFloatOr(s.experienceAllowance, 0),
		Personal: entryFloatOr(s.personalAllowance, 0),
		Training: training,
		OtherHourly: otherH,
		OtherFixed: otherF,
		Evening: entryFloatOr(s.eveningAllowance, 0),
		EveningDouble: entryFloatOr(s.eveningDoubleAllowance, 0),
		Night: entryFloatOr(s.nightAllowance, 0),
		Saturday: entryFloatOr(s.saturdayAllowance, 0),
		Sunday: entryFloatOr(s.sundayAllowance, 0),
		Holiday: entryFloatOr(s.holidayAllowance, 0),
		Perehdytys: entryFloatOr(s.perehdytysAllowance, 0),
	}
}

func (s *settingsTab) calcRules() calc.Rules {
	ar := s.allowanceRules()
	r := calc.Rules{
		EveningStartMin: ar.eveningStartMin,
		EveningEndMin: ar.eveningEndMin,
		NightStartMin: ar.nightStartMin,
		NightEndMin: ar.nightEndMin,
		SaturdayStartMin: ar.saturdayStartMin,
		SaturdayEndMin: ar.saturdayEndMin,
		EveningExcludeSaturday: s.eveningExcludeSaturday != nil && s.eveningExcludeSaturday.Checked,
		NightExcludeSunday: s.nightExcludeSunday != nil && s.nightExcludeSunday.Checked,
		NightExcludeHoliday: s.nightExcludeHoliday != nil && s.nightExcludeHoliday.Checked,
		EveningDoubleMonthFrom: s.eveningDoubleMonthFrom,
		EveningDoubleMonthTo: s.eveningDoubleMonthTo,
		EveningDoubleSundayOnly: s.eveningDoubleSundayOnly,
		Overtime50AfterH: ar.overtime50AfterH,
		Overtime100AfterH: ar.overtime100AfterH,
		PeriodOTEnabled: s.periodOTEnabled != nil && s.periodOTEnabled.Checked,
		PeriodThresholdH: calc.PeriodThreshold120,
		PeriodOT50AfterH: calc.DefaultPeriodOT50H,
		WeeklyOTEnabled: s.weeklyOTEnabled != nil && s.weeklyOTEnabled.Checked,
		WeeklyOTThresholdH: entryFloatOr(s.weeklyOTThreshold, 37.5),
		CalloutFixedH: s.calloutFixedH,
	}
	useShiftOT := s.shiftOTEnabled == nil || s.shiftOTEnabled.Checked
	if useShiftOT {
		shiftAfter := ar.overtime50AfterH
		if shiftAfter <= 0 {
			shiftAfter = calc.ShiftOTAfterHDefault
		}
		cap100 := ar.overtime100AfterH
		if cap100 <= 0 {
			cap100 = calc.ShiftOT50CapHDefault
		}
		r.ShiftOTAfterH = shiftAfter
		r.ShiftOT50CapH = cap100
	} else {
		r.ShiftOTAfterH = 0
		r.ShiftOT50CapH = 0
	}
	return r
}

// periodThresholdForRange picks 120 / 128 / 112 from settings for the given from date.
func (s *settingsTab) periodThresholdForRange(from time.Time) float64 {
	if s.periodMode == nil || s.periodMode.Selected == periodMode120 {
		return calc.PeriodThreshold120
	}
	first := calc.PeriodThreshold128
	if s.periodFirstThreshold != nil && strings.HasPrefix(s.periodFirstThreshold.Selected, "112") {
		first = calc.PeriodThreshold112
	}
	anchor := from
	if s.periodAnchor != nil {
		if a, err := parseFIDate(s.periodAnchor.Text); err == nil {
			anchor = a
		}
	}
	return calc.PeriodThresholdAt(anchor, from, first)
}

func (s *settingsTab) periodAnchorDate() (time.Time, bool) {
	if s.periodAnchor == nil {
		return time.Time{}, false
	}
	a, err := parseFIDate(s.periodAnchor.Text)
	if err != nil {
		return time.Time{}, false
	}
	return a, true
}

func entryFloatOr(e *widget.Entry, fallback float64) float64 {
	if v, ok := parseHoursEntry(e); ok {
		return v
	}
	return fallback
}
