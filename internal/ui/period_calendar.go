package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"payroll/internal/calc"
)

// Year period grid: J1 starts every 12 January (e.g. 12.01.2026 - 01.02.2026).
// Numbers run J1, J2, ... through the year; never negative.
const (
	periodYearStartMonth = time.January
	periodYearStartDay   = 12
)

// Distinct badge colors per jakso (J1, J2, ...) — no cell background tint.
var periodBadgePalette = []color.NRGBA{
	{R: 0x6a, G: 0xb0, B: 0xe8, A: 0xff}, // sky
	{R: 0xb0, G: 0x8a, B: 0xe0, A: 0xff}, // violet
	{R: 0x6a, G: 0xc4, B: 0x8a, A: 0xff}, // green
	{R: 0xe0, G: 0xa8, B: 0x5a, A: 0xff}, // amber
	{R: 0xe0, G: 0x7a, B: 0x9a, A: 0xff}, // rose
	{R: 0x5a, G: 0xc4, B: 0xc8, A: 0xff}, // teal
}

// periodYearAnchorFor returns 12.01 of day's year, or previous year if day is before that.
func periodYearAnchorFor(day time.Time) time.Time {
	loc := day.Location()
	d := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
	start := time.Date(d.Year(), periodYearStartMonth, periodYearStartDay, 0, 0, 0, 0, loc)
	if d.Before(start) {
		return time.Date(d.Year()-1, periodYearStartMonth, periodYearStartDay, 0, 0, 0, 0, loc)
	}
	return start
}

// defaultPeriodYearAnchor is 12.01 of the current payroll year (for settings default).
func defaultPeriodYearAnchor(now time.Time) time.Time {
	return periodYearAnchorFor(now)
}

// periodYearNumber is 1-based jakso number within the year grid (always >= 1).
func periodYearNumber(day time.Time) int {
	anchor := periodYearAnchorFor(day)
	return calc.PeriodIndexContaining(anchor, day) + 1
}

func periodBadgeColor(yearNum int) color.NRGBA {
	if len(periodBadgePalette) == 0 {
		return color.NRGBA{R: 0xaa, G: 0xaa, B: 0xaa, A: 0xff}
	}
	// yearNum is 1-based
	i := (yearNum - 1) % len(periodBadgePalette)
	if i < 0 {
		i += len(periodBadgePalette)
	}
	return periodBadgePalette[i]
}

func periodLegend() fyne.CanvasObject {
	txt := canvas.NewText("J1 alkaa 12.01 (esim. 12.01-01.02.2026). Numerointi J1, J2... vuoden loppuun; oma väri per jakso.", color.NRGBA{R: 0xaa, G: 0xaa, B: 0xaa, A: 0xff})
	txt.TextSize = 11
	return txt
}

func (s *shiftsTab) periodInfo(date time.Time) (yearNum int, label string, selected bool, ok bool) {
	yearAnchor := periodYearAnchorFor(date)
	idx0 := calc.PeriodIndexContaining(yearAnchor, date)
	yearNum = idx0 + 1
	label = "J" + strconv.Itoa(yearNum)
	from, _ := calc.PeriodIndexWindow(yearAnchor, idx0)
	if sameDate(date, from) && s.periodThreshold != nil {
		if th, thOK := s.periodThreshold(date); thOK && th > 0 {
			label = fmt.Sprintf("%s / %.0fh", label, th)
		}
	}
	if s.selectedPeriod != nil {
		selFrom, selTo, selOK := s.selectedPeriod()
		if selOK && !date.Before(selFrom) && !date.After(selTo) {
			selected = true
		}
	}
	return yearNum, label, selected, true
}

func (s *shiftsTab) dayHeader(date time.Time, isToday bool) fyne.CanvasObject {
	dayLabel := widget.NewLabel(fmt.Sprintf("%d", date.Day()))
	dayLabel.TextStyle = fyne.TextStyle{Bold: true}
	if isToday {
		dayLabel.Importance = widget.HighImportance
	}

	yearNum, periodLabel, selected, ok := s.periodInfo(date)
	if !ok {
		return dayLabel
	}

	badge := canvas.NewText(periodLabel, periodBadgeColor(yearNum))
	badge.TextStyle = fyne.TextStyle{Bold: true}
	badge.TextSize = theme.CaptionTextSize()
	if selected {
		badge.TextSize = theme.TextSize()
	}

	return container.NewBorder(nil, nil, dayLabel, badge, layout.NewSpacer())
}
