package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"payroll/internal/pdfimport"
)

type pdfTab struct {
	window    fyne.Window
	shifts    *shiftsTab
	calc      *calcTab
	pathLbl   *widget.Label
	status    *widget.Label
	preview   *widget.Label
	importBtn *widget.Button
	clearBtn  *widget.Button
	result    *pdfimport.Result
	lastDir   string
}

type pdfFileRow struct {
	path string
	name string
	mod  time.Time
}

func newPDFTab(w fyne.Window, shifts *shiftsTab, calc *calcTab) *pdfTab {
	p := &pdfTab{
		window:  w,
		shifts:  shifts,
		calc:    calc,
		pathLbl: widget.NewLabel("Ei tiedostoa valittuna."),
		status:  widget.NewLabel(""),
		preview: widget.NewLabel("Valitse TyövuoroVelho-PDF (Toteutuneet tai Suunnitellut työvuorot)."),
		lastDir: defaultPDFDir(),
	}
	p.pathLbl.Wrapping = fyne.TextWrapWord
	p.status.Wrapping = fyne.TextWrapWord
	p.preview.Wrapping = fyne.TextWrapWord
	p.importBtn = widget.NewButton("Tuo vuoroihin", func() {
		p.applyImport()
	})
	p.importBtn.Importance = widget.HighImportance
	p.importBtn.Disable()
	p.clearBtn = widget.NewButton("Tyhjennä tuonti", func() {
		p.clearImport()
	})
	p.clearBtn.Disable()
	return p
}

func defaultPDFDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	dl := filepath.Join(home, "Downloads")
	if st, err := os.Stat(dl); err == nil && st.IsDir() {
		return dl
	}
	return home
}

func (p *pdfTab) canvas() fyne.CanvasObject {
	heading := widget.NewLabel("PDF-tuonti")
	heading.TextStyle = fyne.TextStyle{Bold: true}

	hint := widget.NewLabel("Tuo TyövuoroVelho-PDF (toteutuneet tai suunnitellut). Useita PDF:itä kerralla; saman päivän erimielisyydessä kysytään kumpi pidetään. Täydempi jakso voittaa automaattisesti. Tyhjennä tuonti nollaa esikatselun jos jokin menee pieleen.")
	hint.Wrapping = fyne.TextWrapWord

	openBtn := widget.NewButton("Valitse PDF...", func() {
		p.openFile()
	})

	previewBox := container.NewVScroll(p.preview)
	previewBox.SetMinSize(fyne.NewSize(0, 280))

	return container.NewPadded(container.NewBorder(
		container.NewVBox(
			heading,
			hint,
			widget.NewSeparator(),
			container.NewHBox(openBtn, p.clearBtn),
			p.pathLbl,
			p.status,
			widget.NewSeparator(),
			widget.NewLabel("Esikatselu"),
			p.importBtn,
		),
		nil, nil, nil,
		previewBox,
	))
}

func (p *pdfTab) clearImport() {
	p.result = nil
	p.pathLbl.SetText("Ei tiedostoa valittuna.")
	p.preview.SetText("Tuonti tyhjennetty.")
	p.status.SetText("Tuonti tyhjennetty. Voit valita PDF:t uudelleen.")
	p.importBtn.Disable()
	p.clearBtn.Disable()
}

