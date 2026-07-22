package ui

import "fyne.io/fyne/v2"

// TES settings UI wiring (visibility / selectors). Dispatch lives in tes_utils.go.

// ensureTESFamilyOptions keeps the TES dropdown complete (Oma, Vartiointi, Kauppa).
func (s *settingsTab) ensureTESFamilyOptions() {
	if s.tesFamily == nil {
		return
	}
	want := tesFamilyNames()
	s.tesFamily.Options = append([]string(nil), want...)
	s.tesFamily.PlaceHolder = "Valitse TES"
	s.tesFamily.Refresh()
}

func (s *settingsTab) applyTESFamily(name string) {
	if s.suppressTESCallback {
		return
	}
	s.ensureTESFamilyOptions()
	dispatchTESFamilyApply(s, name)
}

func (s *settingsTab) applyFromSelectors() {
	if s.tesFamily == nil {
		return
	}
	dispatchTESFromSelectors(s, s.tesFamily.Selected)
}

func (s *settingsTab) applyExperienceTier(tier string) {
	if s.suppressTESCallback {
		return
	}
	_ = tier
	s.applyFromSelectors()
}

func (s *settingsTab) setProfileSelectorsForFamily(family string) {
	if s.tesLevel == nil || s.experienceTier == nil {
		return
	}
	was := s.suppressTESCallback
	s.suppressTESCallback = true
	defer func() { s.suppressTESCallback = was }()

	spec := profileSelectorsForFamily(family)
	if !spec.KeepCurrentOpts {
		s.tesLevel.Options = spec.Levels
		s.experienceTier.Options = spec.Services
		if !containsString(spec.Levels, s.tesLevel.Selected) {
			s.tesLevel.SetSelected(spec.DefaultLevel)
		}
		if !containsString(spec.Services, s.experienceTier.Selected) {
			s.experienceTier.SetSelected(spec.DefaultService)
		}
	}
	s.tesLevel.Refresh()
	s.experienceTier.Refresh()
}

func (s *settingsTab) syncTESVisibility(family string) {
	f := featuresForFamily(family)
	if s.levelFormItem != nil && f.LevelLabel != "" {
		s.levelFormItem.Text = f.LevelLabel
	}
	setObjVisible(s.profileSelectorsSection, f.ProfileSelectors)
	setObjVisible(s.trainingSection, f.Training)
	setObjVisible(s.eveningDoubleSection, f.EveningDouble)
	setObjVisible(s.weeklyOTSection, f.WeeklyOT)
	setObjVisible(s.periodOTSection, f.PeriodOT)
	setObjVisible(s.periodAdvancedSection, f.PeriodOT && !f.PeriodFixed120)
	setObjVisible(s.shiftOTSection, f.ShiftOT)
	setObjVisible(s.kaupanFlagsSection, f.KaupanFlags)

	if f.PeriodFixed120 {
		if s.periodMode != nil {
			s.periodMode.SetSelected(periodMode120)
		}
		if s.periodOTEnabled != nil {
			s.periodOTEnabled.SetChecked(true)
		}
		if s.periodHeading != nil {
			s.periodHeading.SetText("Jaksoylityö")
		}
		if s.periodHint != nil {
			s.periodHint.SetText("Kiinteä 120 h / 3 vk. Poissaolotunnit syötetään Laskelma-välilehdellä.")
		}
	} else if f.PeriodOT {
		if s.periodHeading != nil {
			s.periodHeading.SetText("Tasoittumisjakso (jaksoylityö)")
		}
		if s.periodHint != nil {
			s.periodHint.SetText("Tasoittumisjaksossa 128 h ja 112 h vuorottelevat (keskiarvo 120 h / 3 vk). Ankkuri = ensimmäisen jakson alkupäivä. Poissaolotunnit syötetään Laskelma-välilehdellä.")
		}
	}

	if !f.Training {
		if s.trainingEnabled != nil {
			s.trainingEnabled.SetChecked(false)
		}
		if s.trainingAllowance != nil {
			s.trainingAllowance.SetText("0.00")
		}
	}
	if s.personalHeading != nil {
		s.personalHeading.SetText(personalHeadingText(f.Training))
	}
	if s.profileHint != nil {
		s.profileHint.SetText(profileHintText(family))
	}
	if s.expHint != nil {
		s.expHint.SetText(expHintText(family))
	}
	if s.otHint != nil {
		s.otHint.SetText(otHintText(family))
	}
}

