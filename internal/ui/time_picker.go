package ui

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// timePicker is HH:MM via numeric entry fields (hour 1–24, minute 0–59).
type timePicker struct {
	hour   *widget.Entry
	minute *widget.Entry
	err    *widget.Label
	box    *fyne.Container
}

func newTimePicker(initial string) *timePicker {
	h, m := parseClock(initial)
	tp := &timePicker{
		hour:   newClockEntry(h),
		minute: newClockEntry(m),
		err:    widget.NewLabel(""),
	}
	tp.err.Importance = widget.DangerImportance
	tp.err.Hide()

	fields := container.NewHBox(tp.hour, widget.NewLabel(":"), tp.minute)
	tp.box = container.NewVBox(fields, tp.err)

	tp.hour.OnChanged = func(string) {
		tp.filterEntry(tp.hour)
		tp.refreshError()
	}
	tp.minute.OnChanged = func(string) {
		tp.filterEntry(tp.minute)
		tp.refreshError()
	}
	return tp
}

func newClockEntry(initial string) *widget.Entry {
	e := widget.NewEntry()
	e.SetText(stripLeadingZeros(initial))
	e.SetPlaceHolder("0")
	return e
}

func (tp *timePicker) filterEntry(e *widget.Entry) {
	filtered := filterDigits(e.Text)
	if filtered != e.Text {
		e.SetText(filtered)
	}
}

func filterDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if len(out) > 2 {
		out = out[:2]
	}
	return out
}

func stripLeadingZeros(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return s
	}
	return strconv.Itoa(n)
}

func (tp *timePicker) canvas() fyne.CanvasObject {
	return tp.box
}

func (tp *timePicker) value() string {
	h, m, ok := tp.parsed()
	if !ok {
		return "00:00"
	}
	return fmt.Sprintf("%02d:%02d", h, m)
}

func (tp *timePicker) set(value string) {
	h, m := parseClock(value)
	tp.hour.SetText(stripLeadingZeros(h))
	tp.minute.SetText(stripLeadingZeros(m))
	tp.refreshError()
}

func (tp *timePicker) refreshError() {
	if tp.valid() {
		tp.err.SetText("")
		tp.err.Hide()
		return
	}
	tp.err.SetText("Virheellinen aika (tunnit 1–24, minuutit 0–59)")
	tp.err.Show()
}

func (tp *timePicker) parsed() (hour, minute int, ok bool) {
	h, err1 := strconv.Atoi(strings.TrimSpace(tp.hour.Text))
	m, err2 := strconv.Atoi(strings.TrimSpace(tp.minute.Text))
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	if h < 1 || h > 24 || m < 0 || m > 59 {
		return 0, 0, false
	}
	// 24:00 is the only valid "24" clock time.
	if h == 24 && m != 0 {
		return 0, 0, false
	}
	return h, m, true
}

func (tp *timePicker) valid() bool {
	_, _, ok := tp.parsed()
	return ok
}

func parseClock(s string) (hour, minute string) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) != 2 {
		return "1", "0"
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return "1", "0"
	}
	if h == 0 && m >= 0 && m <= 59 {
		return "24", fmt.Sprintf("%d", m)
	}
	if h < 1 || h > 24 || m < 0 || m > 59 {
		return "1", "0"
	}
	if h == 24 && m != 0 {
		return "1", "0"
	}
	return fmt.Sprintf("%d", h), fmt.Sprintf("%d", m)
}