func (p *pdfTab) openFile() {
	if p.window == nil {
		return
	}
	dir := p.lastDir
	if dir == "" {
		dir = defaultPDFDir()
	}

	dirLbl := widget.NewLabel(dir)
	dirLbl.Wrapping = fyne.TextWrapWord
	hint := widget.NewLabel("PDF:t uusin ensin. Klikkaa rivistä valitaksesi useita (multiselect).")
	hint.Wrapping = fyne.TextWrapWord
	selLbl := widget.NewLabel("Valittu: 0")

	var files []pdfFileRow
	selected := map[string]bool{}

	list := widget.NewList(
		func() int { return len(files) },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(files) {
				return
			}
			f := files[id]
			mark := "[ ]"
			if selected[f.path] {
				mark = "[x]"
			}
			obj.(*widget.Label).SetText(fmt.Sprintf("%s  %s    %s", mark, f.name, f.mod.Format("02.01.2006 15:04")))
		},
	)
	updateSel := func() {
		n := 0
		for _, on := range selected {
			if on {
				n++
			}
		}
		selLbl.SetText(fmt.Sprintf("Valittu: %d", n))
		list.Refresh()
	}
	list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(files) {
			return
		}
		path := files[id].path
		selected[path] = !selected[path]
		list.Unselect(id)
		updateSel()
	}

	reload := func(path string) {
		dir = path
		p.lastDir = path
		dirLbl.SetText(path)
		files = listPDFFilesNewestFirst(path)
		selected = map[string]bool{}
		list.UnselectAll()
		updateSel()
	}
	reload(dir)

	listScroll := container.NewVScroll(list)
	listScroll.SetMinSize(fyne.NewSize(520, 320))

	var d dialog.Dialog
	pickFolder := widget.NewButton("Muu kansio...", func() {
		fd := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				p.status.SetText("Kansio: " + err.Error())
				return
			}
			if uri == nil {
				return
			}
			reload(uri.Path())
			d.Show()
		}, p.window)
		fd.Show()
	})
	selectAll := widget.NewButton("Valitse kaikki", func() {
		for _, f := range files {
			selected[f.path] = true
		}
		updateSel()
	})
	clearSel := widget.NewButton("Tyhjennä valinta", func() {
		selected = map[string]bool{}
		updateSel()
	})

	body := container.NewBorder(
		container.NewVBox(hint, dirLbl, container.NewHBox(pickFolder, selectAll, clearSel), selLbl),
		nil, nil, nil,
		listScroll,
	)

	d = dialog.NewCustomConfirm(
		"Valitse PDF",
		"Avaa",
		"Peruuta",
		body,
		func(ok bool) {
			if !ok {
				return
			}
			paths := selectedPaths(files, selected)
			if len(paths) == 0 {
				p.status.SetText("Valitse vähintään yksi PDF.")
				return
			}
			p.loadPDFPaths(paths)
		},
		p.window,
	)
	d.Show()
}

func selectedPaths(files []pdfFileRow, selected map[string]bool) []string {
	var out []string
	for _, f := range files {
		if selected[f.path] {
			out = append(out, f.path)
		}
	}
	return out
}

func listPDFFilesNewestFirst(dir string) []pdfFileRow {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []pdfFileRow
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.ToLower(filepath.Ext(name)) != ".pdf" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, pdfFileRow{
			path: filepath.Join(dir, name),
			name: name,
			mod:  info.ModTime(),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].mod.Equal(out[j].mod) {
			return strings.ToLower(out[i].name) < strings.ToLower(out[j].name)
		}
		return out[i].mod.After(out[j].mod)
	})
	return out
}

func (p *pdfTab) loadPDFPaths(paths []string) {
	var results []*pdfimport.Result
	var labels []string
	var names []string
	var errs []string
	for _, path := range paths {
		res, err := pdfimport.ParseFile(path)
		base := filepath.Base(path)
		if err != nil {
			errs = append(errs, base+": "+err.Error())
			continue
		}
		results = append(results, res)
		labels = append(labels, base)
		names = append(names, base)
	}
	if len(results) == 0 {
		p.result = nil
		p.importBtn.Disable()
		p.clearBtn.Enable()
		p.status.SetText("Parsinta epäonnistui: " + strings.Join(errs, "; "))
		p.preview.SetText("")
		p.pathLbl.SetText(strings.Join(pathsBase(paths), ", "))
		return
	}
	merged, conflicts := pdfimport.MergeResultsWithConflicts(labels, results...)
	if len(errs) > 0 {
		merged.Warnings = append(merged.Warnings, errs...)
	}
	p.pathLbl.SetText(strings.Join(names, ", "))
	if len(paths) > 0 {
		p.lastDir = filepath.Dir(paths[0])
	}
	p.clearBtn.Enable()

	if len(conflicts) == 0 {
		p.finishLoad(merged)
		return
	}
	p.resolveConflicts(merged, conflicts, 0, map[string][]pdfimport.Shift{})
}

func pathsBase(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = filepath.Base(p)
	}
	return out
}

