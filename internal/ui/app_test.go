package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestBuildUIHasExpectedTabs(t *testing.T) {
	test.NewApp()
	w := test.NewWindow(nil)
	defer w.Close()

	content, tabs, _, _, _ := buildUI(w)
	w.SetContent(content)

	want := []string{"Asetukset", "Vuorot", "PDF-tuonti", "Laskelma", "Vertailu"}
	if len(tabs.Items) != len(want) {
		t.Fatalf("tabs=%d want %d", len(tabs.Items), len(want))
	}
	for i, name := range want {
		if tabs.Items[i].Text != name {
			t.Fatalf("tab[%d]=%q want %q", i, tabs.Items[i].Text, name)
		}
	}

	tabs.SelectIndex(1)
	if tabs.Selected().Text != "Vuorot" {
		t.Fatalf("selected=%q", tabs.Selected().Text)
	}
	tabs.SelectIndex(0)
	if tabs.Selected().Text != "Asetukset" {
		t.Fatalf("selected=%q", tabs.Selected().Text)
	}
}
