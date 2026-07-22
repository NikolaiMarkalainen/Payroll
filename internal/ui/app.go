package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const appID = "com.nikolaimarkalainen.payroll"

func Run() {
	a := app.NewWithID(appID)
	w := a.NewWindow("Palkkatarkistus")
	w.Resize(fyne.NewSize(960, 640))
	w.SetMaster()

	content, tabs := buildUI()
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
func buildUI() (fyne.CanvasObject, *container.AppTabs) {
	settings := newSettingsTab()

	tabs := container.NewAppTabs(
		container.NewTabItem("Asetukset", settings.canvas()),
		container.NewTabItem("Vuorot", emptyTab("Vuorot")),
		container.NewTabItem("PDF-tuonti", emptyTab("PDF-tuonti")),
		container.NewTabItem("Laskelma", emptyTab("Laskelma")),
		container.NewTabItem("Vertailu", emptyTab("Vertailu")),
	)
	tabs.SetTabLocation(container.TabLocationTop)

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
	return content, tabs
}

func emptyTab(name string) fyne.CanvasObject {
	label := widget.NewLabel(name + " — sisältö lisätään myöhemmin.")
	label.Wrapping = fyne.TextWrapWord
	return container.NewPadded(label)
}
