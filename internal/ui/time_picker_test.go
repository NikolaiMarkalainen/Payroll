package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestTimePickerDefaults(t *testing.T) {
	test.NewApp()

	tp := newTimePicker("18:00")
	if got := tp.value(); got != "18:00" {
		t.Fatalf("value=%q want 18:00", got)
	}
}

func TestTimePickerSetAndSnap(t *testing.T) {
	test.NewApp()

	tp := newTimePicker("00:00")
	tp.set("22:17") // snaps down to 22:15
	if got := tp.value(); got != "22:15" {
		t.Fatalf("value=%q want 22:15", got)
	}
}

func TestTimePickerChangeHourMinute(t *testing.T) {
	test.NewApp()

	tp := newTimePicker("06:00")
	tp.hour.SetSelected("09")
	tp.minute.SetSelected("30")
	if got := tp.value(); got != "09:30" {
		t.Fatalf("value=%q want 09:30", got)
	}
}

func TestParseClockInvalid(t *testing.T) {
	h, m := parseClock("bad")
	if h != "00" || m != "00" {
		t.Fatalf("got %s:%s", h, m)
	}
	h, m = parseClock("25:99")
	if h != "00" || m != "00" {
		t.Fatalf("got %s:%s", h, m)
	}
}
