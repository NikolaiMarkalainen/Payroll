package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	paycalc "payroll/internal/calc"
	"payroll/internal/pdfimport"
)

func TestPDFTabInAppTabs(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()
	_, tabs, _, _, _ := buildUI(w)
	found := false
	for _, it := range tabs.Items {
		if it.Text == "PDF-tuonti" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("PDF-tuonti tab missing from app")
	}
}

func TestPDFTabPreviewAndImportSample(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	settings := newSettingsTab()
	_ = settings.canvas()
	shifts := newShiftsTab(w)
	_ = shifts.canvas()
	calc := newCalcTab(settings, shifts)
	_ = calc.canvas()

	p := newPDFTab(w, shifts, calc)
	_ = p.canvas()

	path := filepath.Join("..", "pdfimport", "testdata", "henkilokohtainen-2.pdf")
	res, err := pdfimport.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p.result = res
	p.showPreview(res)
	p.importBtn.Enable()

	if issue := uiTextIssue(p.preview.Text); issue != "" {
		t.Fatalf("preview %s: %s", issue, truncateUI(p.preview.Text, 80))
	}
	if issue := uiTextIssue(p.status.Text); issue != "" {
		t.Fatalf("status %s", issue)
	}
	for _, needle := range []string{"Vuoroja: 7", "04:30", "21:48", "05:00", "13/2026"} {
		if !strings.Contains(p.preview.Text, needle) {
			t.Fatalf("preview missing %q:\n%s", needle, p.preview.Text)
		}
	}
	if strings.Contains(p.preview.Text, "04:40") {
		t.Fatal("preview must use realized 04:30, not planned 04:40")
	}

	p.applyImport()
	if len(shifts.shifts) != 7 {
		t.Fatalf("imported shifts=%d want 7", len(shifts.shifts))
	}
	// Realized pe 3.7. 04:30 + shift codes from PDF
	foundRealized := false
	foundCoded := false
	for _, sh := range shifts.shifts {
		if sh.Date.Day() == 3 && sh.Start == "04:30" && sh.End == "14:30" {
			foundRealized = true
			if sh.Code == "" {
				t.Fatal("3.7. missing place code after import")
			}
		}
		if sh.Date.Day() == 15 && sh.Code != "" {
			foundCoded = true
		}
		if sh.Start == "00:00" && sh.End == "00:00" {
			t.Fatalf("TH/zero shift leaked into calendar: %+v", sh)
		}
	}
	if !foundRealized {
		t.Fatal("realized 3.7. 04:30-14:30 missing after import")
	}
	if !foundCoded {
		t.Fatal("15.7. place code missing after import")
	}

	// First import sets ankkuri from first shift day and picks covering 3 vk jakso.
	anchor, err := parseFIDate(calc.periodAnchor.Text)
	if err != nil {
		t.Fatalf("period anchor: %v (text=%q)", err, calc.periodAnchor.Text)
	}
	wantFrom, wantTo := paycalc.PeriodWindow(anchor, res.From)
	if calc.from.Text != formatFIDate(wantFrom) || calc.to.Text != formatFIDate(wantTo) {
		t.Fatalf("calc range %s-%s want %s-%s (anchor %s)", calc.from.Text, calc.to.Text,
			formatFIDate(wantFrom), formatFIDate(wantTo), formatFIDate(anchor))
	}
	if !strings.Contains(p.status.Text, "Tuotu 7") {
		t.Fatalf("status=%q", p.status.Text)
	}
	if shifts.month.Month() != time.July || shifts.month.Year() != 2026 {
		t.Fatalf("calendar month=%v", shifts.month)
	}
}

