package ui

// Shared TES / settings labels and types (no switch dispatch here).
const (
	tesFamilyCustom = "Oma (tyhjä)"
	tesFamilyVartio = "Turvallisuusala (Vartiointiala TES)"
	tesFamilyKauppa = "Kaupan TES (myyjät)"

	tesRegionPKS = "PK-seutu"
	tesRegionMuu = "Muu Suomi"

	periodMode120 = "Kiintea 120 h / 3 vk"
	periodMode128_112 = "Tasoittuminen 128 / 112 h"

	otherModeNone = "Ei"
	otherModeHourly = "Tunneittain (e/h)"
	otherModeFixed = "Kiintea (e)"
)

func tesFamilyNames() []string {
	return []string{tesFamilyCustom, tesFamilyVartio, tesFamilyKauppa}
}

func tesRegionNames() []string {
	return []string{tesRegionPKS, tesRegionMuu}
}

type tesPayCell struct {
	Month float64
	Hourly float64
}

// tesUIFeatures controls which shared settings widgets a TES shows.
// Only fields that belong to the agreement should be true.
type tesUIFeatures struct {
	LevelLabel        string // form label for tesLevel ("Taso" / "Palkkaryhmä")
	ProfileSelectors  bool   // taso/palkkaryhmä, alue, palvelusaika
	Training          bool
	EveningDouble     bool
	WeeklyOT          bool
	PeriodOT          bool
	PeriodFixed120    bool // Oma barebones: only kiinteä 120 h / 3 vk (no 128/112 UI)
	ShiftOT           bool
	KaupanFlags       bool // la-ikkuna + evening/night exclusion checks
}
