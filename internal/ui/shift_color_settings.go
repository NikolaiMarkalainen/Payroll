package ui

import (
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type shiftColorGroup struct {
	Key        string
	Codes      []string // distinct codes in calendar for this key
	ManualOnly bool     // added by hand, not (yet) in calendar
}

func (s *settingsTab) initShiftColors() {
	if s.shiftColorOverrides == nil {
		s.shiftColorOverrides = make(map[string]color.NRGBA)
	}
	if s.shiftColorManual == nil {
		s.shiftColorManual = make(map[string]struct{})
	}
}

func (s *settingsTab) colorForShiftCode(code string) color.NRGBA {
	s.initShiftColors()
	return shiftTitleColorFor(s.shiftColorOverrides, code)
}

func (s *settingsTab) colorForKey(key string) color.NRGBA {
	s.initShiftColors()
	if c, ok := s.shiftColorOverrides[key]; ok {
		return c
	}
	return shiftTitleColorFor(nil, key)
}

func (s *settingsTab) setShiftColor(key string, c color.NRGBA) {
	s.initShiftColors()
	s.shiftColorOverrides[key] = c
	s.refreshColorRows()
	if s.onSaved != nil {
		s.onSaved()
	}
}

func (s *settingsTab) addManualColorKey(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	key := shiftColorKey(raw)
	if key == "" {
		key = strings.ToUpper(raw)
	}
	if key == "" {
		return false
	}
	s.initShiftColors()
	s.shiftColorManual[key] = struct{}{}
	if _, ok := s.shiftColorOverrides[key]; !ok {
		s.shiftColorOverrides[key] = shiftTitleColorFor(nil, key)
	}
	s.refreshColorRows()
	if s.onSaved != nil {
		s.onSaved()
	}
	return true
}

func (s *settingsTab) removeManualColorKey(key string) {
	s.initShiftColors()
	delete(s.shiftColorManual, key)
	delete(s.shiftColorOverrides, key)
	s.refreshColorRows()
	if s.onSaved != nil {
		s.onSaved()
	}
}

func (s *settingsTab) buildCalendarSection() fyne.CanvasObject {
	s.initShiftColors()

	heading := widget.NewLabel("Vuorojen värit")
	heading.TextStyle = fyne.TextStyle{Bold: true}
	hint := widget.NewLabel("Ryhmät tulevat kalenterin vuorokoodeista; voit myös lisätä ryhmän käsin. Sama alkunumero jakaa värin. Klikkaa väripalkkia vaihtaaksesi.")
	hint.Wrapping = fyne.TextWrapWord

	s.colorEmptyHint = widget.NewLabel("Ei ryhmiä vielä. Tuo vuoroja kalenteriin tai lisää ryhmä käsin.")
	s.colorEmptyHint.Wrapping = fyne.TextWrapWord

	s.colorRows = container.NewVBox()
	s.refreshColorRows()

	s.colorExtraEntry = widget.NewEntry()
	s.colorExtraEntry.SetPlaceHolder("Numero tai koodi, esim. 3 tai ALPHA")
	addBtn := widget.NewButton("Lisää", func() {
		if s.addManualColorKey(s.colorExtraEntry.Text) {
			s.colorExtraEntry.SetText("")
		}
	})

	return container.NewVBox(
		heading,
		hint,
		widget.NewForm(widget.NewFormItem("", s.colorShiftTitles)),
		s.colorEmptyHint,
		s.colorRows,
		container.NewBorder(nil, nil, nil, addBtn, s.colorExtraEntry),
	)
}

func (s *settingsTab) refreshColorRows() {
	if s.colorRows == nil {
		return
	}
	s.initShiftColors()
	groups := s.allShiftColorGroups()
	if s.colorEmptyHint != nil {
		if len(groups) == 0 {
			s.colorEmptyHint.Show()
		} else {
			s.colorEmptyHint.Hide()
		}
	}
	rows := make([]fyne.CanvasObject, 0, len(groups))
	for _, g := range groups {
		grp := g
		swatch := canvas.NewRectangle(s.colorForKey(grp.Key))
		swatch.CornerRadius = 4
		swatch.SetMinSize(fyne.NewSize(36, 22))
		pick := newTapBox(container.NewPadded(swatch), func() {
			s.openShiftColorPicker(grp.Key, shiftColorGroupLabel(grp))
		})
		lbl := widget.NewLabel(shiftColorGroupLabel(grp))
		right := []fyne.CanvasObject{pick}
		if grp.ManualOnly {
			del := widget.NewButton("Poista", func() {
				s.removeManualColorKey(grp.Key)
			})
			del.Importance = widget.LowImportance
			right = append(right, del)
		}
		rows = append(rows, container.NewBorder(nil, nil, lbl, nil, container.NewHBox(right...)))
	}
	s.colorRows.Objects = rows
	s.colorRows.Refresh()
}

func (s *settingsTab) allShiftColorGroups() []shiftColorGroup {
	cal := s.shiftColorGroupsFromCalendar()
	seen := make(map[string]bool, len(cal))
	for _, g := range cal {
		seen[g.Key] = true
	}
	out := append([]shiftColorGroup(nil), cal...)
	for k := range s.shiftColorManual {
		if seen[k] {
			continue
		}
		out = append(out, shiftColorGroup{Key: k, ManualOnly: true})
	}
	sort.Slice(out, func(i, j int) bool {
		return shiftColorKeySort(out[i].Key, out[j].Key)
	})
	return out
}

func (s *settingsTab) shiftColorGroupsFromCalendar() []shiftColorGroup {
	if s.shiftsSource == nil {
		return nil
	}
	return collectShiftColorGroups(s.shiftsSource())
}

// collectShiftColorGroups builds color groups from calendar shift codes.
func collectShiftColorGroups(shifts []calendarShift) []shiftColorGroup {
	byKey := map[string]map[string]struct{}{}
	for _, sh := range shifts {
		code := strings.TrimSpace(sh.Code)
		if code == "" {
			continue
		}
		key := shiftColorKey(code)
		if key == "" {
			continue
		}
		if byKey[key] == nil {
			byKey[key] = map[string]struct{}{}
		}
		byKey[key][code] = struct{}{}
	}
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return shiftColorKeySort(keys[i], keys[j])
	})
	out := make([]shiftColorGroup, 0, len(keys))
	for _, k := range keys {
		codes := make([]string, 0, len(byKey[k]))
		for c := range byKey[k] {
			codes = append(codes, c)
		}
		sort.Strings(codes)
		out = append(out, shiftColorGroup{Key: k, Codes: codes})
	}
	return out
}

