package ui

// TES family dispatch and pure helpers (no Fyne layout). Keep switches here, not in tes_common.

func featuresForFamily(family string) tesUIFeatures {
	switch family {
	case tesFamilyVartio:
		return tesUIFeatures{
			LevelLabel:       "Taso",
			ProfileSelectors: true,
			Training:         true,
			PeriodOT:         true,
			ShiftOT:          true,
		}
	case tesFamilyKauppa:
		return tesUIFeatures{
			LevelLabel:       "Palkkaryhmä",
			ProfileSelectors: true,
			EveningDouble:    true,
			WeeklyOT:         true,
			KaupanFlags:      true,
		}
	default: // Oma (tyhjä) - barebones: ei TES-valitsimia, vain lisät + kiinteä 120 h / 3 vk
		return tesUIFeatures{
			PeriodOT:       true,
			PeriodFixed120: true,
		}
	}
}

// profileSelectorSpec is level/service options for a TES family.
type profileSelectorSpec struct {
	Levels []string
	Services []string
	DefaultLevel string
	DefaultService string
	KeepCurrentOpts bool // if true, leave level/service Options as-is
}

func profileSelectorsForFamily(family string) profileSelectorSpec {
	switch family {
	case tesFamilyKauppa:
		return profileSelectorSpec{
			Levels: kaupanGroupNames(),
			Services: kaupanServiceNames(),
			DefaultLevel: "B",
			DefaultService: kaupanService2y,
		}
	case tesFamilyVartio:
		return profileSelectorSpec{
			Levels: vartioLevelNames(),
			Services: vartioServiceNames(),
			DefaultLevel: "Taso IV",
			DefaultService: tesServicePerus,
		}
	default:
		// Oma: no level/service options — clear any leftover TES selection.
		return profileSelectorSpec{}
	}
}

func profileHintText(family string) string {
	switch family {
	case tesFamilyVartio:
		return "Vartiointiala (PAM/PALTA). Palkat 1.8.2025. Su/pyhä = 100 % tuntipalkasta. Koulutuslisä TES 25."
	case tesFamilyKauppa:
		return "Kaupan TES myyjät (PAM/Kaupan liitto). Taulukot 1.5.2025. Su/pyhä = 100 % tuntipalkasta. Ei koulutuslisää tässä profiilissa."
	default:
		return "Oma pohja (barebones): syötä tuntipalkka ja lisät itse. Ei alue-/taso-/palvelusaikavalitsimia; jakso = kiinteä 120 h / 3 vk."
	}
}

func expHintText(family string) string {
	switch family {
	case tesFamilyVartio:
		return "Kokemuslisä = palvelusaikalisän erotus perustasoon. Henkilökohtainen lisä (TES 26) erikseen. " +
			"Koulutuslisä (TES 25): 40 e/kk tai 0.25 e/h; ei kuulu ylityön perustuntipalkkaan. " +
			"Perehdytyslisä: 1.00 e/h merkityille perehdytystunneille vuorossa (PERE). " +
			"Muu lisä: tunneittain tai kiinteä summa jaksolle."
	case tesFamilyKauppa:
		return "Taulukkopalkka: tuntipalkka = 2 v -perus; palvelusvuodet (4/6/9) -> kokemuslisä e/h erikseen. " +
			"Henkilökohtainen lisä ja muu lisä tarvittaessa. Muu lisä: tunneittain tai kiinteä summa jaksolle."
	default:
		return "Syötä tarvittaessa kokemuslisä (e/h), henkilökohtainen lisä ja muu lisä."
	}
}

func otHintText(family string) string {
	switch family {
	case tesFamilyVartio:
		return "TES 31: jos vuoro ylittää 12 h (pidennys), ylimenevät tunnit kuten ylityö 50 %. " +
			"Jakson aikana näistä ensimmäiset 18 h @ 50 %, loput @ 100 %. Nämä tunnit eivät kuulu jaksoylityöhön. " +
			"Jaksoylityö (29): ensimmäiset 18 ylityötuntia 50 %, loput 100 %."
	case tesFamilyKauppa:
		return "Vuorokausi >10 h ja/tai viikko >37,5 h -> +50 %. Ei jaksoylityötä. Iltalisä 2x = marras-joulu su."
	default:
		return "Jaksoylityö: kiinteä 120 h / 3 vk. Ilta-/yö-/la-/su-/pyhälisät syötetään yllä."
	}
}


func personalHeadingText(hasTraining bool) string {
	if hasTraining {
		return "Kokemus-, koulutus- ja henkilökohtainen lisä"
	}
	return "Kokemus- ja henkilökohtainen lisä"
}

func containsString(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// dispatchTESFamilyApply runs the correct TES apply path after family select.
func dispatchTESFamilyApply(s *settingsTab, name string) {
	switch name {
	case tesFamilyVartio:
		s.setProfileSelectorsForFamily(tesFamilyVartio)
		s.syncTESVisibility(tesFamilyVartio)
		s.applyVartiointiFromSelectors()
	case tesFamilyKauppa:
		s.setProfileSelectorsForFamily(tesFamilyKauppa)
		s.syncTESVisibility(tesFamilyKauppa)
		s.applyKaupanFromSelectors()
	default:
		s.setProfileSelectorsForFamily(tesFamilyCustom)
		s.applyCustomBarebones()
		s.syncTESVisibility(tesFamilyCustom)
		s.status.SetText("TES-pohja: oma (barebones) - tuntipalkka + lisät; jakso kiinteä 120 h / 3 vk.")
	}
}

// dispatchTESFromSelectors re-applies pay when level/region/service changes.
func dispatchTESFromSelectors(s *settingsTab, family string) {
	switch family {
	case tesFamilyVartio:
		s.applyVartiointiFromSelectors()
	case tesFamilyKauppa:
		s.applyKaupanFromSelectors()
	}
}
