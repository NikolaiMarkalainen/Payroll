package ui

import (
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// settingsTab holds pay-calculation inputs the user provides.
type settingsTab struct {
	hourlyWage           *widget.Entry
	levelPay             *widget.Entry
	eveningAllowance     *widget.Entry
	nightAllowance       *widget.Entry
	saturdayAllowance    *widget.Entry
	sundayAllowance      *widget.Entry
	holidayAllowance     *widget.Entry
	dailyRestHours       *widget.Entry
	restViolationPercent *widget.Entry
	overtime50After      *widget.Entry
	overtime100After     *widget.Entry
	eveningStart         *timePicker
	eveningEnd           *timePicker
	nightStart           *timePicker
	nightEnd             *timePicker
	status               *widget.Label
	saveBtn              *widget.Button
	onSaved              func()
}

func newSettingsTab() *settingsTab {
	s := &settingsTab{
		hourlyWage:           widget.NewEntry(),
		levelPay:             widget.NewEntry(),
		eveningAllowance:     widget.NewEntry(),
		nightAllowance:       widget.NewEntry(),
		saturdayAllowance:    widget.NewEntry(),
		sundayAllowance:      widget.NewEntry(),
		holidayAllowance:     widget.NewEntry(),
		dailyRestHours:       widget.NewEntry(),
		restViolationPercent: widget.NewEntry(),
		overtime50After:      widget.NewEntry(),
		overtime100After:     widget.NewEntry(),
		eveningStart:         newTimePicker("18:00"),
		eveningEnd:           newTimePicker("22:00"),
		nightStart:           newTimePicker("22:00"),
		nightEnd:             newTimePicker("06:00"),
		status:               widget.NewLabel(""),
	}

	s.hourlyWage.SetPlaceHolder("0.00")
	s.levelPay.SetPlaceHolder("0.00")
	s.eveningAllowance.SetPlaceHolder("0.00")
	s.nightAllowance.SetPlaceHolder("0.00")
	s.saturdayAllowance.SetPlaceHolder("0.00")
	s.sundayAllowance.SetPlaceHolder("0.00")
	s.holidayAllowance.SetPlaceHolder("0.00")
	s.dailyRestHours.SetText("10")
	s.restViolationPercent.SetText("50")
	s.overtime50After.SetText("8")
	s.overtime100After.SetText("10")

	return s
}

func (s *settingsTab) canvas() fyne.CanvasObject {
	hint := widget.NewLabel("Syötä palkkaan vaikuttavat tiedot. Näitä käytetään myöhemmin laskennassa.")
	hint.Wrapping = fyne.TextWrapWord

	payForm := widget.NewForm(
		widget.NewFormItem("Tuntipalkka (€/h)", s.hourlyWage),
		widget.NewFormItem("Tasopalkka (€/kk)", s.levelPay),
		widget.NewFormItem("Iltalisä (€/h)", s.eveningAllowance),
		widget.NewFormItem("Yölisä (€/h)", s.nightAllowance),
		widget.NewFormItem("Lauantailisä (€/h)", s.saturdayAllowance),
		widget.NewFormItem("Sunnuntailisä (€/h)", s.sundayAllowance),
		widget.NewFormItem("Pyhälisä (€/h)", s.holidayAllowance),
	)

	otHint := widget.NewLabel("TES: usein vuorokautinen ylityö 50 % 8 h jälkeen ja 100 % 10 h jälkeen (ensimmäiset 2 ylityötuntia 50 %).")
	otHint.Wrapping = fyne.TextWrapWord

	rulesForm := widget.NewForm(
		widget.NewFormItem("Vuorokausilepo (h)", s.dailyRestHours),
		widget.NewFormItem("Leporikkomuskorvaus (%)", s.restViolationPercent),
		widget.NewFormItem("Ylityö 50 % alkaa (h jälkeen)", s.overtime50After),
		widget.NewFormItem("Ylityö 100 % alkaa (h jälkeen)", s.overtime100After),
		widget.NewFormItem("Ilta alkaa", s.eveningStart.canvas()),
		widget.NewFormItem("Ilta päättyy", s.eveningEnd.canvas()),
		widget.NewFormItem("Yö alkaa", s.nightStart.canvas()),
		widget.NewFormItem("Yö päättyy", s.nightEnd.canvas()),
	)

	payHeading := widget.NewLabel("Palkka ja lisät")
	payHeading.TextStyle = fyne.TextStyle{Bold: true}
	rulesHeading := widget.NewLabel("Aikaikkunat, lepo ja ylityö")
	rulesHeading.TextStyle = fyne.TextStyle{Bold: true}

	save := widget.NewButton("Tallenna", func() {
		s.status.SetText("Asetukset tallennettu istuntoon. (" +
			"ilta " + s.eveningStart.value() + "–" + s.eveningEnd.value() +
			", yö " + s.nightStart.value() + "–" + s.nightEnd.value() +
			", ylityö 50 % @" + s.overtime50After.Text + " h" +
			", 100 % @" + s.overtime100After.Text + " h)")
		if s.onSaved != nil {
			s.onSaved()
		}
	})
	s.saveBtn = save

	body := container.NewVBox(
		hint,
		widget.NewSeparator(),
		payHeading,
		payForm,
		widget.NewSeparator(),
		rulesHeading,
		otHint,
		rulesForm,
		save,
		s.status,
	)

	return container.NewPadded(container.NewVScroll(body))
}

func (s *settingsTab) allowanceRules() allowanceRules {
	r := defaultAllowanceRules()
	if s.eveningStart != nil {
		if m, err := clockToMinutes(s.eveningStart.value()); err == nil {
			r.eveningStartMin = m
		}
	}
	if s.eveningEnd != nil {
		if m, err := clockToMinutes(s.eveningEnd.value()); err == nil {
			r.eveningEndMin = m
		}
	}
	if s.nightStart != nil {
		if m, err := clockToMinutes(s.nightStart.value()); err == nil {
			r.nightStartMin = m
		}
	}
	if s.nightEnd != nil {
		if m, err := clockToMinutes(s.nightEnd.value()); err == nil {
			r.nightEndMin = m
		}
	}
	if v, ok := parseHoursEntry(s.overtime50After); ok {
		r.overtime50AfterH = v
	}
	if v, ok := parseHoursEntry(s.overtime100After); ok {
		r.overtime100AfterH = v
	}
	year := time.Now().Year()
	return r.withYearHolidays(year).withYearHolidays(year + 1).withYearHolidays(year - 1)
}

func parseHoursEntry(e *widget.Entry) (float64, bool) {
	if e == nil {
		return 0, false
	}
	raw := strings.TrimSpace(strings.ReplaceAll(e.Text, ",", "."))
	if raw == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 {
		return 0, false
	}
	return v, true
}
