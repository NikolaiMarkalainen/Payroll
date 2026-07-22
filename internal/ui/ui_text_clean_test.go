package ui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// Fancy / brittle glyphs that often render as boxes or "escape junk" in Fyne UI fonts.
var forbiddenUIRunes = []rune{
	'\uFFFD', // replacement <>
	'\u2014', // em dash
	'\u2013', // en dash
	'\u2026', // ellipsis
	'\u00D7', // multiplication ×
	'\u2715', // ✕
	'\u2192', // →
	'\u00B7', // middle dot
	'\u20AC', // euro
	'\u00A7', // section §
	'\u2265', // ≥ (often missing in UI fonts → box/FFFD lookalike)
	'\u2264', // ≤
	'\u00B1', // ±
}

func TestUIStringsHaveNoBrokenGlyphs(t *testing.T) {
	root := "."
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var bad []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(root, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if bytesContainsFFFD(raw) {
			bad = append(bad, path+": raw UTF-8 replacement bytes (EF BF BD)")
		}
		file, err := parser.ParseFile(fset, path, raw, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			bl, ok := n.(*ast.BasicLit)
			if !ok || bl.Kind != token.STRING {
				return true
			}
			s, err := strconv.Unquote(bl.Value)
			if err != nil {
				return true
			}
			if issue := uiTextIssue(s); issue != "" {
				bad = append(bad, path+": "+issue+" in "+truncateUI(s, 60))
			}
			return true
		})
	}
	if len(bad) > 0 {
		t.Fatalf("UI string glyph/escape issues:\n%s", strings.Join(bad, "\n"))
	}
}

func TestLiveSettingsUITextClean(t *testing.T) {
	test.NewApp()
	s := newSettingsTab()
	_ = s.canvas()

	// Cover each TES family so all dynamic hints/labels are filled.
	for _, family := range []string{tesFamilyCustom, tesFamilyVartio, tesFamilyKauppa} {
		s.tesFamily.SetSelected(family)
		for _, text := range collectSettingsUITexts(s) {
			if issue := uiTextIssue(text); issue != "" {
				t.Fatalf("family %q: %s in %s", family, issue, truncateUI(text, 80))
			}
		}
	}

	// Explicit overtime labels (screenshot regression: "Vuoro >? h ->? 50 %").
	s.applyDemoTES()
	if s.overtime50Item == nil || s.overtime100Item == nil {
		t.Fatal("overtime form items missing")
	}
	ot50, ot100 := s.overtime50Item.Text, s.overtime100Item.Text
	for _, label := range []string{ot50, ot100, s.otHint.Text} {
		if issue := uiTextIssue(label); issue != "" {
			t.Fatalf("%s in %s", issue, truncateUI(label, 80))
		}
	}
	if strings.Contains(ot50, "> h") || strings.ContainsRune(ot50, '\uFFFD') {
		t.Fatalf("broken overtime 50 label: %q", ot50)
	}
	if !strings.Contains(ot50, "50") || !strings.Contains(ot100, "100") {
		t.Fatalf("unexpected OT labels: %q / %q", ot50, ot100)
	}
}

func bytesContainsFFFD(b []byte) bool {
	for i := 0; i+2 < len(b); i++ {
		if b[i] == 0xEF && b[i+1] == 0xBF && b[i+2] == 0xBD {
			return true
		}
	}
	return false
}

func uiTextIssue(s string) string {
	if !utf8.ValidString(s) {
		return "invalid UTF-8"
	}
	if strings.ContainsRune(s, '\uFFFD') {
		return "replacement char U+FFFD"
	}
	for _, r := range forbiddenUIRunes {
		if strings.ContainsRune(s, r) {
			return "forbidden rune U+" + strconv.FormatInt(int64(r), 16)
		}
	}
	for _, esc := range []string{`\u00`, `\x{`, `&auml;`, `&ouml;`, `&euro;`, `&#`} {
		if strings.Contains(s, esc) {
			return "escape remnant " + esc
		}
	}
	return ""
}

func collectSettingsUITexts(s *settingsTab) []string {
	out := []string{}
	add := func(v string) {
		if strings.TrimSpace(v) != "" {
			out = append(out, v)
		}
	}
	if s.tesFamily != nil {
		out = append(out, s.tesFamily.Options...)
		add(s.tesFamily.Selected)
	}
	if s.tesLevel != nil {
		out = append(out, s.tesLevel.Options...)
		add(s.tesLevel.Selected)
	}
	if s.tesRegion != nil {
		out = append(out, s.tesRegion.Options...)
	}
	if s.experienceTier != nil {
		out = append(out, s.experienceTier.Options...)
	}
	for _, lbl := range []*widget.Label{
		s.profileHint, s.expHint, s.otHint, s.personalHeading,
		s.periodHeading, s.periodHint, s.status,
	} {
		if lbl != nil {
			add(lbl.Text)
		}
	}
	if s.levelFormItem != nil {
		add(s.levelFormItem.Text)
	}
	if s.overtime50Item != nil {
		add(s.overtime50Item.Text)
	}
	if s.overtime100Item != nil {
		add(s.overtime100Item.Text)
	}
	for _, sec := range []fyne.CanvasObject{
		s.profileSelectorsSection, s.trainingSection, s.eveningDoubleSection,
		s.shiftOTSection, s.weeklyOTSection, s.kaupanFlagsSection,
		s.periodOTSection, s.periodAdvancedSection,
	} {
		out = append(out, formItemTexts(sec)...)
	}
	// Checks / selects with user-visible text.
	for _, c := range []*widget.Check{
		s.trainingEnabled, s.shiftOTEnabled, s.weeklyOTEnabled,
		s.eveningExcludeSaturday, s.nightExcludeSunday, s.nightExcludeHoliday,
		s.periodOTEnabled,
	} {
		if c != nil {
			add(c.Text)
		}
	}
	return out
}

func formItemTexts(o fyne.CanvasObject) []string {
	var out []string
	switch v := o.(type) {
	case *widget.Form:
		for _, it := range v.Items {
			if it != nil {
				out = append(out, it.Text)
			}
		}
	case *fyne.Container:
		for _, child := range v.Objects {
			out = append(out, formItemTexts(child)...)
		}
	}
	return out
}

func truncateUI(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	if len(s) <= n {
		return strconv.Quote(s)
	}
	return strconv.Quote(s[:n] + "...")
}