func shiftColorKeySort(a, b string) bool {
	an, aErr := strconv.Atoi(a)
	bn, bErr := strconv.Atoi(b)
	if aErr == nil && bErr == nil {
		return an < bn
	}
	if aErr == nil {
		return true
	}
	if bErr == nil {
		return false
	}
	return a < b
}

func shiftColorGroupLabel(g shiftColorGroup) string {
	names := make([]string, 0, len(g.Codes))
	for _, c := range g.Codes {
		names = append(names, "*"+c)
	}
	joined := strings.Join(names, ", ")
	if _, err := strconv.Atoi(g.Key); err == nil {
		if joined != "" {
			if len(g.Codes) > 1 {
				return fmt.Sprintf("Ryhmä %s (%s)", g.Key, joined)
			}
			return fmt.Sprintf("Ryhmä %s (%s)", g.Key, joined)
		}
		return fmt.Sprintf("Ryhmä %s", g.Key)
	}
	if joined != "" {
		return joined
	}
	if g.ManualOnly {
		return g.Key + " (käsin)"
	}
	return g.Key
}

func (s *settingsTab) openShiftColorPicker(key, label string) {
	if s.window == nil {
		return
	}
	cur := s.colorForKey(key)
	d := dialog.NewColorPicker(
		"Vuoroväri",
		"Valitse väri: "+label,
		func(c color.Color) {
			s.setShiftColor(key, toNRGBA(c))
		},
		s.window,
	)
	d.SetColor(cur)
	d.Show()
}

func toNRGBA(c color.Color) color.NRGBA {
	if n, ok := c.(color.NRGBA); ok {
		return n
	}
	r, g, b, a := c.RGBA()
	return color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}
