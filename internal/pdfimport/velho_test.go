package pdfimport

import (
	"path/filepath"
	"testing"
	"time"
)

func TestParseVelhoSamplePDF(t *testing.T) {
	path := filepath.Join("testdata", "henkilokohtainen-2.pdf")
	res, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !res.From.Equal(time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)) {
		t.Fatalf("from=%v", res.From)
	}
	if !res.To.Equal(time.Date(2026, 7, 15, 0, 0, 0, 0, time.Local)) {
		t.Fatalf("to=%v", res.To)
	}
	if res.Period != "13/2026" {
		t.Fatalf("period=%q", res.Period)
	}

	want := []struct {
		day, start, end string
		callout         bool
	}{
		{"1.7.2026", "04:55", "14:30", false},
		{"3.7.2026", "04:30", "14:30", false}, // realized, not planned 04:40
		{"4.7.2026", "04:30", "14:30", false},
		{"5.7.2026", "05:30", "15:00", false},
		{"6.7.2026", "08:00", "16:00", true},  // Hälytys aika 2:00 → Callout
		{"11.7.2026", "21:48", "05:00", true}, // overnight callout + Hälytys 2:00
		{"15.7.2026", "04:55", "14:30", false},
	}
	if len(res.Shifts) != len(want) {
		for _, s := range res.Shifts {
			t.Logf("got %s %s-%s callout=%v code=%s", s.Date.Format("2.1.2006"), s.Start, s.End, s.Callout, s.Code)
		}
		t.Fatalf("shifts=%d want %d", len(res.Shifts), len(want))
	}
	for i, w := range want {
		got := res.Shifts[i]
		day := got.Date.Format("2.1.2006")
		if day != w.day || got.Start != w.start || got.End != w.end || got.Callout != w.callout {
			t.Fatalf("[%d] got %s %s-%s callout=%v want %s %s-%s callout=%v",
				i, day, got.Start, got.End, got.Callout, w.day, w.start, w.end, w.callout)
		}
	}
}
