package ui

import (
	"fmt"

	"payroll/internal/calc"
)

const (
	tesServicePerus = "Perus"
	tesService2y = "2 vuotta"
	tesService7y = "7 vuotta"
)

func vartioLevelNames() []string {
	return []string{"Taso I", "Taso II", "Taso III", "Taso IIIA", "Taso IV", "Taso IVA", "Taso V"}
}

func vartioServiceNames() []string {
	return []string{tesServicePerus, tesService2y, tesService7y}
}

// Vartiointialan TES 1.8.2025 tasopalkat (kk + tuntipalkka ÷ 162).
// Indeksit: 0=perus, 1=2 vuotta, 2=7 vuotta. Taso I: vain perus.
var vartioPayPKS = map[string][3]tesPayCell{
	"Taso I": {{1926, 11.89}, {1926, 11.89}, {1926, 11.89}},
	"Taso II": {{2014, 12.43}, {2087, 12.88}, {2163, 13.35}},
	"Taso III": {{2122, 13.10}, {2199, 13.57}, {2276, 14.05}},
	"Taso IIIA": {{2193, 13.54}, {2271, 14.02}, {2351, 14.51}},
	"Taso IV": {{2263, 13.97}, {2345, 14.48}, {2425, 14.97}},
	"Taso IVA": {{2351, 14.51}, {2437, 15.04}, {2519, 15.55}},
	"Taso V": {{2439, 15.06}, {2529, 15.61}, {2617, 16.15}},
}

var vartioPayMuu = map[string][3]tesPayCell{
	"Taso I": {{1860, 11.48}, {1860, 11.48}, {1860, 11.48}},
	"Taso II": {{1948, 12.02}, {2021, 12.48}, {2090, 12.90}},
	"Taso III": {{2052, 12.67}, {2129, 13.14}, {2200, 13.58}},
	"Taso IIIA": {{2120, 13.09}, {2200, 13.58}, {2273, 14.03}},
	"Taso IV": {{2188, 13.51}, {2269, 14.01}, {2348, 14.49}},
	"Taso IVA": {{2232, 13.78}, {2315, 14.29}, {2396, 14.79}},
	"Taso V": {{2359, 14.56}, {2448, 15.11}, {2532, 15.63}},
}

const (
	vartioEvening = 1.11
	vartioNight = 2.45
	vartioSaturday = 2.18
	vartioTrainingHourly = 0.25 // TES 25
	vartioTrainingMonthly = 40.0
)

func vartioServiceIndex(name string) int {
	switch name {
	case tesService2y:
		return 1
	case tesService7y:
		return 2
	default:
		return 0
	}
}

func (s *settingsTab) applyVartiointiFromSelectors() {
	if s.tesFamily == nil || s.tesFamily.Selected != tesFamilyVartio {
		return
	}
	level := "Taso IV"
	if s.tesLevel != nil && s.tesLevel.Selected != "" {
		level = s.tesLevel.Selected
	}
	pks := true
	if s.tesRegion != nil && s.tesRegion.Selected == tesRegionMuu {
		pks = false
	}
	service := tesServicePerus
	if s.experienceTier != nil && s.experienceTier.Selected != "" {
		service = s.experienceTier.Selected
	}
	s.applyVartiointiPay(level, pks, service)
}

func (s *settingsTab) applyVartiointiPay(level string, pks bool, service string) {
	table := vartioPayMuu
	regionLabel := tesRegionMuu
	if pks {
		table = vartioPayPKS
		regionLabel = tesRegionPKS
	}
	row, ok := table[level]
	if !ok {
		s.status.SetText("Tuntematon taso: " + level)
		return
	}
	idx := vartioServiceIndex(service)
	perus := row[0]
	sel := row[idx]
	exp := sel.Hourly - perus.Hourly
	if exp < 0 {
		exp = 0
	}

	s.suppressTESCallback = true
	defer func() { s.suppressTESCallback = false }()

	s.setProfileSelectorsForFamily(tesFamilyVartio)

	s.hourlyWage.SetText(fmt.Sprintf("%.2f", perus.Hourly))
	s.levelPay.SetText(fmt.Sprintf("%.2f", sel.Month))
	s.experienceAllowance.SetText(fmt.Sprintf("%.2f", exp))
	s.personalAllowance.SetText("0.00")
	if s.trainingAllowance != nil {
		s.trainingAllowance.SetText(fmt.Sprintf("%.2f", vartioTrainingHourly))
	}
	s.eveningAllowance.SetText(fmt.Sprintf("%.2f", vartioEvening))
	if s.eveningDoubleAllowance != nil {
		s.eveningDoubleAllowance.SetText("0.00")
	}
	s.nightAllowance.SetText(fmt.Sprintf("%.2f", vartioNight))
	s.saturdayAllowance.SetText(fmt.Sprintf("%.2f", vartioSaturday))
	s.sundayAllowance.SetText(fmt.Sprintf("%.2f", sel.Hourly))
	s.holidayAllowance.SetText(fmt.Sprintf("%.2f", sel.Hourly))
	s.dailyRestHours.SetText("10")
	s.restViolationPercent.SetText("50")
	s.overtime50After.SetText("12")
	s.overtime100After.SetText("18")
	s.eveningStart.set("18:00")
	s.eveningEnd.set("22:00")
	s.nightStart.set("22:00")
	s.nightEnd.set("06:00")
	if s.saturdayStart != nil {
		s.saturdayStart.set("00:00")
	}
	if s.saturdayEnd != nil {
		s.saturdayEnd.set("24:00")
	}
	if s.shiftOTEnabled != nil {
		s.shiftOTEnabled.SetChecked(true)
	}
	if s.weeklyOTEnabled != nil {
		s.weeklyOTEnabled.SetChecked(false)
	}
	if s.weeklyOTThreshold != nil {
		s.weeklyOTThreshold.SetText("37.5")
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
	s.eveningDoubleMonthFrom = 0
	s.eveningDoubleMonthTo = 0
	s.eveningDoubleSundayOnly = false
	s.calloutFixedH = calc.VartiointiCalloutFixedH
	if s.periodOTEnabled != nil {
		s.periodOTEnabled.SetChecked(true)
	}
	if s.periodMode != nil {
		s.periodMode.SetSelected(periodMode128_112)
	}
	if s.periodFirstThreshold != nil {
		s.periodFirstThreshold.SetSelected("128 h (1. jakso)")
	}

	if s.tesFamily != nil {
		s.tesFamily.SetSelected(tesFamilyVartio)
	}
	if s.tesLevel != nil {
		s.tesLevel.SetSelected(level)
	}
	if s.tesRegion != nil {
		s.tesRegion.SetSelected(regionLabel)
	}
	if s.experienceTier != nil {
		s.experienceTier.SetSelected(service)
	}

	s.status.SetText(fmt.Sprintf(
		"TES: Turvallisuusala - %s, %s, %s (1.8.2025).",
		level, regionLabel, service,
	))
	s.syncTESVisibility(tesFamilyVartio)
}
