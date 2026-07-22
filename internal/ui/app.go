package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const appID = "com.calculator.payroll"

func Run() {
	a := app.NewWithID(appID)
	w := a.NewWindow("Payroll")
	w.Resize(fyne.NewSize(800, 600))
	w.SetMaster()

	w.SetContent(buildMainContent())
	w.ShowAndRun()
}

func buildMainContent() fyne.CanvasObject {
	title := widget.NewLabel("Payroll")
	title.TextStyle = fyne.TextStyle{Bold: true}
 
	subtitle := widget.NewLabel("Compare your payroll to verify you get proper compensation based on your worked hours.")
	subtitle.Wrapping = fyne.TextWrapWord

	return container.NewBorder(
		container.NewVBox(title, subtitle, widget.NewSeparator()),
		nil,
		nil,
		nil,
		widget.NewLabel("Ready."),
	)
}
