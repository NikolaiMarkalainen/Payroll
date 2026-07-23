package pdfimport

import (
	"path/filepath"
	"testing"
	"time"
)

func TestMergeResultsPrefersFullerDay(t *testing.T) {
	d := time.Date(2026, 7, 6, 0, 0, 0, 0, time.Local)
	thin := &Result{
		From: d, To: d,
		Shifts: []Shift{{Date: d, Start: "08:00", End: "12:00"}}, // 4h
	}
	full := &Result{
		From: d, To: d,
		Shifts: []Shift{{Date: d, Start: "08:00", End: "16:00", Callout: true}}, // 8h
	}
	got := MergeResults(thin, full)
	if len(got.Shifts) != 1 {
		t.Fatalf("shifts=%d", len(got.Shifts))
	}
	if got.Shifts[0].End != "16:00" || !got.Shifts[0].Callout {
		t.Fatalf("want fuller day: %+v", got.Shifts[0])
	}
}

func TestMergeResultsUnionsDifferentDays(t *testing.T) {
	d1 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	d2 := time.Date(2026, 7, 15, 0, 0, 0, 0, time.Local)
	a := &Result{
		From: d1, To: d1,
		Shifts: []Shift{{Date: d1, Start: "06:00", End: "14:00"}},
	}
	b := &Result{
		From: d2, To: d2,
		Shifts: []Shift{{Date: d2, Start: "08:00", End: "16:00"}},
	}
	got := MergeResults(a, b)
	if len(got.Shifts) != 2 {
		t.Fatalf("shifts=%d", len(got.Shifts))
	}
	if !got.From.Equal(d1) || !got.To.Equal(d2) {
		t.Fatalf("range %v-%v", got.From, got.To)
	}
}

func TestMergeResultsKeepsWholeDayFromWinner(t *testing.T) {
	d := time.Date(2026, 7, 11, 0, 0, 0, 0, time.Local)
	a := &Result{Shifts: []Shift{
		{Date: d, Start: "21:48", End: "05:00", Callout: true},
	}}
	b := &Result{Shifts: []Shift{
		{Date: d, Start: "08:00", End: "10:00"},
		{Date: d, Start: "12:00", End: "14:00"},
	}}
	got := MergeResults(b, a)
	if len(got.Shifts) != 1 || got.Shifts[0].Start != "21:48" {
		t.Fatalf("%+v", got.Shifts)
	}
}

func TestMergeEqualFullnessAsksConflict(t *testing.T) {
	d := time.Date(2026, 6, 23, 0, 0, 0, 0, time.Local)
	a := &Result{Shifts: []Shift{{Date: d, Start: "14:30", End: "00:15", Callout: true}}}
	b := &Result{Shifts: []Shift{{Date: d, Start: "14:30", End: "24:00"}}}
	// overnight ~9h45 vs 9h30 — not equal; make equal length
	a = &Result{Shifts: []Shift{{Date: d, Start: "08:00", End: "16:00"}}}
	b = &Result{Shifts: []Shift{{Date: d, Start: "09:00", End: "17:00"}}}
	_, conflicts := MergeResultsWithConflicts([]string{"a.pdf", "b.pdf"}, a, b)
	if len(conflicts) != 1 {
		t.Fatalf("conflicts=%d want 1", len(conflicts))
	}
	if len(conflicts[0].Options) != 2 {
		t.Fatalf("options=%d", len(conflicts[0].Options))
	}
}

func TestMergeIdenticalDaysNoConflict(t *testing.T) {
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	a := &Result{Shifts: []Shift{{Date: d, Start: "04:55", End: "14:30"}}}
	b := &Result{Shifts: []Shift{{Date: d, Start: "04:55", End: "14:30"}}}
	got, conflicts := MergeResultsWithConflicts(nil, a, b)
	if len(conflicts) != 0 {
		t.Fatalf("conflicts=%v", conflicts)
	}
	if len(got.Shifts) != 1 {
		t.Fatalf("shifts=%d", len(got.Shifts))
	}
}

func TestParseSuunnitellutPDF(t *testing.T) {
	path := filepath.Join("testdata", "henkilokohtainen-3.pdf")
	res, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if res.From.IsZero() || res.To.IsZero() {
		t.Fatal("missing range")
	}
	if len(res.Shifts) == 0 {
		t.Fatal("expected planned shifts")
	}
	// ma 29.6. 04:55-16:55
	found := false
	for _, sh := range res.Shifts {
		if sh.Date.Day() == 29 && sh.Date.Month() == time.June && sh.Start == "04:55" && sh.End == "16:55" {
			found = true
		}
	}
	if !found {
		for _, sh := range res.Shifts {
			t.Logf("%s %s-%s", sh.Date.Format("2.1.2006"), sh.Start, sh.End)
		}
		t.Fatal("missing 29.6. 04:55-16:55")
	}
}

func TestDedupeExactShifts(t *testing.T) {
	d := time.Date(2026, 6, 23, 0, 0, 0, 0, time.Local)
	in := []Shift{
		{Date: d, Start: "14:30", End: "24:00", Code: "A"},
		{Date: d, Start: "14:30", End: "24:00", Code: "A"},
		{Date: d.AddDate(0, 0, 1), Start: "00:00", End: "00:15", Code: "A"},
	}
	out := normalizeShifts(in)
	if len(out) != 1 {
		t.Fatalf("got %d want overnight merge to 1: %+v", len(out), out)
	}
	if out[0].Start != "14:30" || out[0].End != "00:15" {
		t.Fatalf("%+v", out[0])
	}
}
