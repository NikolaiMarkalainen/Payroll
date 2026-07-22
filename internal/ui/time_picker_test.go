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
	if tp.err.Visible() {
		t.Fatal("expected no error for valid time")
	}
}

func TestTimePickerExactMinutes(t *testing.T) {
	test.NewApp()

	tp := newTimePicker("06:00")
	tp.set("06:47")
	if got := tp.value(); got != "06:47" {
		t.Fatalf("value=%q want 06:47", got)
	}
}

func TestTimePickerChangeHourMinute(t *testing.T) {
	test.NewApp()

	tp := newTimePicker("06:00")
	tp.hour.SetText("9")
	tp.minute.SetText("30")
	tp.refreshError()
	if got := tp.value(); got != "09:30" {
		t.Fatalf("value=%q want 09:30", got)
	}
}

func TestTimePickerShowsErrorWhenInvalid(t *testing.T) {
	test.NewApp()

	tp := newTimePicker("12:00")
	tp.hour.SetText("25")
	tp.refreshError()
	if !tp.err.Visible() || tp.err.Text == "" {
		t.Fatal("expected error label for invalid hour")
	}

	tp.hour.SetText("12")
	tp.minute.SetText("60")
	tp.refreshError()
	if !tp.err.Visible() {
		t.Fatal("expected error for minute 60")
	}

	tp.minute.SetText("0")
	tp.hour.SetText("24")
	tp.refreshError()
	if tp.err.Visible() {
		t.Fatal("24:00 should be valid")
	}

	tp.minute.SetText("1")
	tp.refreshError()
	if !tp.err.Visible() {
		t.Fatal("24:01 should be invalid")
	}
}

func TestParseClockInvalid(t *testing.T) {
	h, m := parseClock("bad")
	if h != "1" || m != "0" {
		t.Fatalf("got %s:%s", h, m)
	}
	h, m = parseClock("25:99")
	if h != "1" || m != "0" {
		t.Fatalf("got %s:%s", h, m)
	}
}

func TestParseClockMidnightZero(t *testing.T) {
	h, m := parseClock("00:00")
	if h != "24" || m != "0" {
		t.Fatalf("got %s:%s want 24:0", h, m)
	}
}

func TestFilterDigits(t *testing.T) {
	if got := filterDigits("1a2b"); got != "12" {
		t.Fatalf("got %q", got)
	}
	if got := filterDigits("1234"); got != "12" {
		t.Fatalf("got %q", got)
	}
}
