package ui

import (
	"fmt"

	"payroll/internal/calc"
)

const (
	kaupanService2y = "2 vuotta"
	kaupanService4y = "4 vuotta"
	kaupanService6y = "6 vuotta"
	kaupanService9y = "9 vuotta"
)

func kaupanGroupNames() []string {
	return []string{"A", "B", "C", "D"}
}

func kaupanServiceNames() []string {
	return []string{kaupanService2y, kaupanService4y, kaupanService6y, kaupanService9y}
}

// Kaupan TES työntekijät (myyjät) taulukkopalkat 1.5.2025-31.7.2026.
// Indeksit: 0=2v, 1=4v, 2=6v, 3=9v.
var kaupanPayPKS = map[string][4]tesPayCell{
	"A": {{2011, 12.57}, {2083, 13.02}, {2195, 13.72}, {2301, 14.38}},
	"B": {{2132, 13.33}, {2212, 13.83}, {2337, 14.61}, {2441, 15.26}},
	"C": {{2277, 14.23}, {2359, 14.74}, {2517, 15.73}, {2638, 16.49}},
	"D": {{2398, 14.99}, {2488, 15.55}, {2655, 16.59}, {2857, 17.86}},
}

var kaupanPayMuu = map[string][4]tesPayCell{
	"A": {{1931, 12.07}, {1999, 12.49}, {2101, 13.13}, {2197, 13.73}},
	"B": {{2050, 12.81}, {2127, 13.29}, {2234, 13.96}, {2331, 14.57}},
	"C": {{2179, 13.62}, {2258, 14.11}, {2399, 14.99}, {2509, 15.68}},
	"D": {{2297, 14.36}, {2405, 15.03}, {2528, 15.80}, {2707, 16.92}},
}

func kaupanServiceIndex(name string) int {
	switch name {
	case kaupanService4y:
		return 1
	case kaupanService6y:
		return 2
	case kaupanService9y:
		return 3
	default:
		return 0
	}
}

func minutesToClock(m int) string {
	if m >= 24*60 {
		return "24:00"
	}
	if m < 0 {
		m = 0
	}
	return fmt.Sprintf("%02d:%02d", m/60, m%60)
}

func (s *settingsTab) applyKaupanFromSelectors() {
	if s.tesFamily == nil || s.tesFamily.Selected != tesFamilyKauppa {
		return
	}
	group := "B"
	if s.tesLevel != nil && s.tesLevel.Selected != "" {
		group = s.tesLevel.Selected
	}
	pks := true
	if s.tesRegion != nil && s.tesRegion.Selected == tesRegionMuu {
		pks = false
	}
	service := kaupanService2y
	if s.experienceTier != nil && s.experienceTier.Selected != "" {
		service = s.experienceTier.Selected
	}
	s.applyKaupanPay(group, pks, service)
}

