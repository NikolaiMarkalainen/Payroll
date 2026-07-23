package pdfimport

import (
	"path/filepath"
	"testing"
	"time"
)

func sampleHeaderTokens() []string {
	return []string{
		"TP Testi Testaaja",
		"Toteutuneet työvuorot: 01.07.2026 - 15.07.2026 (13/2026)",
		"Suunnitellut",
	}
}

func TestParseVelhoPlannedVsRealized(t *testing.T) {
	tokens := append(sampleHeaderTokens(),
		"pe 3.7.",
		"04:40-14:30*1LOC", // planned
		"04:30-14:30*1LOC", // realized
		"Yhteensä",
	)
	res, err := ParseVelho(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Shifts) != 1 {
		t.Fatalf("shifts=%d %+v", len(res.Shifts), res.Shifts)
	}
	s := res.Shifts[0]
	if s.Start != "04:30" || s.End != "14:30" {
		t.Fatalf("got %s-%s want realized 04:30-14:30", s.Start, s.End)
	}
}

func TestParseVelhoSkipsZeroTH(t *testing.T) {
	tokens := append(sampleHeaderTokens(),
		"ma 6.7.",
		"08:00-16:00*1LOC",
		"00:00-00:00*TH",
		"Yhteensä",
	)
	res, err := ParseVelho(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Shifts) != 1 {
		t.Fatalf("shifts=%d want 1 (TH skipped)", len(res.Shifts))
	}
	if res.Shifts[0].Code == "TH" {
		t.Fatal("TH should not become a shift")
	}
}

func TestParseVelhoMergesOvernightHalves(t *testing.T) {
	tokens := append(sampleHeaderTokens(),
		"la 11.7.",
		"@21:48-24:00*EXTRA",
		"su 12.7.",
		"@00:00-05:00*EXTRA",
		"Yhteensä",
	)
	res, err := ParseVelho(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Shifts) != 1 {
		t.Fatalf("shifts=%d want 1 merged overnight", len(res.Shifts))
	}
	s := res.Shifts[0]
	if s.Start != "21:48" || s.End != "05:00" || !s.Callout {
		t.Fatalf("got %s-%s callout=%v", s.Start, s.End, s.Callout)
	}
	if s.Date.Day() != 11 {
		t.Fatalf("date day=%d want 11", s.Date.Day())
	}
}

func TestParseVelhoMissingHeader(t *testing.T) {
	_, err := ParseVelho([]string{"Suunnitellut", "ma 1.7.", "08:00-16:00*1LOC"})
	if err == nil {
		t.Fatal("expected header error")
	}
}

func TestParseVelhoSkipsHalytysSection(t *testing.T) {
	tokens := append(sampleHeaderTokens(),
		"ma 1.7.",
		"04:55-14:30*1LOC",
		"Yhteensä",
		"Hälytys aika",
		"ti 2.7.",
		"08:00-09:00*EXTRA", // must not be imported as a new shift
		"Vuosivapaatilasto",
	)
	res, err := ParseVelho(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Shifts) != 1 {
		t.Fatalf("shifts=%d want only pre-Hälytys shift", len(res.Shifts))
	}
}

func TestParseVelhoHalytysTagsCalloutCheckbox(t *testing.T) {
	tokens := append(sampleHeaderTokens(),
		"ma 6.7.",
		"08:00-16:00*BETA1X",
		"ti 7.7.",
		"08:00-16:00*1LOC",
		"Yhteensä",
		"Hälytys aika",
		"ma 6.7.",
		"2:00",
		"ti 7.7.",
		"0:00",
		"Yhteensä",
	)
	res, err := ParseVelho(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Shifts) != 2 {
		t.Fatalf("shifts=%d want 2", len(res.Shifts))
	}
	if !res.Shifts[0].Callout {
		t.Fatal("6.7. should be Callout (Hälytysvuoro) from Hälytys aika")
	}
	if res.Shifts[1].Callout {
		t.Fatal("7.7. with 0:00 hälytys should not be Callout")
	}
}

func TestPickRealizedOnlyActual(t *testing.T) {
	got := pickRealized([]rawRange{{start: "08:00", end: "16:00", code: "1LOC"}})
	if len(got) != 1 || got[0].start != "08:00" {
		t.Fatalf("%+v", got)
	}
}

func TestPickRealizedMultipleShiftsPerDay(t *testing.T) {
	// Column-wise order from PDF: all planned, then all realized.
	ranges := []rawRange{
		{start: "00:00", end: "00:15", code: "2BBB"},
		{start: "14:30", end: "24:00", code: "2BBB"},
		{start: "00:00", end: "00:15", code: "2BBB"},
		{start: "14:30", end: "24:00", code: "2BBB"},
	}
	got := pickRealized(ranges)
	if len(got) != 2 {
		t.Fatalf("got %d want 2: %+v", len(got), got)
	}
	if got[0].start != "00:00" || got[0].end != "00:15" {
		t.Fatalf("first=%+v", got[0])
	}
	if got[1].start != "14:30" || got[1].end != "24:00" {
		t.Fatalf("second=%+v", got[1])
	}
}

func TestParseMultiShiftDayPDF(t *testing.T) {
	path := filepath.Join("testdata", "henkilokohtainen-5.pdf")
	res, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// After overnight merge: 22.6 16:55-00:15, 23.6 14:30-00:15, 24.6 14:30-00:15, 25.6 13:00-00:15
	byDay := map[string][]Shift{}
	for _, sh := range res.Shifts {
		k := sh.Date.Format("2.1.2006")
		byDay[k] = append(byDay[k], sh)
	}
	// 23.6 should be one overnight 14:30-00:15 (00:00-00:15 merged from previous evening start on 22)
	d23 := byDay["23.6.2026"]
	if len(d23) != 1 {
		t.Fatalf("23.6 shifts=%d %+v want 1 overnight", len(d23), d23)
	}
	if d23[0].Start != "14:30" || d23[0].End != "00:15" {
		t.Fatalf("23.6=%+v want 14:30-00:15", d23[0])
	}
	d22 := byDay["22.6.2026"]
	if len(d22) != 1 || d22[0].Start != "16:55" || d22[0].End != "00:15" {
		t.Fatalf("22.6=%+v want 16:55-00:15", d22)
	}
	// No orphan duplicate 14:30-24:00 left on 23/24/25
	for _, day := range []string{"23.6.2026", "24.6.2026", "25.6.2026"} {
		for _, sh := range byDay[day] {
			if sh.End == "24:00" {
				t.Fatalf("%s still has unmerged half: %+v", day, sh)
			}
		}
	}
}

func TestMergeOvernightRequiresContiguousCode(t *testing.T) {
	d1 := time.Date(2026, 7, 11, 0, 0, 0, 0, time.Local)
	d2 := d1.AddDate(0, 0, 1)
	in := []Shift{
		{Date: d1, Start: "21:00", End: "24:00", Code: "A"},
		{Date: d2, Start: "00:00", End: "05:00", Code: "B"}, // different code — still merge
	}
	out := mergeOvernightHalves(in)
	if len(out) != 1 || out[0].End != "05:00" {
		t.Fatalf("%+v", out)
	}
}

func TestNormalizeClock(t *testing.T) {
	if normalizeClock("4:5") != "04:05" {
		t.Fatalf("%q", normalizeClock("4:5"))
	}
	if normalizeClock("04:30") != "04:30" {
		t.Fatalf("%q", normalizeClock("04:30"))
	}
}
