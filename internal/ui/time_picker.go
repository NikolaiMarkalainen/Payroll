package ui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// timePicker is a simple HH:MM clock control (hour + minute selects).
type timePicker struct {
	hour   *widget.Select
	minute *widget.Select
	box    *fyne.Container
}

func newTimePicker(initial string) *timePicker {
	hours := make([]string, 24)
	for i := 0; i < 24; i++ {
		hours[i] = fmt.Sprintf("%02d", i)
	}
	mins := make([]string, 12) // 5-minute steps
	for i := 0; i < 12; i++ {
		mins[i] = fmt.Sprintf("%02d", i*5)
	}

	h, m := parseClock(initial)
	tp := &timePicker{
		hour:   widget.NewSelect(hours, nil),
		minute: widget.NewSelect(mins, nil),
	}
	tp.hour.SetSelected(h)
	tp.minute.SetSelected(snapMinute(m))
	tp.hour.PlaceHolder = "hh"
	tp.minute.PlaceHolder = "mm"
	tp.box = container.NewHBox(tp.hour, widget.NewLabel(":"), tp.minute)
	return tp
}

func (tp *timePicker) canvas() fyne.CanvasObject {
	return tp.box
}

func (tp *timePicker) value() string {
	h := tp.hour.Selected
	m := tp.minute.Selected
	if h == "" {
		h = "00"
	}
	if m == "" {
		m = "00"
	}
	return h + ":" + m
}

func (tp *timePicker) set(value string) {
	h, m := parseClock(value)
	tp.hour.SetSelected(h)
	tp.minute.SetSelected(snapMinute(m))
}

func parseClock(s string) (hour, minute string) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) != 2 {
		return "00", "00"
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return "00", "00"
	}
	return fmt.Sprintf("%02d", h), fmt.Sprintf("%02d", m)
}

func snapMinute(m string) string {
	n, err := strconv.Atoi(m)
	if err != nil {
		return "00"
	}
	n = (n / 5) * 5
	return fmt.Sprintf("%02d", n)
}
