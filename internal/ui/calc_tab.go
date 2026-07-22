package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/NikolaiMarkalainen/payroll/internal/calc"
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
	}
	c.err.Importance = widget.DangerImportance
	c.err.Hide()
	c.summary.Wrapping = fyne.TextWrapWord
	c.details.Wrapping = fyne.TextWrapWord
	c.from.SetPlaceHolder("PP.KK.VVVV")
	c.to.SetPlaceHolder("PP.KK.VVVV")
	c.absence.SetPlaceHolder("0.00")
	c.absence.SetText("0")

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, -1)
	c.from.SetText(formatFIDate(start))
	c.to.SetText(formatFIDate(end))

	c.calcBtn = widget.NewButton("Laske palkka", func() {
		c.run()
	})
	c.calcBtn.Importance = widget.HighImportance
	return c
}

func (c *calcTab) canvas() fyne.CanvasObject {
	hint := widget.NewLabel("Valitse aikaväli (mieluiten yksi 3 vk jakso) ja laske TES-pohjainen palkka. Poissaolotunnit (loma 6,7 h/pv, sairaus listan mukaan, arkipyhä, vuosivapaa…) lasketaan mukaan jaksoylityöhön.")
	hint.Wrapping = fyne.TextWrapWord

	heading := widget.NewLabel("Laskelma")
	heading.TextStyle = fyne.TextStyle{Bold: true}

	rangeHeading := widget.NewLabel("Aikaväli")
	rangeHeading.TextStyle = fyne.TextStyle{Bold: true}

	form := widget.NewForm(
		widget.NewFormItem("Alkaa", c.from),
		widget.NewFormItem("Loppuu", c.to),
		widget.NewFormItem("Poissaolotunnit (jakso)", c.absence),
	)

	fillBtn := widget.NewButton("Täytä vuorojen mukaan", func() {
		c.fillFromShifts()
	})
	periodBtn := widget.NewButton("Täytä 3 vk jakso ankkurista", func() {
		c.fillPeriodFromAnchor()
	})

	resultHeading := widget.NewLabel("Tulos")
	resultHeading.TextStyle = fyne.TextStyle{Bold: true}

	body := container.NewVBox(
		heading,
		hint,
		widget.NewSeparator(),
		rangeHeading,
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
	c.from.SetText(formatFIDate(minD))
	c.to.SetText(formatFIDate(maxD))
	c.err.Hide()
}

func (c *calcTab) fillPeriodFromAnchor() {
	if c.settings == nil {
		c.showErr("Asetuksia ei ole.")
		return
	}
	anchor, ok := c.settings.periodAnchorDate()
	if !ok {
		c.showErr("Aseta jakson ankkuri Asetuksissa (PP.KK.VVVV).")
		return
	}
	// Use "Alkaa" if set, else anchor, to pick which 3-week window.
	day := anchor
	if from, err := parseFIDate(c.from.Text); err == nil {
		day = from
	}
	pFrom, pTo := calc.PeriodWindow(anchor, day)
	c.from.SetText(formatFIDate(pFrom))
	c.to.SetText(formatFIDate(pTo))
	c.err.Hide()
}

func (c *calcTab) setDemoRange() {
	c.from.SetText("20.07.2026")
	c.to.SetText("04.08.2026")
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
		From:             from,
		To:               to,
		Shifts:           c.shifts.toCalcShifts(),
		Rates:            c.settings.rates(),
		Rules:            rules,
		CreditedAbsenceH: absence,
	}
	out := calc.Calculate(in)
	c.summary.SetText(formatCalcSummary(from, to, out, rules))
	c.details.SetText(formatCalcDetails(out, rules))
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

func formatCalcDetails(out calc.Breakdown, rules calc.Rules) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(
		"Pohja: %.2f h / %.2f e\n"+
			"Kokemuslisä: %.2f h / %.2f e\n"+
			"Henkilökohtainen: %.2f h / %.2f e\n"+
			"Koulutuslisä: %.2f h / %.2f e\n"+
			"Muu lisä: %.2f e\n"+
			"Iltalisä: %.2f h / %.2f e\n"+
			"Yölisä: %.2f h / %.2f e\n"+
			"Lauantai: %.2f h / %.2f e\n"+
			"Sunnuntai: %.2f h / %.2f e\n"+
			"Pyhä: %.2f h / %.2f e\n"+
			"Pidennysylityö 50 %% (TES 31 §): %.2f h / %.2f e\n"+
			"Pidennysylityö 100 %% (yli 18 h jaksossa): %.2f h / %.2f e\n",
		out.BaseHours, out.BasePay,
		out.BaseHours, out.ExperiencePay,
		out.BaseHours, out.PersonalPay,
		out.BaseHours, out.TrainingPay,
		out.OtherPay,
		out.EveningHours, out.EveningPay,
		out.NightHours, out.NightPay,
		out.SaturdayHours, out.SaturdayPay,
		out.SundayHours, out.SundayPay,
		out.HolidayHours, out.HolidayPay,
		out.Overtime50Hours, out.Overtime50Pay,
		out.Overtime100Hours, out.Overtime100Pay,
	))
	if rules.PeriodOTEnabled {
		b.WriteString(fmt.Sprintf(
			"Jaksoylityö 50 %% (ensimmäiset 18 h): %.2f h / %.2f e\n"+
				"Jaksoylityö 100 %%: %.2f h / %.2f e\n"+
				"Jakson tunnit: %.2f + %.2f poissaoloa = %.2f (kynnys %.0f)\n",
			out.PeriodOT50Hours, out.PeriodOT50Pay,
			out.PeriodOT100Hours, out.PeriodOT100Pay,
			out.PeriodWorkedHours, out.PeriodCreditedHours, out.PeriodTotalHours, out.PeriodThresholdH,
		))
	}
	b.WriteString(fmt.Sprintf("Hälytystunnit (seuranta): %.2f h\n", out.CalloutHours))
	if len(out.Days) > 0 {
		b.WriteString("\nPäiväkohtaisesti:\n")
		for _, d := range out.Days {
			line := fmt.Sprintf("- %s: %.2f h", formatFIDate(d.Date), d.Total)
			if d.HolidayName != "" {
				line += " (" + d.HolidayName + ")"
			}
			b.WriteString(line + "\n")
		}
	}
	return b.String()
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
		out = append(out, calc.Shift{Start: start, End: end, Callout: sh.Callout})
	}
	return out
}

