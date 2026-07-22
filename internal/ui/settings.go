package ui

import (
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
	eveningStart         *timePicker
	eveningEnd           *timePicker
	nightStart           *timePicker
	nightEnd             *timePicker
	status               *widget.Label
	saveBtn              *widget.Button
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

	rulesForm := widget.NewForm(
		widget.NewFormItem("Vuorokausilepo (h)", s.dailyRestHours),
		widget.NewFormItem("Leporikkomuskorvaus (%)", s.restViolationPercent),
		widget.NewFormItem("Ilta alkaa", s.eveningStart.canvas()),
		widget.NewFormItem("Ilta päättyy", s.eveningEnd.canvas()),
		widget.NewFormItem("Yö alkaa", s.nightStart.canvas()),
		widget.NewFormItem("Yö päättyy", s.nightEnd.canvas()),
	)

	payHeading := widget.NewLabel("Palkka ja lisät")
	payHeading.TextStyle = fyne.TextStyle{Bold: true}
	rulesHeading := widget.NewLabel("Aikaikkunat ja lepo")
	rulesHeading.TextStyle = fyne.TextStyle{Bold: true}

	save := widget.NewButton("Tallenna", func() {
		s.status.SetText("Asetukset tallennettu istuntoon. (" +
			"ilta " + s.eveningStart.value() + "–" + s.eveningEnd.value() +
			", yö " + s.nightStart.value() + "–" + s.nightEnd.value() + ")")
	})
	s.saveBtn = save

	body := container.NewVBox(
		hint,
		widget.NewSeparator(),
		payHeading,
		payForm,
		widget.NewSeparator(),
		rulesHeading,
		rulesForm,
		save,
		s.status,
	)

	return container.NewPadded(container.NewVScroll(body))
}
