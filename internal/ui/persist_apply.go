package ui

import (
	"image/color"
	"time"
)

func (s *settingsTab) exportPersisted() persistedSettings {
	out := persistedSettings{
		EveDoubleFrom:    s.eveningDoubleMonthFrom,
		EveDoubleTo:      s.eveningDoubleMonthTo,
		EveDoubleSunOnly: s.eveningDoubleSundayOnly,
		CalloutFixedH:    s.calloutFixedH,
	}
	if s.tesFamily != nil {
		out.TESFamily = s.tesFamily.Selected
	}
	if s.tesLevel != nil {
		out.TESLevel = s.tesLevel.Selected
	}
	if s.tesRegion != nil {
		out.TESRegion = s.tesRegion.Selected
	}
	if s.experienceTier != nil {
		out.ExperienceTier = s.experienceTier.Selected
	}
	if s.hourlyWage != nil {
		out.HourlyWage = s.hourlyWage.Text
	}
	if s.levelPay != nil {
		out.LevelPay = s.levelPay.Text
	}
	if s.experienceAllowance != nil {
		out.Experience = s.experienceAllowance.Text
	}
	if s.personalAllowance != nil {
		out.Personal = s.personalAllowance.Text
	}
	if s.trainingEnabled != nil {
		out.TrainingEnabled = s.trainingEnabled.Checked
	}
	if s.trainingAllowance != nil {
		out.Training = s.trainingAllowance.Text
	}
	if s.otherMode != nil {
		out.OtherMode = s.otherMode.Selected
	}
	if s.otherAllowance != nil {
		out.OtherAllowance = s.otherAllowance.Text
	}
	if s.eveningAllowance != nil {
		out.Evening = s.eveningAllowance.Text
	}
	if s.eveningDoubleAllowance != nil {
		out.EveningDouble = s.eveningDoubleAllowance.Text
	}
	if s.nightAllowance != nil {
		out.Night = s.nightAllowance.Text
	}
	if s.saturdayAllowance != nil {
		out.Saturday = s.saturdayAllowance.Text
	}
	if s.sundayAllowance != nil {
		out.Sunday = s.sundayAllowance.Text
	}
	if s.holidayAllowance != nil {
		out.Holiday = s.holidayAllowance.Text
	}
	if s.perehdytysAllowance != nil {
		out.Perehdytys = s.perehdytysAllowance.Text
	}
	if s.dailyRestHours != nil {
		out.DailyRest = s.dailyRestHours.Text
	}
	if s.restViolationPercent != nil {
		out.RestViolation = s.restViolationPercent.Text
	}
	if s.overtime50After != nil {
		out.Overtime50After = s.overtime50After.Text
	}
	if s.overtime100After != nil {
		out.Overtime100After = s.overtime100After.Text
	}
	if s.weeklyOTThreshold != nil {
		out.WeeklyOTThreshold = s.weeklyOTThreshold.Text
	}
	if s.eveningStart != nil {
		out.EveningStart = s.eveningStart.value()
	}
	if s.eveningEnd != nil {
		out.EveningEnd = s.eveningEnd.value()
	}
	if s.nightStart != nil {
		out.NightStart = s.nightStart.value()
	}
	if s.nightEnd != nil {
		out.NightEnd = s.nightEnd.value()
	}
	if s.saturdayStart != nil {
		out.SaturdayStart = s.saturdayStart.value()
	}
	if s.saturdayEnd != nil {
		out.SaturdayEnd = s.saturdayEnd.value()
	}
	if s.shiftOTEnabled != nil {
		out.ShiftOTEnabled = s.shiftOTEnabled.Checked
	}
	if s.weeklyOTEnabled != nil {
		out.WeeklyOTEnabled = s.weeklyOTEnabled.Checked
	}
	if s.eveningExcludeSaturday != nil {
		out.EveExcludeSat = s.eveningExcludeSaturday.Checked
	}
	if s.nightExcludeSunday != nil {
		out.NightExcludeSun = s.nightExcludeSunday.Checked
	}
	if s.nightExcludeHoliday != nil {
		out.NightExcludeHol = s.nightExcludeHoliday.Checked
	}
	if s.periodOTEnabled != nil {
		out.PeriodOTEnabled = s.periodOTEnabled.Checked
	}
	if s.periodMode != nil {
		out.PeriodMode = s.periodMode.Selected
	}
	if s.periodFirstThreshold != nil {
		out.PeriodFirstThr = s.periodFirstThreshold.Selected
	}
	if s.periodAnchor != nil {
		out.PeriodAnchor = s.periodAnchor.Text
	}
	if s.colorShiftTitles != nil {
		out.ColorTitles = s.colorShiftTitles.Checked
	}
	s.initShiftColors()
	if len(s.shiftColorOverrides) > 0 {
		out.ColorOverrides = make(map[string]string, len(s.shiftColorOverrides))
		for k, c := range s.shiftColorOverrides {
			out.ColorOverrides[k] = colorToHex(c)
		}
	}
	if len(s.shiftColorManual) > 0 {
		out.ColorManual = make([]string, 0, len(s.shiftColorManual))
		for k := range s.shiftColorManual {
			out.ColorManual = append(out.ColorManual, k)
		}
	}
	return out
}

