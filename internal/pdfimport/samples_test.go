package pdfimport

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

type sampleExpect struct {
	file    string
	from    string // 2.1.2006
	to      string
	period  string
	shifts  int
	checks  []shiftCheck
	// atLeastOneCallout requires Callout on some shift (Hälytys aika page).
	atLeastOneCallout bool
	// noZeroDuration forbids 00:00-00:00 style imports.
	noZeroDuration bool
}

type shiftCheck struct {
	day, start, end string
	callout         *bool // nil = ignore
}

func boolPtr(v bool) *bool { return &v }

func TestParseAllSamplePDFs(t *testing.T) {
	cases := []sampleExpect{
		{
			file: "henkilokohtainen.pdf",
			from: "1.6.2026", to: "15.6.2026", period: "11/2026",
			shifts: 8, noZeroDuration: true,
			checks: []shiftCheck{
				{day: "1.6.2026", start: "04:55", end: "16:50"},
				{day: "3.6.2026", start: "14:30", end: "00:20"}, // crosses midnight
				{day: "15.6.2026", start: "08:00", end: "16:00"},
			},
		},
		{
			file: "henkilokohtainen-1.pdf",
			from: "16.6.2026", to: "30.6.2026", period: "12/2026",
			shifts: 9, noZeroDuration: true,
			checks: []shiftCheck{
				{day: "16.6.2026", start: "04:55", end: "14:30"},
				{day: "22.6.2026", start: "16:55", end: "00:15"},
				{day: "28.6.2026", start: "05:45", end: "15:00"},
			},
		},
		{
			file: "henkilokohtainen-2.pdf",
			from: "1.7.2026", to: "15.7.2026", period: "13/2026",
			shifts: 7, noZeroDuration: true, atLeastOneCallout: true,
			checks: []shiftCheck{
				{day: "3.7.2026", start: "04:30", end: "14:30"}, // realized, not planned 04:40
				{day: "6.7.2026", start: "08:00", end: "16:00", callout: boolPtr(true)},
				{day: "11.7.2026", start: "21:48", end: "05:00", callout: boolPtr(true)}, // overnight merged
			},
		},
		{
			file: "henkilokohtainen-3.pdf",
			from: "29.6.2026", to: "19.7.2026", period: "9/2026",
			shifts: 11, noZeroDuration: true,
			checks: []shiftCheck{
				// Suunniteltu-jakso: 04:40 (ei erillistä toteutunutta 04:30)
				{day: "3.7.2026", start: "04:40", end: "14:30"},
				{day: "29.6.2026", start: "04:55", end: "16:55"},
				{day: "19.7.2026", start: "05:45", end: "15:00"},
			},
		},
		{
			file: "henkilokohtainen-4.pdf",
			from: "1.7.2026", to: "15.7.2026", period: "13/2026",
			shifts: 7, noZeroDuration: true, atLeastOneCallout: true,
			checks: []shiftCheck{
				{day: "3.7.2026", start: "04:30", end: "14:30"},
				{day: "11.7.2026", start: "21:48", end: "05:00", callout: boolPtr(true)},
			},
		},
		{
			file: "henkilokohtainen-5.pdf",
			from: "16.6.2026", to: "30.6.2026", period: "12/2026",
			shifts: 11, noZeroDuration: true,
			checks: []shiftCheck{
				{day: "16.6.2026", start: "04:55", end: "14:30"},
				{day: "25.6.2026", start: "13:00", end: "00:15"},
				{day: "29.6.2026", start: "04:55", end: "16:55"}, // vs -1 which stops at 28.
				{day: "30.6.2026", start: "04:55", end: "14:30"},
			},
		},
		{
			file: "henkilokohtainen-6.pdf",
			from: "16.7.2026", to: "31.7.2026", period: "14/2026",
			shifts: 11, noZeroDuration: true, atLeastOneCallout: true,
			checks: []shiftCheck{
				{day: "16.7.2026", start: "04:55", end: "16:15"},
				{day: "28.7.2026", start: "13:15", end: "00:20", callout: boolPtr(true)},
				{day: "31.7.2026", start: "04:55", end: "14:30"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			path := filepath.Join("testdata", tc.file)
			if _, err := os.Stat(path); err != nil {
				t.Skipf("missing %s (gitignored local sample): %v", path, err)
			}
			res, err := ParseFile(path)
			if err != nil {
				t.Fatal(err)
			}
			assertDate(t, "from", res.From, tc.from)
			assertDate(t, "to", res.To, tc.to)
			if res.Period != tc.period {
				t.Fatalf("period=%q want %q", res.Period, tc.period)
			}
			if len(res.Shifts) != tc.shifts {
				for _, s := range res.Shifts {
					t.Logf("  %s %s-%s callout=%v code=%s",
						s.Date.Format("2.1.2006"), s.Start, s.End, s.Callout, s.Code)
				}
				t.Fatalf("shifts=%d want %d", len(res.Shifts), tc.shifts)
			}
			if tc.noZeroDuration {
				for _, s := range res.Shifts {
					if s.Start == s.End {
						t.Fatalf("zero-duration shift leaked: %+v", s)
					}
				}
			}
			if tc.atLeastOneCallout {
				ok := false
				for _, s := range res.Shifts {
					if s.Callout {
						ok = true
						break
					}
				}
				if !ok {
					t.Fatal("expected at least one Callout shift (Hälytys aika)")
				}
			}
			byDay := indexShifts(res.Shifts)
			for _, c := range tc.checks {
				got, ok := byDay[c.day+"|"+c.start]
				if !ok {
					t.Fatalf("missing shift %s %s-%s", c.day, c.start, c.end)
				}
				if got.End != c.end {
					t.Fatalf("%s: end=%s want %s", c.day, got.End, c.end)
				}
				if c.callout != nil && got.Callout != *c.callout {
					t.Fatalf("%s %s: callout=%v want %v", c.day, c.start, got.Callout, *c.callout)
				}
			}
			// Shifts stay inside report window (±1 day for overnight end next morning).
			lo := res.From.AddDate(0, 0, -1)
			hi := res.To.AddDate(0, 0, 1)
			for _, s := range res.Shifts {
				if s.Date.Before(lo) || s.Date.After(hi) {
					t.Fatalf("shift date %s outside window %s-%s",
						s.Date.Format("2.1.2006"), tc.from, tc.to)
				}
			}
		})
	}
}

func assertDate(t *testing.T, label string, got time.Time, wantFI string) {
	t.Helper()
	want, err := time.ParseInLocation("2.1.2006", wantFI, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatalf("%s=%s want %s", label, got.Format("2.1.2006"), wantFI)
	}
}

func indexShifts(shifts []Shift) map[string]Shift {
	out := make(map[string]Shift, len(shifts))
	for _, s := range shifts {
		key := s.Date.Format("2.1.2006") + "|" + s.Start
		out[key] = s
	}
	return out
}
