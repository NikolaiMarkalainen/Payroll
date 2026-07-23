package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"time"
)

const appID = "fi.palkkatarkistus.app"

// Options configures app startup.
type Options struct {
	Demo bool // load sample roster and open Vuorot
}

func Run() {
	RunWith(Options{})
}

func RunWith(opts Options) {
	a := app.NewWithID(appID)
	w := a.NewWindow("Palkkatarkistus")
	w.Resize(fyne.NewSize(1100, 720))
	w.SetMaster()

	content, tabs, shifts := buildUI(w)
	if opts.Demo {
		shifts.loadDemoSeed()
		tabs.SelectIndex(1)
		w.SetTitle("Palkkatarkistus (demo)")
	}
	w.SetContent(content)

	w.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Tiedosto",
			fyne.NewMenuItem("Lopeta", func() { a.Quit() }),
		),
		fyne.NewMenu("Näytä",
			fyne.NewMenuItem("Asetukset", func() { tabs.SelectIndex(0) }),
			fyne.NewMenuItem("Vuorot", func() { tabs.SelectIndex(1) }),
			fyne.NewMenuItem("PDF-tuonti", func() { tabs.SelectIndex(2) }),
			fyne.NewMenuItem("Laskelma", func() { tabs.SelectIndex(3) }),
			fyne.NewMenuItem("Vertailu", func() { tabs.SelectIndex(4) }),
		),
	))

	w.ShowAndRun()
}

// buildUI constructs the main layout and returns it with the tab bar for tests.
func buildUI(w fyne.Window) (fyne.CanvasObject, *container.AppTabs, *shiftsTab) {
	settings := newSettingsTab()
	settings.window = w
	shifts := newShiftsTab(w)
	shifts.rules = settings.allowanceRules
	shifts.colorTitles = settings.colorShiftTitlesEnabled
	shifts.colorFor = settings.colorForShiftCode
	shifts.periodAnchor = settings.periodAnchorDate
	shifts.periodThreshold = func(day time.Time) (float64, bool) {
		if _, ok := settings.periodAnchorDate(); !ok {
			return 0, false
		}
		return settings.periodThresholdForRange(day), true
	}
	settings.shiftsSource = func() []calendarShift { return shifts.shifts }
	shifts.onChanged = func() { settings.refreshColorRows() }
	settings.onSaved = func() { shifts.refresh() }
	calcView := newCalcTab(settings, shifts)
	shifts.selectedPeriod = func() (time.Time, time.Time, bool) {
		from, err1 := parseFIDate(calcView.from.Text)
		to, err2 := parseFIDate(calcView.to.Text)
		if err1 != nil || err2 != nil {
			return time.Time{}, time.Time{}, false
		}
		return from, to, true
	}
	calcView.onPeriodRangeChanged = func() { shifts.refresh() }

	// Keep Laskelma ankkuri in sync with Asetukset; refresh calendar period highlights.
	if settings.periodAnchor != nil {
		settings.periodAnchor.OnChanged = func(s string) {
			if calcView.suppressPeriod {
				return
			}
			calcView.suppressPeriod = true
			calcView.periodAnchor.SetText(s)
			calcView.suppressPeriod = false
			calcView.refreshPeriodOptions()
			shifts.refresh()
		}
	}

	pdfView := newPDFTab(w, shifts, calcView)

	tabs := container.NewAppTabs(
		container.NewTabItem("Asetukset", settings.canvas()),
		container.NewTabItem("Vuorot", shifts.canvas()),
		container.NewTabItem("PDF-tuonti", pdfView.canvas()),
		container.NewTabItem("Laskelma", calcView.canvas()),
		container.NewTabItem("Vertailu", emptyTab("Vertailu")),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	// Demo loads roster + calc date range only. Do NOT apply Vartiointi TES —
	// that left profile selectors (taso) sticky when switching back to Oma.
	shifts.onDemoLoaded = func() {
		calcView.setDemoRange()
	}

	header := widget.NewLabel("Palkkatarkistus")
	header.TextStyle = fyne.TextStyle{Bold: true}
	subtitle := widget.NewLabel("Vertaa TES-pohjaista laskelmaa maksettuun palkkaan.")
	subtitle.Wrapping = fyne.TextWrapWord

	content := container.NewBorder(
		container.NewVBox(header, subtitle, widget.NewSeparator()),
		nil,
		nil,
		nil,
		tabs,
	)
	return content, tabs, shifts
}

func emptyTab(name string) fyne.CanvasObject {
	label := widget.NewLabel(name + " - sisältö lisätään myöhemmin.")
	label.Wrapping = fyne.TextWrapWord
	return container.NewPadded(label)
}