func TestPDFTabImportAllSamples(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	settings := newSettingsTab()
	_ = settings.canvas()
	shifts := newShiftsTab(w)
	_ = shifts.canvas()
	calc := newCalcTab(settings, shifts)
	_ = calc.canvas()
	p := newPDFTab(w, shifts, calc)
	_ = p.canvas()

	samples := []struct {
		file   string
		shifts int
	}{
		{"henkilokohtainen.pdf", 8},
		{"henkilokohtainen-1.pdf", 9},
		{"henkilokohtainen-2.pdf", 7},
		{"henkilokohtainen-3.pdf", 11},
		{"henkilokohtainen-4.pdf", 7},
		{"henkilokohtainen-5.pdf", 11},
		{"henkilokohtainen-6.pdf", 11},
	}
	for _, tc := range samples {
		t.Run(tc.file, func(t *testing.T) {
			path := filepath.Join("..", "pdfimport", "testdata", tc.file)
			if _, err := os.Stat(path); err != nil {
				t.Skipf("missing %s: %v", path, err)
			}
			res, err := pdfimport.ParseFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Shifts) != tc.shifts {
				t.Fatalf("parse shifts=%d want %d", len(res.Shifts), tc.shifts)
			}
			shifts.replaceShifts(nil, time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local))
			p.result = res
			p.showPreview(res)
			if issue := uiTextIssue(p.preview.Text); issue != "" {
				t.Fatalf("preview %s", issue)
			}
			p.applyImport()
			if len(shifts.shifts) != tc.shifts {
				t.Fatalf("imported=%d want %d", len(shifts.shifts), tc.shifts)
			}
			for _, sh := range shifts.shifts {
				if sh.Start == sh.End {
					t.Fatalf("zero shift in calendar: %+v", sh)
				}
			}
			if calc.from.Text == "" || calc.to.Text == "" {
				t.Fatal("calc range empty after import")
			}
		})
	}
}

func TestPDFImportPreservesExistingShifts(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	settings := newSettingsTab()
	_ = settings.canvas()
	shifts := newShiftsTab(w)
	_ = shifts.canvas()
	calc := newCalcTab(settings, shifts)
	_ = calc.canvas()
	p := newPDFTab(w, shifts, calc)
	_ = p.canvas()

	loc := time.Local
	existing := []calendarShift{
		{Date: time.Date(2026, 5, 10, 0, 0, 0, 0, loc), Start: "08:00", End: "16:00", Code: "KEEP"},
		{Date: time.Date(2026, 5, 11, 0, 0, 0, 0, loc), Start: "12:00", End: "20:00", Code: "KEEP2"},
	}
	shifts.replaceShifts(existing, existing[0].Date)

	path := filepath.Join("..", "pdfimport", "testdata", "henkilokohtainen-2.pdf")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("missing %s: %v", path, err)
	}
	res, err := pdfimport.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p.result = res
	p.applyImport()

	if len(shifts.shifts) != 2+len(res.Shifts) {
		t.Fatalf("calendar shifts=%d want %d (existing+imported)", len(shifts.shifts), 2+len(res.Shifts))
	}
	foundKeep := false
	foundKeep2 := false
	foundImported := false
	for _, sh := range shifts.shifts {
		if sh.Code == "KEEP" && sh.Start == "08:00" {
			foundKeep = true
		}
		if sh.Code == "KEEP2" {
			foundKeep2 = true
		}
		if sh.Date.Month() == time.July && sh.Date.Day() == 3 {
			foundImported = true
		}
	}
	if !foundKeep || !foundKeep2 {
		t.Fatalf("existing May shifts were wiped: %+v", shifts.shifts)
	}
	if !foundImported {
		t.Fatal("imported July shift missing")
	}
	if !strings.Contains(p.status.Text, "säilyivät") {
		t.Fatalf("status should mention preserved shifts: %q", p.status.Text)
	}
}

func TestPDFTabUIStringsClean(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()
	p := newPDFTab(w, nil, nil)
	_ = p.canvas()
	for _, s := range []string{
		p.pathLbl.Text, p.status.Text, p.preview.Text, p.importBtn.Text,
		"PDF-tuonti",
		"Valitse PDF...",
		"Tuo vuoroihin",
	} {
		if issue := uiTextIssue(s); issue != "" {
			t.Fatalf("%s in %q", issue, s)
		}
	}
}