// applyCustomBarebones resets Oma (tyhjä) to a minimal editable pay sheet.
func (s *settingsTab) applyCustomBarebones() {
	was := s.suppressTESCallback
	s.suppressTESCallback = true
	defer func() { s.suppressTESCallback = was }()

	if s.tesFamily != nil {
		s.tesFamily.SetSelected(tesFamilyCustom)
	}
	// Drop any leftover Vartio/Kauppa level label so it cannot stick on Oma.
	if s.tesLevel != nil {
		s.tesLevel.Options = nil
		s.tesLevel.SetSelected("")
	}
	if s.tesRegion != nil {
		s.tesRegion.SetSelected("")
	}
	if s.experienceTier != nil {
		s.experienceTier.Options = nil
		s.experienceTier.SetSelected("")
	}
	if s.hourlyWage != nil {
		s.hourlyWage.SetText("0.00")
	}
	if s.levelPay != nil {
		s.levelPay.SetText("0.00")
	}
	if s.eveningAllowance != nil {
		s.eveningAllowance.SetText("0.00")
	}
	if s.nightAllowance != nil {
		s.nightAllowance.SetText("0.00")
	}
	if s.saturdayAllowance != nil {
		s.saturdayAllowance.SetText("0.00")
	}
	if s.sundayAllowance != nil {
		s.sundayAllowance.SetText("0.00")
	}
	if s.holidayAllowance != nil {
		s.holidayAllowance.SetText("0.00")
	}
	if s.experienceAllowance != nil {
		s.experienceAllowance.SetText("0.00")
	}
	if s.shiftOTEnabled != nil {
		s.shiftOTEnabled.SetChecked(false)
	}
	if s.weeklyOTEnabled != nil {
		s.weeklyOTEnabled.SetChecked(false)
	}
	if s.eveningExcludeSaturday != nil {
		s.eveningExcludeSaturday.SetChecked(false)
	}
	if s.nightExcludeSunday != nil {
		s.nightExcludeSunday.SetChecked(false)
	}
	if s.nightExcludeHoliday != nil {
		s.nightExcludeHoliday.SetChecked(false)
	}
	if s.eveningDoubleAllowance != nil {
		s.eveningDoubleAllowance.SetText("0.00")
	}
	s.eveningDoubleMonthFrom = 0
	s.eveningDoubleMonthTo = 0
	s.eveningDoubleSundayOnly = false
	if s.periodOTEnabled != nil {
		s.periodOTEnabled.SetChecked(true)
	}
	if s.periodMode != nil {
		s.periodMode.SetSelected(periodMode120)
	}
	if s.trainingEnabled != nil {
		s.trainingEnabled.SetChecked(false)
	}
	if s.trainingAllowance != nil {
		s.trainingAllowance.SetText("0.00")
	}
}

func setObjVisible(o fyne.CanvasObject, on bool) {
	if o == nil {
		return
	}
	if on {
		o.Show()
	} else {
		o.Hide()
	}
}

// applyDemoTES loads Vartiointi Taso IV (PK-seutu, perus, 128/112) for tests.
// Startup / make air must not call this — demo only seeds roster + calc dates.
func (s *settingsTab) applyDemoTES() {
	s.ensureTESFamilyOptions()
	s.applyVartiointiPay("Taso IV", true, tesServicePerus)
	s.syncTESVisibility(tesFamilyVartio)
	if s.periodAnchor != nil {
		s.periodAnchor.SetText("06.07.2026")
	}
}