func (s *settingsTab) applyPersisted(in persistedSettings) {
	s.suppressTESCallback = true
	defer func() { s.suppressTESCallback = false }()

	// Restore family + selector options first (no default overwrite of saved level).
	if s.tesFamily != nil && in.TESFamily != "" {
		s.ensureTESFamilyOptions()
		s.tesFamily.forceSelected(in.TESFamily)
		s.setProfileSelectorsForFamilyOpts(in.TESFamily, false)
	}
	if s.tesRegion != nil {
		s.tesRegion.Options = tesRegionNames()
		if in.TESRegion != "" {
			s.tesRegion.forceSelected(in.TESRegion)
		}
	}
	if s.tesLevel != nil && in.TESLevel != "" {
		if !containsString(s.tesLevel.Options, in.TESLevel) {
			s.tesLevel.Options = append(s.tesLevel.Options, in.TESLevel)
		}
		s.tesLevel.forceSelected(in.TESLevel)
	}
	if s.experienceTier != nil && in.ExperienceTier != "" {
		if !containsString(s.experienceTier.Options, in.ExperienceTier) {
			s.experienceTier.Options = append(s.experienceTier.Options, in.ExperienceTier)
		}
		s.experienceTier.forceSelected(in.ExperienceTier)
	}
	if in.TESFamily != "" {
		s.syncTESVisibility(in.TESFamily)
	}

	setEntry := func(e interface{ SetText(string) }, v string) {
		if e != nil && v != "" {
			e.SetText(v)
		}
	}
	setEntry(s.hourlyWage, in.HourlyWage)
	setEntry(s.levelPay, in.LevelPay)
	setEntry(s.experienceAllowance, in.Experience)
	setEntry(s.personalAllowance, in.Personal)
	setEntry(s.trainingAllowance, in.Training)
	setEntry(s.otherAllowance, in.OtherAllowance)
	setEntry(s.eveningAllowance, in.Evening)
	setEntry(s.eveningDoubleAllowance, in.EveningDouble)
	setEntry(s.nightAllowance, in.Night)
	setEntry(s.saturdayAllowance, in.Saturday)
	setEntry(s.sundayAllowance, in.Sunday)
	setEntry(s.holidayAllowance, in.Holiday)
	setEntry(s.perehdytysAllowance, in.Perehdytys)
	setEntry(s.dailyRestHours, in.DailyRest)
	setEntry(s.restViolationPercent, in.RestViolation)
	setEntry(s.overtime50After, in.Overtime50After)
	setEntry(s.overtime100After, in.Overtime100After)
	setEntry(s.weeklyOTThreshold, in.WeeklyOTThreshold)
	setEntry(s.periodAnchor, in.PeriodAnchor)

	if s.trainingEnabled != nil {
		s.trainingEnabled.SetChecked(in.TrainingEnabled)
	}
	// Re-apply training after syncTESVisibility (Oma path clears it).
	if in.TESFamily == tesFamilyVartio || in.TrainingEnabled || in.Training != "" {
		if s.trainingEnabled != nil {
			s.trainingEnabled.SetChecked(in.TrainingEnabled)
		}
		setEntry(s.trainingAllowance, in.Training)
	}
	if s.otherMode != nil && in.OtherMode != "" {
		s.otherMode.SetSelected(in.OtherMode)
	}
	if s.eveningStart != nil && in.EveningStart != "" {
		s.eveningStart.set(in.EveningStart)
	}
	if s.eveningEnd != nil && in.EveningEnd != "" {
		s.eveningEnd.set(in.EveningEnd)
	}
	if s.nightStart != nil && in.NightStart != "" {
		s.nightStart.set(in.NightStart)
	}
	if s.nightEnd != nil && in.NightEnd != "" {
		s.nightEnd.set(in.NightEnd)
	}
	if s.saturdayStart != nil && in.SaturdayStart != "" {
		s.saturdayStart.set(in.SaturdayStart)
	}
	if s.saturdayEnd != nil && in.SaturdayEnd != "" {
		s.saturdayEnd.set(in.SaturdayEnd)
	}
	if s.shiftOTEnabled != nil {
		s.shiftOTEnabled.SetChecked(in.ShiftOTEnabled)
	}
	if s.weeklyOTEnabled != nil {
		s.weeklyOTEnabled.SetChecked(in.WeeklyOTEnabled)
	}
	if s.eveningExcludeSaturday != nil {
		s.eveningExcludeSaturday.SetChecked(in.EveExcludeSat)
	}
	if s.nightExcludeSunday != nil {
		s.nightExcludeSunday.SetChecked(in.NightExcludeSun)
	}
	if s.nightExcludeHoliday != nil {
		s.nightExcludeHoliday.SetChecked(in.NightExcludeHol)
	}
	if s.periodOTEnabled != nil {
		s.periodOTEnabled.SetChecked(in.PeriodOTEnabled)
	}
	if s.periodMode != nil && in.PeriodMode != "" {
		s.periodMode.SetSelected(in.PeriodMode)
	}
	if s.periodFirstThreshold != nil && in.PeriodFirstThr != "" {
		s.periodFirstThreshold.SetSelected(in.PeriodFirstThr)
	}

	s.eveningDoubleMonthFrom = in.EveDoubleFrom
	s.eveningDoubleMonthTo = in.EveDoubleTo
	s.eveningDoubleSundayOnly = in.EveDoubleSunOnly
	s.calloutFixedH = in.CalloutFixedH

	if s.colorShiftTitles != nil {
		s.colorShiftTitles.SetChecked(in.ColorTitles)
	}
	s.initShiftColors()
	s.shiftColorOverrides = make(map[string]color.NRGBA)
	for k, hex := range in.ColorOverrides {
		if c, ok := hexToColor(hex); ok {
			s.shiftColorOverrides[k] = c
		}
	}
	s.shiftColorManual = make(map[string]struct{})
	for _, k := range in.ColorManual {
		s.shiftColorManual[k] = struct{}{}
	}
	s.refreshColorRows()
}