func (p *pdfTab) finishLoad(merged *pdfimport.Result) {
	p.result = merged
	p.showPreview(merged)
	p.importBtn.Enable()
	p.clearBtn.Enable()
}

func (p *pdfTab) resolveConflicts(base *pdfimport.Result, conflicts []pdfimport.DayConflict, idx int, choices map[string][]pdfimport.Shift) {
	if idx >= len(conflicts) {
		p.finishLoad(pdfimport.ApplyConflictChoices(base, choices))
		return
	}
	c := conflicts[idx]
	labels := make([]string, 0, len(c.Options))
	for _, o := range c.Options {
		labels = append(labels, o.Label)
	}
	sel := widget.NewRadioGroup(labels, nil)
	if len(labels) > 0 {
		sel.SetSelected(labels[0])
	}
	msg := widget.NewLabel(fmt.Sprintf(
		"Päivä %s: saman mittainen mutta eri sisältö useassa PDF:ssä.\nValitse kumpi pidetään:",
		formatFIDate(c.Day),
	))
	msg.Wrapping = fyne.TextWrapWord
	body := container.NewVBox(msg, sel)

	dialog.NewCustomConfirm(
		"Törmäys: valitse vuoro",
		"Käytä tätä",
		"Peruuta tuonti",
		body,
		func(ok bool) {
			if !ok {
				p.status.SetText("Tuonti peruttu törmäyksen kohdalla. Voit tyhjentää ja yrittää uudelleen.")
				p.clearBtn.Enable()
				return
			}
			chosen := sel.Selected
			for _, o := range c.Options {
				if o.Label == chosen {
					choices[c.DayKey] = o.Shifts
					break
				}
			}
			p.resolveConflicts(base, conflicts, idx+1, choices)
		},
		p.window,
	).Show()
}

func (p *pdfTab) showPreview(res *pdfimport.Result) {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", res.Person)
	fmt.Fprintf(&b, "Jakso: %s - %s", formatFIDate(res.From), formatFIDate(res.To))
	if res.Period != "" {
		fmt.Fprintf(&b, " (%s)", res.Period)
	}
	fmt.Fprintf(&b, "\nVuoroja: %d\n\n", len(res.Shifts))
	for _, s := range res.Shifts {
		tag := ""
		if s.Callout {
			tag = " [hälytys]"
		}
		code := ""
		if s.Code != "" {
			code = " *" + s.Code
		}
		fmt.Fprintf(&b, "%s  %s-%s%s%s\n", formatFIDate(s.Date), s.Start, s.End, code, tag)
	}
	if len(res.Warnings) > 0 {
		b.WriteString("\nHuom:\n")
		for _, w := range res.Warnings {
			fmt.Fprintf(&b, "- %s\n", w)
		}
	}
	p.preview.SetText(b.String())
	p.status.SetText(fmt.Sprintf("Valmis tuontiin: %d vuoroa.", len(res.Shifts)))
}

func (p *pdfTab) applyImport() {
	if p.result == nil || p.shifts == nil {
		return
	}
	loc := time.Local
	out := make([]calendarShift, 0, len(p.result.Shifts))
	for _, s := range p.result.Shifts {
		d := s.Date.In(loc)
		out = append(out, calendarShift{
			Date:    time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc),
			Start:   s.Start,
			End:     s.End,
			Callout: s.Callout,
			Code:    s.Code,
		})
	}
	focus := p.result.From
	if len(out) > 0 {
		focus = out[0].Date
	}
	p.shifts.replaceShifts(out, focus)
	if p.calc != nil {
		if _, ok := p.calc.anchorDate(); !ok && !focus.IsZero() {
			// First import: use first shift day as 1. jakson ankkuri if unset.
			p.calc.periodAnchor.SetText(formatFIDate(focus))
			p.calc.syncAnchorToSettings()
		}
		p.calc.refreshPeriodOptions()
		if _, ok := p.calc.anchorDate(); ok {
			p.calc.selectPeriodCoveringShifts()
		} else if !p.result.From.IsZero() && !p.result.To.IsZero() {
			p.calc.setRange(p.result.From, p.result.To)
		}
	}
	p.status.SetText(fmt.Sprintf("Tuotu %d vuoroa kalenteriin. Katso Vuorot-välilehti.", len(out)))
}
