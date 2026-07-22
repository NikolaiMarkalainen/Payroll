package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestOptionSelectMenuExcludesSelected(t *testing.T) {
	test.NewApp()
	o := newOptionSelect([]string{"A", "B", "C", "D"}, nil)
	o.SetSelected("A")

	var listed []string
	for _, opt := range o.Options {
		if opt == o.Selected {
			continue
		}
		listed = append(listed, opt)
	}
	if len(listed) != 3 || listed[0] != "B" || listed[2] != "D" {
		t.Fatalf("menu should be B,C,D got %v", listed)
	}
	if containsString(listed, "A") {
		t.Fatal("selected A must not appear in menu options")
	}
}