func (s *settingsTab) rates() calc.Rates {
	training := 0.0
	if s.trainingEnabled != nil && s.trainingEnabled.Checked {
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
		Hourly:      entryFloatOr(s.hourlyWage, 0),
		Experience:  entryFloatOr(s.experienceAllowance, 0),
		Personal:    entryFloatOr(s.personalAllowance, 0),
		Training:    training,
		OtherHourly: otherH,
		OtherFixed:  otherF,
		Evening:     entryFloatOr(s.eveningAllowance, 0),
		Night:       entryFloatOr(s.nightAllowance, 0),
		Saturday:    entryFloatOr(s.saturdayAllowance, 0),
		Sunday:      entryFloatOr(s.sundayAllowance, 0),
		Holiday:     entryFloatOr(s.holidayAllowance, 0),
	}
}

func (s *settingsTab) calcRules() calc.Rules {
	ar := s.allowanceRules()
	shiftAfter := ar.overtime50AfterH
	if shiftAfter <= 0 {
		shiftAfter = calc.ShiftOTAfterHDefault
	}
	cap100 := ar.overtime100AfterH
	if cap100 <= 0 {
		cap100 = calc.ShiftOT50CapHDefault
	}
	r := calc.Rules{
		EveningStartMin:   ar.eveningStartMin,
		EveningEndMin:     ar.eveningEndMin,
		NightStartMin:     ar.nightStartMin,
		NightEndMin:       ar.nightEndMin,
		Overtime50AfterH:  ar.overtime50AfterH,
		Overtime100AfterH: ar.overtime100AfterH,
		ShiftOTAfterH:     shiftAfter,
		ShiftOT50CapH:     cap100,
		PeriodOTEnabled:   s.periodOTEnabled != nil && s.periodOTEnabled.Checked,
		PeriodThresholdH:  calc.PeriodThreshold120,
		PeriodOT50AfterH:  calc.DefaultPeriodOT50H,
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
