package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"payroll/internal/assets"
)

const appID = "fi.palkkatarkistus.app"

// Options configures app startup.
type Options struct {
	Demo bool // load sample roster and open Vuorot (does not persist)
}

func Run() {
	RunWith(Options{})
}

func RunWith(opts Options) {
	a := app.NewWithID(appID)
	a.SetIcon(assets.AppIcon())
	w := a.NewWindow("Palkkatarkistus")
	w.SetIcon(assets.AppIcon())
	w.Resize(fyne.NewSize(1100, 720))
	w.SetMaster()

	content, tabs, shifts, settings, calcView := buildUI(w)

	var persister *appPersister
	if !opts.Demo {
		dir, err := defaultDataDir()
		if err == nil {
			persister = newAppPersister(dir, settings, shifts, calcView)
			if err := persister.load(); err != nil {
				settings.status.SetText("Tallennuksen lataus epäonnistui: " + err.Error())
			}
		}
	}

	wirePersistence(settings, shifts, calcView, persister)

	if opts.Demo {
		shifts.loadDemoSeed()
		tabs.SelectIndex(1)
		w.SetTitle("Palkkatarkistus (demo)")
	}
	w.SetContent(content)

	w.SetCloseIntercept(func() {
		if persister != nil {
			persister.flush()
		}
		w.Close()
	})

	w.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Tiedosto",
			fyne.NewMenuItem("Lopeta", func() {
				if persister != nil {
					persister.flush()
				}
				a.Quit()
			}),
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

func wirePersistence(settings *settingsTab, shifts *shiftsTab, calcView *calcTab, p *appPersister) {
	prevOnSaved := settings.onSaved
	settings.onSaved = func() {
		if prevOnSaved != nil {
			prevOnSaved()
		}
		if p == nil {
			settings.status.SetText("Tallennus ei käytössä (ei datakansiota).")
			return
		}
		if err := p.saveNow(); err != nil {
			settings.status.SetText("Tallennus epäonnistui: " + err.Error())
			return
		}
		settings.status.SetText(fmt.Sprintf("Tallennettu: %s", statePath(p.dir)))
	}
	settings.onPersist = func() {
		if p != nil {
			p.scheduleSave()
		}
	}

	prevShiftChanged := shifts.onChanged
	shifts.onChanged = func() {
		if prevShiftChanged != nil {
			prevShiftChanged()
		}
		if p != nil {
			p.scheduleSave()
		}
	}

	calcView.onPersist = func() {
		if p != nil {
			p.scheduleSave()
		}
	}
}

// buildUI constructs the main layout and returns it with the tab bar for tests.
func buildUI(w fyne.Window) (fyne.CanvasObject, *container.AppTabs, *shiftsTab, *settingsTab, *calcTab) {
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
	return content, tabs, shifts, settings, calcView
}

func emptyTab(name string) fyne.CanvasObject {
	label := widget.NewLabel(name + " - sisältö lisätään myöhemmin.")
	label.Wrapping = fyne.TextWrapWord
	return container.NewPadded(label)
}
