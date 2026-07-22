package ui

import (
	"fmt"
)

// TES family / region / level labels.
const (
	tesFamilyCustom = "Oma (tyhjä)"
	tesFamilyVartio = "Turvallisuusala (Vartiointiala TES)"
	tesRegionPKS    = "PK-seutu"
	tesRegionMuu    = "Muu Suomi"
	tesServicePerus = "Perus"
	tesService2y    = "2 vuotta"
	tesService7y    = "7 vuotta"

	periodMode120     = "Kiintea 120 h / 3 vk"
	periodMode128_112 = "Tasoittuminen 128 / 112 h"

	otherModeNone   = "Ei"
	otherModeHourly = "Tunneittain (e/h)"
	otherModeFixed  = "Kiintea (e)"
)

func tesFamilyNames() []string {
	return []string{tesFamilyCustom, tesFamilyVartio}
}

func tesLevelNames() []string {
	return []string{"Taso I", "Taso II", "Taso III", "Taso IIIA", "Taso IV", "Taso IVA", "Taso V"}
}

func tesRegionNames() []string {
	return []string{tesRegionPKS, tesRegionMuu}
}

func tesServiceNames() []string {
	return []string{tesServicePerus, tesService2y, tesService7y}
}

type tesPayCell struct {
	Month  float64
	Hourly float64
}

// Vartiointialan TES 1.8.2025 tasopalkat (kk + tuntipalkka ÷ 162).
// Indeksit: 0=perus, 1=2 vuotta, 2=7 vuotta. Taso I: vain perus.
var vartioPayPKS = map[string][3]tesPayCell{
	"Taso I":    {{1926, 11.89}, {1926, 11.89}, {1926, 11.89}},
	"Taso II":   {{2014, 12.43}, {2087, 12.88}, {2163, 13.35}},
	"Taso III":  {{2122, 13.10}, {2199, 13.57}, {2276, 14.05}},
	"Taso IIIA": {{2193, 13.54}, {2271, 14.02}, {2351, 14.51}},
	"Taso IV":   {{2263, 13.97}, {2345, 14.48}, {2425, 14.97}},
	"Taso IVA":  {{2351, 14.51}, {2437, 15.04}, {2519, 15.55}},
	"Taso V":    {{2439, 15.06}, {2529, 15.61}, {2617, 16.15}},
}

var vartioPayMuu = map[string][3]tesPayCell{
	"Taso I":    {{1860, 11.48}, {1860, 11.48}, {1860, 11.48}},
	"Taso II":   {{1948, 12.02}, {2021, 12.48}, {2090, 12.90}},
	"Taso III":  {{2052, 12.67}, {2129, 13.14}, {2200, 13.58}},
	"Taso IIIA": {{2120, 13.09}, {2200, 13.58}, {2273, 14.03}},
	"Taso IV":   {{2188, 13.51}, {2269, 14.01}, {2348, 14.49}},
	"Taso IVA":  {{2232, 13.78}, {2315, 14.29}, {2396, 14.79}},
	"Taso V":    {{2359, 14.56}, {2448, 15.11}, {2532, 15.63}},
}

// Fixed allowances 1.8.2025 (same nationwide).
const (
	vartioEvening         = 1.11
	vartioNight           = 2.45
	vartioSaturday        = 2.18
	vartioTrainingHourly  = 0.25 // TES 25 § koulutuslisä
	vartioTrainingMonthly = 40.0
)

func serviceIndex(name string) int {
	switch name {
	case tesService2y:
		return 1
	case tesService7y:
		return 2
	default:
		return 0
	}
}

func (s *settingsTab) applyTESFamily(name string) {
	if s.suppressTESCallback {
		return
	}
	switch name {
	case tesFamilyVartio:
		s.setVartioSelectorsEnabled(true)
		s.applyVartiointiFromSelectors()
	default:
		s.setVartioSelectorsEnabled(false)
		s.status.SetText("TES-pohja: oma (kentät muokattavissa vapaasti).")
	}
}

func (s *settingsTab) setVartioSelectorsEnabled(on bool) {
	// Fyne Select has no Disable that greys out reliably on all platforms;
	// we still apply only when family is Vartio.
	_ = on
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
	idx := serviceIndex(service)
	perus := row[0]
	sel := row[idx]
	exp := sel.Hourly - perus.Hourly
	if exp < 0 {
		exp = 0
	}

	s.suppressTESCallback = true
	defer func() { s.suppressTESCallback = false }()

	s.hourlyWage.SetText(fmt.Sprintf("%.2f", perus.Hourly))
	s.levelPay.SetText(fmt.Sprintf("%.2f", sel.Month))
	s.experienceAllowance.SetText(fmt.Sprintf("%.2f", exp))
	s.personalAllowance.SetText("0.00")
	if s.trainingAllowance != nil {
		s.trainingAllowance.SetText(fmt.Sprintf("%.2f", vartioTrainingHourly))
	}
	// Keep user's koulutuslisä on/off; only refresh the TES amount.
	s.eveningAllowance.SetText(fmt.Sprintf("%.2f", vartioEvening))
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
		"TES: Turvallisuusala — %s, %s, %s (1.8.2025).",
		level, regionLabel, service,
	))
}

// applyDemoTES loads demo defaults: Taso IV, PK-seutu, perus, jakso 128/112.
func (s *settingsTab) applyDemoTES() {
	s.applyVartiointiPay("Taso IV", true, tesServicePerus)
	if s.periodAnchor != nil {
		// Monday 6.7.2026 — period covering demo roster start (20.7).
		s.periodAnchor.SetText("06.07.2026")
	}
}

// applyExperienceTier is kept for the palvelusaika select callback.
func (s *settingsTab) applyExperienceTier(tier string) {
	if s.suppressTESCallback {
		return
	}
	if s.tesFamily != nil && s.tesFamily.Selected == tesFamilyVartio {
		s.applyVartiointiFromSelectors()
		return
	}
	_ = tier
}
