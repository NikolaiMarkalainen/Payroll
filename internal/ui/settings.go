package ui

import (
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

// settingsTab holds pay-calculation inputs the user provides.
type settingsTab struct {
	tesFamily *optionSelect
	tesLevel *optionSelect
	tesRegion *optionSelect
	levelFormItem *widget.FormItem
	profileHint *widget.Label
	expHint *widget.Label
	otHint *widget.Label
	personalHeading *widget.Label
	hourlyWage *widget.Entry
	levelPay *widget.Entry
	experienceTier *optionSelect
	experienceAllowance *widget.Entry
	personalAllowance *widget.Entry
	trainingEnabled *widget.Check
	trainingAllowance *widget.Entry
	trainingSection fyne.CanvasObject
	otherMode *widget.Select
	otherAllowance *widget.Entry
	eveningAllowance *widget.Entry
	eveningDoubleAllowance *widget.Entry
	eveningDoubleSection fyne.CanvasObject
	nightAllowance *widget.Entry
	saturdayAllowance *widget.Entry
	sundayAllowance *widget.Entry
	holidayAllowance *widget.Entry
	perehdytysAllowance *widget.Entry
	dailyRestHours *widget.Entry
	restViolationPercent *widget.Entry
	overtime50After *widget.Entry
	overtime100After *widget.Entry
	overtime50Item *widget.FormItem
	overtime100Item *widget.FormItem
	eveningStart *timePicker
	eveningEnd *timePicker
	nightStart *timePicker
	nightEnd *timePicker
	saturdayStart *timePicker
	saturdayEnd *timePicker
	shiftOTEnabled *widget.Check
	shiftOTSection fyne.CanvasObject
	weeklyOTEnabled *widget.Check
	weeklyOTThreshold *widget.Entry
	weeklyOTSection fyne.CanvasObject
	eveningExcludeSaturday *widget.Check
	nightExcludeSunday *widget.Check
	nightExcludeHoliday *widget.Check
	kaupanFlagsSection fyne.CanvasObject
	eveningDoubleMonthFrom int
	eveningDoubleMonthTo int
	eveningDoubleSundayOnly bool
	calloutFixedH           float64 // TES 31 hälytystyö: Vartio 2 h, else 0
	periodOTEnabled         *widget.Check
	periodMode              *widget.Select
	periodFirstThreshold    *widget.Select
	periodAnchor            *widget.Entry
	periodHeading           *widget.Label
	periodHint              *widget.Label
	periodOTSection         fyne.CanvasObject
	periodAdvancedSection   fyne.CanvasObject
	profileSelectorsSection fyne.CanvasObject
	colorShiftTitles        *widget.Check
	shiftColorOverrides     map[string]color.NRGBA
	shiftColorManual        map[string]struct{}
	colorRows               *fyne.Container
	colorEmptyHint          *widget.Label
	colorExtraEntry         *widget.Entry
	shiftsSource            func() []calendarShift
	window                  fyne.Window
	status                  *widget.Label
	saveBtn                 *widget.Button
	onSaved                 func()
	onPersist               func() // silent auto-save (TES change, etc.)
	suppressTESCallback     bool
}

func newSettingsTab() *settingsTab {
	s := &settingsTab{
		hourlyWage: widget.NewEntry(),
		levelPay: widget.NewEntry(),
		experienceAllowance: widget.NewEntry(),
		personalAllowance: widget.NewEntry(),
		trainingAllowance: widget.NewEntry(),
		otherAllowance: widget.NewEntry(),
		eveningAllowance: widget.NewEntry(),
		eveningDoubleAllowance: widget.NewEntry(),
		nightAllowance: widget.NewEntry(),
		saturdayAllowance: widget.NewEntry(),
		sundayAllowance: widget.NewEntry(),
		holidayAllowance: widget.NewEntry(),
		perehdytysAllowance: widget.NewEntry(),
		dailyRestHours: widget.NewEntry(),
		restViolationPercent: widget.NewEntry(),
		overtime50After: widget.NewEntry(),
		overtime100After: widget.NewEntry(),
		weeklyOTThreshold: widget.NewEntry(),
		eveningStart: newTimePicker("18:00"),
		eveningEnd: newTimePicker("22:00"),
		nightStart: newTimePicker("22:00"),
		nightEnd: newTimePicker("06:00"),
		saturdayStart: newTimePicker("00:00"),
		saturdayEnd: newTimePicker("24:00"),
		periodAnchor: widget.NewEntry(),
		status: widget.NewLabel(""),
	}
	s.trainingEnabled = widget.NewCheck("Koulutuslisä (tutkinto)", nil)
	s.otherMode = widget.NewSelect([]string{otherModeNone, otherModeHourly, otherModeFixed}, nil)
	s.shiftOTEnabled = widget.NewCheck("Pidennysylityö vuorosta (TES 31)", nil)
	s.weeklyOTEnabled = widget.NewCheck("Viikkoylitys (+50 %)", nil)
	s.eveningExcludeSaturday = widget.NewCheck("Ei iltalisää lauantaina", nil)
	s.nightExcludeSunday = widget.NewCheck("Ei yölisää sunnuntaina", nil)
	s.nightExcludeHoliday = widget.NewCheck("Ei yölisää pyhänä", nil)
	s.periodOTEnabled = widget.NewCheck("Jaksoylityö (TES 29)", nil)
	s.periodMode = widget.NewSelect([]string{
		periodMode120,
		periodMode128_112,
	}, nil)
	s.periodFirstThreshold = widget.NewSelect([]string{
		"128 h (1. jakso)",
		"112 h (1. jakso)",
	}, nil)

	s.tesFamily = newOptionSelect(nil, func(v string) {
		s.applyTESFamily(v)
	})
	s.ensureTESFamilyOptions()
	// Level/region/service options come from the chosen TES — Oma starts empty.
	s.tesLevel = newOptionSelect(nil, func(string) {
		if s.suppressTESCallback {
			return
		}
		s.applyFromSelectors()
	})
	s.tesRegion = newOptionSelect(tesRegionNames(), func(string) {
		if s.suppressTESCallback {
			return
		}
		s.applyFromSelectors()
	})
	s.experienceTier = newOptionSelect(nil, func(v string) {
		s.applyExperienceTier(v)
	})

	s.tesFamily.SetSelected(tesFamilyCustom)

	s.hourlyWage.SetPlaceHolder("0.00")
	s.levelPay.SetPlaceHolder("0.00")
	s.experienceAllowance.SetPlaceHolder("0.00")
	s.personalAllowance.SetPlaceHolder("0.00")
	s.trainingAllowance.SetPlaceHolder("0.00")
	s.otherAllowance.SetPlaceHolder("0.00")
	s.eveningAllowance.SetPlaceHolder("0.00")
	s.eveningDoubleAllowance.SetPlaceHolder("0.00")
	s.nightAllowance.SetPlaceHolder("0.00")
	s.saturdayAllowance.SetPlaceHolder("0.00")
	s.sundayAllowance.SetPlaceHolder("0.00")
	s.holidayAllowance.SetPlaceHolder("0.00")
	s.perehdytysAllowance.SetPlaceHolder("0.00")
	s.experienceAllowance.SetText("0.00")
	s.personalAllowance.SetText("0.00")
	s.trainingAllowance.SetText("0.00")
	s.trainingEnabled.SetChecked(false)
	s.otherMode.SetSelected(otherModeNone)
	s.otherAllowance.SetText("0.00")
	s.eveningDoubleAllowance.SetText("0.00")
	s.perehdytysAllowance.SetText("0.00")
	s.dailyRestHours.SetText("10")
	s.restViolationPercent.SetText("50")
	s.overtime50After.SetText("12")
	s.overtime100After.SetText("18")
	s.weeklyOTThreshold.SetText("37.5")
	s.shiftOTEnabled.SetChecked(false)
	s.weeklyOTEnabled.SetChecked(false)
	s.eveningExcludeSaturday.SetChecked(false)
	s.nightExcludeSunday.SetChecked(false)
	s.nightExcludeHoliday.SetChecked(false)
	s.periodOTEnabled.SetChecked(true)
	s.periodMode.SetSelected(periodMode120)
	s.periodFirstThreshold.SetSelected("128 h (1. jakso)")
	s.periodAnchor.SetPlaceHolder("PP.KK.VVVV")
	// Default: 12.01 of the current payroll year (J1 = 12.01 - 01.02).
	s.periodAnchor.SetText(formatFIDate(defaultPeriodYearAnchor(time.Now())))

	s.colorShiftTitles = widget.NewCheck("Väritä vuorojen otsikot koodin mukaan", nil)
	s.colorShiftTitles.SetChecked(true)

	return s
}

func (s *settingsTab) canvas() fyne.CanvasObject {
	hint := widget.NewLabel("Valitse TES - näkymä vaihtuu sopimuksen mukaan. Tai käytä Oma ja syötä arvot itse.")
	hint.Wrapping = fyne.TextWrapWord

	s.profileHint = widget.NewLabel("")
	s.profileHint.Wrapping = fyne.TextWrapWord

	s.ensureTESFamilyOptions()

	// TES family select stays outside VScroll so the dropdown popup is not clipped.
	tesForm := widget.NewForm(
		widget.NewFormItem("TES", s.tesFamily),
	)
	s.levelFormItem = widget.NewFormItem("Taso / palkkaryhmä", s.tesLevel)
	s.profileSelectorsSection = widget.NewForm(
		s.levelFormItem,
		widget.NewFormItem("Alue", s.tesRegion),
		widget.NewFormItem("Palvelusaika", s.experienceTier),
	)

	s.eveningDoubleSection = widget.NewForm(
		widget.NewFormItem("Iltalisä 2x (e/h)", s.eveningDoubleAllowance),
	)

	payForm := widget.NewForm(
		widget.NewFormItem("Tuntipalkka (e/h)", s.hourlyWage),
		widget.NewFormItem("Tasopalkka (e/kk)", s.levelPay),
		widget.NewFormItem("Iltalisä (e/h)", s.eveningAllowance),
		widget.NewFormItem("Yölisä (e/h)", s.nightAllowance),
		widget.NewFormItem("Lauantailisä (e/h)", s.saturdayAllowance),
		widget.NewFormItem("Sunnuntailisä (e/h)", s.sundayAllowance),
		widget.NewFormItem("Pyhälisä (e/h)", s.holidayAllowance),
		widget.NewFormItem("Perehdytyslisä (e/h)", s.perehdytysAllowance),
	)

	s.expHint = widget.NewLabel("")
	s.expHint.Wrapping = fyne.TextWrapWord

	s.trainingSection = widget.NewForm(
		widget.NewFormItem("", s.trainingEnabled),
		widget.NewFormItem("Koulutuslisä (e/h)", s.trainingAllowance),
	)

	personalForm := widget.NewForm(
		widget.NewFormItem("Kokemuslisä (e/h)", s.experienceAllowance),
		widget.NewFormItem("Henkilökohtainen lisä (e/h)", s.personalAllowance),
		widget.NewFormItem("Muu lisä", s.otherMode),
		widget.NewFormItem("Muu lisä määrä", s.otherAllowance),
	)

	s.otHint = widget.NewLabel("")
	s.otHint.Wrapping = fyne.TextWrapWord

	s.shiftOTSection = widget.NewForm(
		widget.NewFormItem("", s.shiftOTEnabled),
	)
	s.weeklyOTSection = widget.NewForm(
		widget.NewFormItem("", s.weeklyOTEnabled),
		widget.NewFormItem("Viikkoylitys kynnys (h)", s.weeklyOTThreshold),
	)
	s.kaupanFlagsSection = widget.NewForm(
		widget.NewFormItem("Lauantailisä alkaa", s.saturdayStart.canvas()),
		widget.NewFormItem("Lauantailisä päättyy", s.saturdayEnd.canvas()),
		widget.NewFormItem("", s.eveningExcludeSaturday),
		widget.NewFormItem("", s.nightExcludeSunday),
		widget.NewFormItem("", s.nightExcludeHoliday),
	)

	// Clean ASCII-safe labels (no § / → / replacement glyphs — Fyne fonts often show those as <>).
	s.overtime50Item = widget.NewFormItem("Vuoro yli (h), 50 % (TES 31)", s.overtime50After)
	s.overtime100Item = widget.NewFormItem("Pidennystunnit 100 % jälkeen (h)", s.overtime100After)

	rulesForm := widget.NewForm(
		widget.NewFormItem("Vuorokausilepo (h)", s.dailyRestHours),
		widget.NewFormItem("Leporikkomuskorvaus (%)", s.restViolationPercent),
		s.overtime50Item,
		s.overtime100Item,
		widget.NewFormItem("Ilta alkaa", s.eveningStart.canvas()),
		widget.NewFormItem("Ilta päättyy", s.eveningEnd.canvas()),
		widget.NewFormItem("Yö alkaa", s.nightStart.canvas()),
		widget.NewFormItem("Yö päättyy", s.nightEnd.canvas()),
	)

	s.periodHint = widget.NewLabel("")
	s.periodHint.Wrapping = fyne.TextWrapWord

	s.periodHeading = widget.NewLabel("Jaksoylityö")
	s.periodHeading.TextStyle = fyne.TextStyle{Bold: true}

	s.periodAdvancedSection = widget.NewForm(
		widget.NewFormItem("Jakson malli", s.periodMode),
		widget.NewFormItem("Ensimmäinen jakso", s.periodFirstThreshold),
		widget.NewFormItem("Vuoden J1 (12.01)", s.periodAnchor),
	)
	s.periodOTSection = container.NewVBox(
		s.periodHeading,
		s.periodHint,
		widget.NewForm(
			widget.NewFormItem("", s.periodOTEnabled),
		),
		s.periodAdvancedSection,
	)

	calendarSection := s.buildCalendarSection()

	profileHeading := widget.NewLabel("TES-pohja")
	profileHeading.TextStyle = fyne.TextStyle{Bold: true}
	payHeading := widget.NewLabel("Palkka ja lisät")
	payHeading.TextStyle = fyne.TextStyle{Bold: true}
	s.personalHeading = widget.NewLabel("Kokemus- ja henkilökohtainen lisä")
	s.personalHeading.TextStyle = fyne.TextStyle{Bold: true}
	rulesHeading := widget.NewLabel("Aikaikkunat, lepo ja ylityö")
	rulesHeading.TextStyle = fyne.TextStyle{Bold: true}

	save := widget.NewButton("Tallenna", func() {
		if s.onSaved != nil {
			s.onSaved()
			return
		}
		s.status.SetText("Tallennus ei ole kytketty.")
	})
	s.saveBtn = save

	if s.colorShiftTitles != nil {
		s.colorShiftTitles.OnChanged = func(bool) {
			if s.onSaved != nil {
				s.onSaved()
			}
		}
	}

	tesHeader := container.NewVBox(
		hint,
		widget.NewSeparator(),
		profileHeading,
		s.profileHint,
		tesForm,
	)

	scrollBody := container.NewVBox(
		s.profileSelectorsSection,
		widget.NewSeparator(),
		calendarSection,
		widget.NewSeparator(),
		payHeading,
		payForm,
		s.eveningDoubleSection,
		widget.NewSeparator(),
		s.personalHeading,
		s.expHint,
		personalForm,
		s.trainingSection,
		widget.NewSeparator(),
		rulesHeading,
		s.otHint,
		rulesForm,
		s.shiftOTSection,
		s.weeklyOTSection,
		s.kaupanFlagsSection,
		widget.NewSeparator(),
		s.periodOTSection,
		save,
		s.status,
	)

	s.syncTESVisibility(tesFamilyCustom)
	return container.NewPadded(container.NewBorder(
		tesHeader,
		nil, nil, nil,
		container.NewVScroll(scrollBody),
	))
}

func (s *settingsTab) allowanceRules() allowanceRules {
	r := defaultAllowanceRules()
	r.calloutFixedH = s.calloutFixedH
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
			// Time picker encodes midnight as 24:00; calc uses 0 as day start.
			r.nightStartMin = midnightStartMinutes(m)
		}
	}
	if s.nightEnd != nil {
		if m, err := clockToMinutes(s.nightEnd.value()); err == nil {
			r.nightEndMin = m
		}
	}
	if s.saturdayStart != nil {
		if m, err := clockToMinutes(s.saturdayStart.value()); err == nil {
			r.saturdayStartMin = midnightStartMinutes(m)
		}
	}
	if s.saturdayEnd != nil {
		if m, err := clockToMinutes(s.saturdayEnd.value()); err == nil {
			r.saturdayEndMin = m
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

// midnightStartMinutes maps 24:00 (1440) to 0 so window-of-day windows start at midnight.
func midnightStartMinutes(m int) int {
	if m >= 24*60 {
		return 0
	}
	return m
}

func (s *settingsTab) colorShiftTitlesEnabled() bool {
	return s != nil && s.colorShiftTitles != nil && s.colorShiftTitles.Checked
}