func (s *settingsTab) applyKaupanPay(group string, pks bool, service string) {
	table := kaupanPayMuu
	regionLabel := tesRegionMuu
	if pks {
		table = kaupanPayPKS
		regionLabel = tesRegionPKS
	}
	row, ok := table[group]
	if !ok {
		s.status.SetText("Tuntematon palkkaryhmä: " + group)
		return
	}
	idx := kaupanServiceIndex(service)
	perus := row[0]
	sel := row[idx]
	exp := sel.Hourly - perus.Hourly
	if exp < 0 {
		exp = 0
	}
	rules := calc.KaupanMyyjaRules()
	eve, night, sat := calc.KaupanAllowances(pks)

	s.suppressTESCallback = true
	defer func() { s.suppressTESCallback = false }()

	s.setProfileSelectorsForFamily(tesFamilyKauppa)

	// Tuntipalkka = 2 v -portaikon perus; palvelusvuodet -> kokemuslisä e/h (ei sisään tuntipalkkaan).
	s.hourlyWage.SetText(fmt.Sprintf("%.2f", perus.Hourly))
	s.levelPay.SetText(fmt.Sprintf("%.2f", sel.Month))
	s.experienceAllowance.SetText(fmt.Sprintf("%.2f", exp))
	s.personalAllowance.SetText("0.00")
	if s.trainingEnabled != nil {
		s.trainingEnabled.SetChecked(false)
	}
	if s.trainingAllowance != nil {
		s.trainingAllowance.SetText("0.00")
	}
	s.eveningAllowance.SetText(fmt.Sprintf("%.2f", eve))
	if s.eveningDoubleAllowance != nil {
		s.eveningDoubleAllowance.SetText(fmt.Sprintf("%.2f", calc.KaupanEveningDouble(eve)))
	}
	s.nightAllowance.SetText(fmt.Sprintf("%.2f", night))
	s.saturdayAllowance.SetText(fmt.Sprintf("%.2f", sat))
	// Su/pyhä = 100 % taulukon tuntipalkasta (perus + kokemus).
	s.sundayAllowance.SetText(fmt.Sprintf("%.2f", sel.Hourly))
	s.holidayAllowance.SetText(fmt.Sprintf("%.2f", sel.Hourly))
	if s.perehdytysAllowance != nil {
		s.perehdytysAllowance.SetText("0.00")
	}
	s.dailyRestHours.SetText("11")
	s.restViolationPercent.SetText("50")
	s.overtime50After.SetText(fmt.Sprintf("%.0f", rules.Overtime50AfterH))
	s.overtime100After.SetText(fmt.Sprintf("%.0f", rules.Overtime100AfterH))
	s.eveningStart.set(minutesToClock(rules.EveningStartMin))
	s.eveningEnd.set(minutesToClock(rules.EveningEndMin))
	s.nightStart.set(minutesToClock(rules.NightStartMin))
	s.nightEnd.set(minutesToClock(rules.NightEndMin))
	if s.saturdayStart != nil {
		s.saturdayStart.set(minutesToClock(rules.SaturdayStartMin))
	}
	if s.saturdayEnd != nil {
		s.saturdayEnd.set(minutesToClock(rules.SaturdayEndMin))
	}
	if s.shiftOTEnabled != nil {
		s.shiftOTEnabled.SetChecked(rules.ShiftOTAfterH > 0)
	}
	if s.weeklyOTEnabled != nil {
		s.weeklyOTEnabled.SetChecked(rules.WeeklyOTEnabled)
	}
	if s.weeklyOTThreshold != nil {
		s.weeklyOTThreshold.SetText(fmt.Sprintf("%.1f", rules.WeeklyOTThresholdH))
	}
	if s.eveningExcludeSaturday != nil {
		s.eveningExcludeSaturday.SetChecked(rules.EveningExcludeSaturday)
	}
	if s.nightExcludeSunday != nil {
		s.nightExcludeSunday.SetChecked(rules.NightExcludeSunday)
	}
	if s.nightExcludeHoliday != nil {
		s.nightExcludeHoliday.SetChecked(rules.NightExcludeHoliday)
	}
	s.eveningDoubleMonthFrom = rules.EveningDoubleMonthFrom
	s.eveningDoubleMonthTo = rules.EveningDoubleMonthTo
	s.eveningDoubleSundayOnly = rules.EveningDoubleSundayOnly
	s.calloutFixedH = rules.CalloutFixedH
	if s.periodOTEnabled != nil {
		s.periodOTEnabled.SetChecked(rules.PeriodOTEnabled)
	}

	if s.tesFamily != nil {
		s.tesFamily.SetSelected(tesFamilyKauppa)
	}
	if s.tesLevel != nil {
		s.tesLevel.SetSelected(group)
	}
	if s.tesRegion != nil {
		s.tesRegion.SetSelected(regionLabel)
	}
	if s.experienceTier != nil {
		s.experienceTier.SetSelected(service)
	}

	s.status.SetText(fmt.Sprintf(
		"TES: Kaupan myyjät - ryhmä %s, %s, %s (1.5.2025). Ilta 18-24, yö 00-06, la 13-24; viikkoylitys %.1f h; ei jaksoylityötä.",
		group, regionLabel, service, rules.WeeklyOTThresholdH,
	))
	s.syncTESVisibility(tesFamilyKauppa)
}
