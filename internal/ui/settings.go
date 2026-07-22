package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// settingsTab holds pay-calculation inputs the user provides.
type settingsTab struct {
	tesFamily            *widget.Select
	tesLevel             *widget.Select
	tesRegion            *widget.Select
	hourlyWage           *widget.Entry
	levelPay             *widget.Entry
	experienceTier       *widget.Select
	experienceAllowance  *widget.Entry
	personalAllowance    *widget.Entry
	trainingEnabled      *widget.Check
	trainingAllowance    *widget.Entry
	otherMode            *widget.Select
	otherAllowance       *widget.Entry
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
	periodOTEnabled      *widget.Check
	periodMode           *widget.Select
	periodFirstThreshold *widget.Select
	periodAnchor         *widget.Entry
	status               *widget.Label
	saveBtn              *widget.Button
	onSaved              func()
	suppressTESCallback  bool
}

func newSettingsTab() *settingsTab {
	s := &settingsTab{
		hourlyWage:           widget.NewEntry(),
		levelPay:             widget.NewEntry(),
		experienceAllowance:  widget.NewEntry(),
		personalAllowance:    widget.NewEntry(),
		trainingAllowance:    widget.NewEntry(),
		otherAllowance:       widget.NewEntry(),
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
		periodAnchor:         widget.NewEntry(),
		status:               widget.NewLabel(""),
	}
	s.trainingEnabled = widget.NewCheck("Koulutuslisä (tutkinto)", nil)
	s.otherMode = widget.NewSelect([]string{otherModeNone, otherModeHourly, otherModeFixed}, nil)
	s.periodOTEnabled = widget.NewCheck("Jaksoylityö (TES 29 §)", nil)
	s.periodMode = widget.NewSelect([]string{
		periodMode120,
		periodMode128_112,
	}, nil)
	s.periodFirstThreshold = widget.NewSelect([]string{
		"128 h (1. jakso)",
		"112 h (1. jakso)",
	}, nil)

	s.tesFamily = widget.NewSelect(tesFamilyNames(), func(v string) {
		s.applyTESFamily(v)
	})
	s.tesLevel = widget.NewSelect(tesLevelNames(), func(string) {
		if s.suppressTESCallback {
			return
		}
		s.applyVartiointiFromSelectors()
	})
	s.tesRegion = widget.NewSelect(tesRegionNames(), func(string) {
		if s.suppressTESCallback {
			return
		}
		s.applyVartiointiFromSelectors()
	})
	s.experienceTier = widget.NewSelect(tesServiceNames(), func(v string) {
		s.applyExperienceTier(v)
	})

	s.tesFamily.SetSelected(tesFamilyCustom)
	s.tesLevel.SetSelected("Taso IV")
	s.tesRegion.SetSelected(tesRegionPKS)
	s.experienceTier.SetSelected(tesServicePerus)

	s.hourlyWage.SetPlaceHolder("0.00")
	s.levelPay.SetPlaceHolder("0.00")
	s.experienceAllowance.SetPlaceHolder("0.00")
	s.personalAllowance.SetPlaceHolder("0.00")
	s.trainingAllowance.SetPlaceHolder("0.25")
	s.otherAllowance.SetPlaceHolder("0.00")
	s.eveningAllowance.SetPlaceHolder("0.00")
	s.nightAllowance.SetPlaceHolder("0.00")
	s.saturdayAllowance.SetPlaceHolder("0.00")
	s.sundayAllowance.SetPlaceHolder("0.00")
	s.holidayAllowance.SetPlaceHolder("0.00")
	s.experienceAllowance.SetText("0.00")
	s.personalAllowance.SetText("0.00")
	s.trainingAllowance.SetText(fmt.Sprintf("%.2f", vartioTrainingHourly))
	s.trainingEnabled.SetChecked(false)
	s.otherMode.SetSelected(otherModeNone)
	s.otherAllowance.SetText("0.00")
	s.dailyRestHours.SetText("10")
	s.restViolationPercent.SetText("50")
	s.overtime50After.SetText("12")
	s.overtime100After.SetText("18")
	s.periodOTEnabled.SetChecked(true)
	s.periodMode.SetSelected(periodMode128_112)
	s.periodFirstThreshold.SetSelected("128 h (1. jakso)")
	s.periodAnchor.SetPlaceHolder("PP.KK.VVVV")
	// Default: Monday of current week aligned to a sample period start.
	now := time.Now()
	anchor := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	for anchor.Weekday() != time.Monday {
		anchor = anchor.AddDate(0, 0, -1)
	}
	s.periodAnchor.SetText(formatFIDate(anchor))

	return s
}

