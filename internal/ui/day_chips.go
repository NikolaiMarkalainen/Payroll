package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func chipColor(code string) color.NRGBA {
	switch code {
	case codeCallout:
		return color.NRGBA{R: 0xc4, G: 0x5c, B: 0x26, A: 0xff} // burnt orange
	case codePerehdytys:
		return color.NRGBA{R: 0x3d, G: 0x7a, B: 0x4a, A: 0xff} // forest green
	case codeSunday:
		return color.NRGBA{R: 0xb3, G: 0x2d, B: 0x3a, A: 0xff} // crimson
	case codeHoliday:
		return color.NRGBA{R: 0x8b, G: 0x1e, B: 0x3f, A: 0xff} // wine
	case codeEvening:
		return color.NRGBA{R: 0xb8, G: 0x86, B: 0x0b, A: 0xff} // gold
	case codeNight:
		return color.NRGBA{R: 0x2f, G: 0x5d, B: 0x8a, A: 0xff} // steel blue
	case codeSaturday:
		return color.NRGBA{R: 0x2a, G: 0x7a, B: 0x6a, A: 0xff} // teal
	case codeOvertime50:
		return color.NRGBA{R: 0x8a, G: 0x5a, B: 0x2b, A: 0xff} // amber brown
	case codeOvertime100:
		return color.NRGBA{R: 0x5c, G: 0x3a, B: 0x21, A: 0xff} // dark brown
	default:
		return color.NRGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xff}
	}
}

func allowanceLegend() fyne.CanvasObject {
	items := []struct {
		code string
		name string
	}{
		{codeCallout, "hälyt"},
		{codePerehdytys, "pere"},
		{codeSunday, "su"},
		{codeHoliday, "pyhä"},
		{codeEvening, "ilta"},
		{codeNight, "yö"},
		{codeSaturday, "la"},
		{codeOvertime50, "ylityö"},
		{codeOvertime100, "ylityö"},
	}
	parts := make([]fyne.CanvasObject, 0, len(items)*2)
	for i, it := range items {
		if i > 0 {
			sep := canvas.NewText("|", color.NRGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xff})
			sep.TextSize = 11
			parts = append(parts, sep)
		}
		dot := canvas.NewText(it.code, chipColor(it.code))
		dot.TextStyle = fyne.TextStyle{Bold: true}
		dot.TextSize = 11
		name := canvas.NewText(it.name, color.NRGBA{R: 0x66, G: 0x66, B: 0x66, A: 0xff})
		name.TextSize = 11
		parts = append(parts, container.NewHBox(dot, name))
	}
	return container.NewHBox(parts...)
}

func chipRow(chips []allowanceChip) fyne.CanvasObject {
	if len(chips) == 0 {
		return widget.NewLabel("")
	}
	objs := make([]fyne.CanvasObject, 0, len(chips))
	for _, c := range chips {
		txt := canvas.NewText(fmt.Sprintf("%s%s", c.Code, formatChipHours(c.Hours)), chipColor(c.Code))
		txt.TextStyle = fyne.TextStyle{Bold: true}
		txt.TextSize = 11
		objs = append(objs, txt)
	}
	return container.NewHBox(objs...)
}