func (s *shiftsTab) exportPersisted() persistedShifts {
	out := persistedShifts{
		Month:  s.month.Format("2006-01-02"),
		NextID: s.nextID,
		Items:  make([]persistedShift, 0, len(s.shifts)),
	}
	for _, sh := range s.shifts {
		out.Items = append(out.Items, persistedShift{
			ID:              sh.ID,
			Date:            sh.Date.Format("2006-01-02"),
			Start:           sh.Start,
			End:             sh.End,
			Callout:         sh.Callout,
			Code:            sh.Code,
			PerehdytysStart: sh.PerehdytysStart,
			PerehdytysEnd:   sh.PerehdytysEnd,
		})
	}
	return out
}

func (s *shiftsTab) applyPersisted(in persistedShifts) {
	loc := time.Local
	month := s.month
	if in.Month != "" {
		if t, err := time.ParseInLocation("2006-01-02", in.Month, loc); err == nil {
			month = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
		}
	}
	items := make([]calendarShift, 0, len(in.Items))
	maxID := 0
	for _, it := range in.Items {
		d, err := time.ParseInLocation("2006-01-02", it.Date, loc)
		if err != nil {
			continue
		}
		sh := calendarShift{
			ID:              it.ID,
			Date:            time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc),
			Start:           it.Start,
			End:             it.End,
			Callout:         it.Callout,
			Code:            it.Code,
			PerehdytysStart: it.PerehdytysStart,
			PerehdytysEnd:   it.PerehdytysEnd,
		}
		if sh.ID > maxID {
			maxID = sh.ID
		}
		items = append(items, sh)
	}
	s.shifts = items
	s.month = month
	if in.NextID > maxID {
		s.nextID = in.NextID
	} else {
		s.nextID = maxID + 1
	}
	if s.nextID < 1 {
		s.nextID = 1
	}
	s.refresh()
}

func (c *calcTab) exportPersisted() persistedCalc {
	out := persistedCalc{}
	if c.from != nil {
		out.From = c.from.Text
	}
	if c.to != nil {
		out.To = c.to.Text
	}
	if c.absence != nil {
		out.Absence = c.absence.Text
	}
	if c.periodAnchor != nil {
		out.Anchor = c.periodAnchor.Text
	}
	return out
}

func (c *calcTab) applyPersisted(in persistedCalc) {
	c.suppressPeriod = true
	defer func() { c.suppressPeriod = false }()
	if c.from != nil && in.From != "" {
		c.from.SetText(in.From)
	}
	if c.to != nil && in.To != "" {
		c.to.SetText(in.To)
	}
	if c.absence != nil && in.Absence != "" {
		c.absence.SetText(in.Absence)
	}
	if c.periodAnchor != nil && in.Anchor != "" {
		c.periodAnchor.SetText(in.Anchor)
	}
	c.refreshPeriodOptions()
	if from, err := parseFIDate(c.from.Text); err == nil {
		c.syncPeriodSelectToRange(from)
	}
}