func (s *settingsTab) canvas() fyne.CanvasObject {
	hint := widget.NewLabel("Valitse TES, taso ja alue — tai syötä arvot itse (Oma).")
	hint.Wrapping = fyne.TextWrapWord

	profileHint := widget.NewLabel("Turvallisuusala = Vartiointialan TES (PAM/PALTA). Palkat 1.8.2025. Su/pyhä = 100 % valitun tason tuntipalkasta.")
	profileHint.Wrapping = fyne.TextWrapWord

	profileForm := widget.NewForm(
		widget.NewFormItem("TES", s.tesFamily),
		widget.NewFormItem("Taso", s.tesLevel),
		widget.NewFormItem("Alue", s.tesRegion),
		widget.NewFormItem("Palvelusaika", s.experienceTier),
	)

	payForm := widget.NewForm(
		widget.NewFormItem("Tuntipalkka (e/h)", s.hourlyWage),
		widget.NewFormItem("Tasopalkka (e/kk)", s.levelPay),
		widget.NewFormItem("Iltalisä (e/h)", s.eveningAllowance),
		widget.NewFormItem("Yölisä (e/h)", s.nightAllowance),
		widget.NewFormItem("Lauantailisä (e/h)", s.saturdayAllowance),
		widget.NewFormItem("Sunnuntailisä (e/h)", s.sundayAllowance),
		widget.NewFormItem("Pyhälisä (e/h)", s.holidayAllowance),
	)

	expHint := widget.NewLabel(fmt.Sprintf(
		"Kokemuslisä = palvelusaikalisän erotus perustasoon. Henkilökohtainen lisä (TES 26 §) syötetään erikseen. Koulutuslisä (TES 25 §): %.0f e/kk tai %.2f e/h tutkinnon perusteella; ei kuulu ylityön perustuntipalkkaan. Muu lisä: tunneittain kaikilta työtunneilta tai kiinteä summa laskentajaksolle.",
		vartioTrainingMonthly, vartioTrainingHourly,
	))
	expHint.Wrapping = fyne.TextWrapWord

	personalForm := widget.NewForm(
		widget.NewFormItem("Kokemuslisä (e/h)", s.experienceAllowance),
		widget.NewFormItem("Henkilökohtainen lisä (e/h)", s.personalAllowance),
		widget.NewFormItem("", s.trainingEnabled),
		widget.NewFormItem("Koulutuslisä (e/h)", s.trainingAllowance),
		widget.NewFormItem("Muu lisä", s.otherMode),
		widget.NewFormItem("Muu lisä määrä", s.otherAllowance),
	)

	otHint := widget.NewLabel("TES 31 §: jos vuoro ylittää 12 h (pidennys), ylimenevät tunnit kuten ylityö 50 %. Jakson aikana näistä ensimmäiset 18 h @ 50 %, loput @ 100 %. Nämä tunnit eivät kuulu jaksoylityöhön. Jaksoylityö (29 §): ensimmäiset 18 ylityötuntia 50 %, loput 100 %.")
	otHint.Wrapping = fyne.TextWrapWord

	rulesForm := widget.NewForm(
		widget.NewFormItem("Vuorokausilepo (h)", s.dailyRestHours),
		widget.NewFormItem("Leporikkomuskorvaus (%)", s.restViolationPercent),
		widget.NewFormItem("Vuoro yli h: 50 % (TES 31)", s.overtime50After),
		widget.NewFormItem("Pidennystunnit 100 % jälkeen (h)", s.overtime100After),
		widget.NewFormItem("Ilta alkaa", s.eveningStart.canvas()),
		widget.NewFormItem("Ilta päättyy", s.eveningEnd.canvas()),
		widget.NewFormItem("Yö alkaa", s.nightStart.canvas()),
		widget.NewFormItem("Yö päättyy", s.nightEnd.canvas()),
	)

	periodHint := widget.NewLabel("Tasoittumisjaksossa 128 h ja 112 h vuorottelevat (keskiarvo 120 h / 3 vk). Ankkuri = ensimmäisen jakson alkupäivä. Poissaolotunnit syötetään Laskelma-välilehdellä.")
	periodHint.Wrapping = fyne.TextWrapWord

	periodForm := widget.NewForm(
		widget.NewFormItem("", s.periodOTEnabled),
		widget.NewFormItem("Jakson malli", s.periodMode),
		widget.NewFormItem("Ensimmäinen jakso", s.periodFirstThreshold),
		widget.NewFormItem("Jakson ankkuri", s.periodAnchor),
	)

	profileHeading := widget.NewLabel("TES-pohja")
	profileHeading.TextStyle = fyne.TextStyle{Bold: true}
	payHeading := widget.NewLabel("Palkka ja lisät")
	payHeading.TextStyle = fyne.TextStyle{Bold: true}
	personalHeading := widget.NewLabel("Kokemus-, koulutus- ja henkilökohtainen lisä")
	personalHeading.TextStyle = fyne.TextStyle{Bold: true}
	rulesHeading := widget.NewLabel("Aikaikkunat, lepo ja ylityö")
	rulesHeading.TextStyle = fyne.TextStyle{Bold: true}
	periodHeading := widget.NewLabel("Tasoittumisjakso (jaksoylityö)")
	periodHeading.TextStyle = fyne.TextStyle{Bold: true}

	save := widget.NewButton("Tallenna", func() {
		s.status.SetText("Asetukset tallennettu istuntoon.")
		if s.onSaved != nil {
			s.onSaved()
		}
	})
	s.saveBtn = save

	body := container.NewVBox(
		hint,
		widget.NewSeparator(),
		profileHeading,
		profileHint,
		profileForm,
		widget.NewSeparator(),
		payHeading,
		payForm,
		widget.NewSeparator(),
		personalHeading,
		expHint,
		personalForm,
		widget.NewSeparator(),
		rulesHeading,
		otHint,
		rulesForm,
		widget.NewSeparator(),
		periodHeading,
		periodHint,
		periodForm,
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
